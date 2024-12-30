package authHandlers

import (
	"clearway-test-task/internal/storage"
	"clearway-test-task/pkg"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func BasicAuthPost(auth storage.Auth, loggerForHandlers func() *slog.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lg := loggerForHandlers()
		var res token
		l, p, ok := r.BasicAuth()
		if !ok {
			lg.Error("BasicAuthPost basic auth required", "error", "basic auth required")
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		t, exp, err := auth.GetToken(r.Context(), l, p)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				lg.Error("BasicAuthPost auth error", "error", err)
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
			lg.Error("BasicAuthPost failed to get auth token", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		res.AccessToken = t
		res.ExpiresIn = exp
		res.TokenType = "Bearer"

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(res); err != nil {
			lg.Error("BasicAuthPost error writing response", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
