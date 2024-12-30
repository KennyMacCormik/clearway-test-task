package authM

import (
	"clearway-test-task/internal/storage"
	"net/http"
	"strings"
)

func WithAuth(next http.Handler, auth storage.Auth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ok := auth.ValidateToken(authHeader[1])
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
