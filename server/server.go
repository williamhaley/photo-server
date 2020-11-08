package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/photo-server/api"
	"github.com/williamhaley/photo-server/datasource"
	"github.com/williamhaley/photo-server/thumbnail"
	"net/http"
	"os"
	"path/filepath"
)

// Server is the API/backend for serving photos over the web.
type Server struct {
	db                      *datasource.Database
	api                     *api.API
	photosDirectoryRootPath string
	thumbnailManager        *thumbnail.Manager
	httpPort                string
}

// New allocates a new instance of the server.
func New(db *datasource.Database, photosDirectoryRootPath string, thumbnailManager *thumbnail.Manager, httpPort string) *Server {
	return &Server{
		db:                      db,
		api:                     api.New(db),
		photosDirectoryRootPath: photosDirectoryRootPath,
		thumbnailManager:        thumbnailManager,
		httpPort:                httpPort,
	}
}

// Start initializes the server so it starts listening for connections.
func (s *Server) Start() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	log.Infof("serving static files from %q", filesDir)
	fs := http.FileServer(http.Dir(filesDir))

	r.Post("/login", s.LogIn)
	r.Route("/api", func(rg chi.Router) {
		rg.Use(TokenMiddleware)
		rg.Get("/buckets/counts", s.BucketCounts)
		rg.Get("/buckets/{id}", s.PhotosForBucket)
	})
	r.Get("/thumbnail/{uuid}.*", s.ThumbnailHandler)
	r.Get("/full/{uuid}.*", s.FullImageHandler)
	r.Get("/*", s.wildcardHandler(fs))

	return http.ListenAndServe(fmt.Sprintf(":%s", s.httpPort), r)
}
