package authHandlers

import (
	myerrors "clearway-test-task/internal/errors"
	"clearway-test-task/internal/net/http/middleware/logMiddleware"
	"clearway-test-task/pkg"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type AuthHandler struct {
	getToken func(ctx context.Context, login string, password string) (string, int64, error)
}

func NewAuthHandler(GetToken func(ctx context.Context, login string,
	password string) (string, int64, error)) *AuthHandler {
	return &AuthHandler{
		getToken: GetToken,
	}
}

func RegAuthHandlers(post http.Handler) {
	http.Handle("POST /auth", post)
}

func (a *AuthHandler) AuthPost() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const fn string = "BasicAuthPost"
		lg := logMiddleware.SetupLoggerFromContext(fn, r)
		var res token
		l, p, ok := r.BasicAuth()
		if !ok {
			lg.Error("basic auth required", "error", "basic auth required")
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		t, exp, err := a.getToken(r.Context(), l, p)
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
	})
}
