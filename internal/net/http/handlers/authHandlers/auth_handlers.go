package authHandlers

import (
	myerrors "clearway-test-task/internal/errors"
	"clearway-test-task/internal/net/http/middleware"
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
		const fn string = "BasicAuthPost"
		lg := middleware.SetupLoggerFromFunc(fn, r, loggerForHandlers)
		var res token
		l, p, ok := r.BasicAuth()
		if !ok {
			lg.Error("basic auth required", "error", "basic auth required")
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		t, exp, err := auth.GetToken(r.Context(), l, p)
		if err != nil {
			var userErr myerrors.ErrUserNotFound
			if errors.As(err, &userErr) {
				lg.Error("auth error",
					"error", err,
					"Login", userErr.Login,
				)
				http.Error(w, "", http.StatusNotFound)
				return
			}
			lg.Error("failed to get auth token", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		res.AccessToken = pkg.Base64Encode(t)
		res.ExpiresIn = exp
		res.TokenType = "Bearer"

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(res); err != nil {
			lg.Error("error writing response", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		lg.Info("success", "Login", l)
	}
}
