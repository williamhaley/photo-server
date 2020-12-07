package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func (s *Server) LogIn(rw http.ResponseWriter, r *http.Request) {
	loginData := struct {
		AccessCode string `json:"accessCode"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		log.WithError(err).Error("error decoding form")
		rw.WriteHeader(http.StatusUnauthorized)
		result := map[string]string{
			"error": "access denied",
		}
		if err := json.NewEncoder(rw).Encode(result); err != nil {
			log.WithError(err).Error("error writing response")
		}
	}

	if loginData.AccessCode != s.accessCode {
		rw.WriteHeader(http.StatusUnauthorized)
		result := map[string]string{
			"error": "access denied",
		}
		if err := json.NewEncoder(rw).Encode(result); err != nil {
			log.WithError(err).Error("error writing response")
		}
		return
	}

	// create the token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["foo"] = "bar"
	claims["exp"] = time.Now().Add(time.Hour * 24 * 60).Unix() // 2 months
	token.Claims = claims

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		log.WithError(err).Error("error signing token")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]string{
		"token": tokenString,
	}

	if err := json.NewEncoder(rw).Encode(result); err != nil {
		log.WithError(err).Error("error writing response")
	}
}

func (s *Server) Profile(rw http.ResponseWriter, r *http.Request) {
	result := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(rw).Encode(result); err != nil {
		log.WithError(err).Error("error writing response")
	}
}

func (s *Server) BucketCounts(rw http.ResponseWriter, r *http.Request) {
	result, err := s.api.BucketCounts()
	if err != nil {
		log.WithError(err).Error("error retrieving bucket counts")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(rw).Encode(result); err != nil {
		log.WithError(err).Error("error writing response")
	}
}

func (s *Server) PhotosForBucket(rw http.ResponseWriter, r *http.Request) {
	bucketID := chi.URLParam(r, "id")
	after := r.URL.Query().Get("after")
	photosAfter, err := url.QueryUnescape(after)
	if err != nil {
		log.WithError(err).Errorf("error decoding url parameter %q", after)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := s.api.BucketPhotos(bucketID, photosAfter)
	if err != nil {
		log.WithError(err).Errorf("error retrieving photos for bucket %q", bucketID)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(rw).Encode(result); err != nil {
		log.WithError(err).Error("error writing response")
	}
}

// FullImageHandler responds to HTTP requests for full resolution single images.
func (s *Server) FullImageHandler(rw http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		http.Error(rw, "Get 'uuid' not specified in url.", http.StatusBadRequest)
		return
	}

	photo, err := s.db.GetPhoto(uuid)
	if err != nil {
		log.WithError(err).Errorf("could not find photo %q", uuid)
		return
	}

	file, err := os.Open(filepath.Join(s.photosDirectoryRootPath, photo.Path))
	if err != nil {
		log.WithError(err).Error("could not open source image")
		return
	}
	defer file.Close()

	header := make([]byte, 512)
	file.Read(header)
	contentType := http.DetectContentType(header)

	rw.Header().Set("Content-Type", contentType)

	stat, err := file.Stat()
	if err != nil {
		log.WithError(err).Error("error getting file stats")
		return
	}
	size := strconv.FormatInt(stat.Size(), 10)
	rw.Header().Set("Content-Length", size)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	file.Seek(0, 0)
	io.Copy(rw, file)
	return
}

// wildcardHandler responds to all unhandled requests.
func (s *Server) wildcardHandler(fileSystemHandler http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		fileSystemHandler.ServeHTTP(rw, r)
	}
}

// ThumbnailHandler responds to HTTP requests for image thumbnails.
func (s *Server) ThumbnailHandler(rw http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		http.Error(rw, "Get 'uuid' not specified in url.", http.StatusBadRequest)
		return
	}

	photo, err := s.db.GetPhoto(uuid)
	if err != nil {
		log.WithError(err).Errorf("could not find photo %q", uuid)
		return
	}

	overwrite := false
	file, _, err := s.thumbnailManager.Generate(photo, overwrite)
	if err != nil {
		log.WithError(err).Error("could not get thumbnail")
		return
	}
	defer file.Close()

	header := make([]byte, 512)
	file.Read(header)
	contentType := http.DetectContentType(header)

	rw.Header().Set("Content-Type", contentType)

	stat, err := file.Stat()
	if err != nil {
		log.WithError(err).Error("error getting file stats")
		return
	}
	size := strconv.FormatInt(stat.Size(), 10)
	rw.Header().Set("Content-Length", size)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	file.Seek(0, 0)
	io.Copy(rw, file)
	return
}
