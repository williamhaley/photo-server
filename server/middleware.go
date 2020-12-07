package server

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

func TokenMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			tokenString := r.URL.Query().Get("token")
			if tokenString == "" {
				tokenString = r.Header.Get("Authorization")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				log.WithError(err).Errorf("token is invalid")
				http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(rw, r)
		})
	}
}
