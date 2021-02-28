package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/photo-server/api"
	"github.com/williamhaley/photo-server/datasource"
	"github.com/williamhaley/photo-server/thumbnail"
)

// Server is the API/backend for serving photos over the web.
type Server struct {
	db                      *datasource.Database
	api                     *api.API
	photosDirectoryRootPath string
	thumbnailManager        *thumbnail.Manager
	httpPort                string
	httpsPort               string
	httpsCertFilePath       string
	httpsCertKeyPath        string
	secret                  string
	accessCode              string
	staticFileSystem        http.FileSystem
}

// New allocates a new instance of the server.
func New(
	db *datasource.Database,
	photosDirectoryRootPath string,
	thumbnailManager *thumbnail.Manager,
	httpPort,
	httpsPort,
	httpsCertFilePath,
	httpsCertKeyPath,
	accessCode string,
	staticFileSystem http.FileSystem,
) *Server {
	return &Server{
		db:                      db,
		api:                     api.New(db),
		photosDirectoryRootPath: photosDirectoryRootPath,
		thumbnailManager:        thumbnailManager,
		httpPort:                httpPort,
		httpsPort:               httpsPort,
		httpsCertFilePath:       httpsCertFilePath,
		httpsCertKeyPath:        httpsCertKeyPath,
		secret:                  accessCode, // TODO WFH not ideal
		accessCode:              accessCode,
		staticFileSystem:        staticFileSystem,
	}
}

// Start initializes the server so it starts listening for connections.
func (s *Server) Start() error {
	appRouter := chi.NewRouter()
	appRouter.Use(middleware.Logger)
	appRouter.Use(middleware.RedirectSlashes)
	appRouter.Use(middleware.Compress(5))
	appRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	tokenMiddleware := TokenMiddleware(s.secret)

	appRouter.Post("/login", s.LogIn)
	appRouter.With(tokenMiddleware).Get("/profile", s.Profile)

	appRouter.Route("/api", func(rg chi.Router) {
		rg.Use(tokenMiddleware)
		rg.Get("/buckets/counts", s.BucketCounts)
		rg.Get("/buckets/{id}", s.PhotosForBucket)
	})
	appRouter.Get("/thumbnail/{uuid}.*", s.ThumbnailHandler)
	appRouter.Get("/full/{uuid}.*", s.FullImageHandler)
	appRouter.Handle("/*", http.FileServer(s.staticFileSystem))

	isUsingHTTPS := s.httpsPort != ""
	if isUsingHTTPS {
		return s.serveHTTPS(appRouter)
	}
	return s.serveHTTP(appRouter)
}

func (s *Server) serveHTTP(appRouter http.Handler) error {
	httpAddress := fmt.Sprintf(":%s", s.httpPort)

	log.Infof("starting http server on %q", httpAddress)

	httpServer := http.Server{
		Addr:    httpAddress,
		Handler: appRouter,
	}
	return httpServer.ListenAndServe()
}

func (s *Server) serveHTTPS(appRouter http.Handler) error {
	httpsAddress := fmt.Sprintf(":%s", s.httpsPort)
	httpAddress := fmt.Sprintf(":%s", s.httpPort)

	log.Infof("starting https server on %q. http traffic on %q will redirect to https", httpsAddress, httpAddress)

	httpsServer := http.Server{
		Addr:    httpsAddress,
		Handler: appRouter,
	}

	go func() {
		httpServer := http.Server{
			Addr:    httpAddress,
			Handler: http.HandlerFunc(s.HTTPtoHTTPSRedirect),
		}

		if err := httpServer.ListenAndServe(); err != nil {
			log.WithError(err).Error("error serving HTTP redirect traffic")
		}
	}()

	return httpsServer.ListenAndServeTLS(s.httpsCertFilePath, s.httpsCertKeyPath)
}
