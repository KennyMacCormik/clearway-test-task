package middleware

import (
	"clearway-test-task/pkg"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const validateTokenTag string = "jwt"
const userKey string = "user"
const loggerKey string = "logger"

func WithMiddleware(next http.Handler, validateToken func(token string) (string, error), loggerForHandlers func() *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const fn string = "WithAuth"
		lg := WithHttpData(r, loggerForHandlers())
		newLg := lg.With("func", fn)

		token, err := getTokenFromRequest(r)
		if err != nil {
			newLg.Error("token error", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		uid, err := validateToken(token)
		if err != nil {
			newLg.Error("authorization error", "error", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, uid)
		ctx = context.WithValue(ctx, loggerKey, lg)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getTokenFromRequest retrieves the auth header, reads token from it and validates it
func getTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", errors.New("authorization header is not bearer")
	}

	token := pkg.Base64Decode(authHeaderParts[1])
	if token == "" {
		return "", errors.New("error decoding token from base64")
	}

	err := pkg.ValidateWithTag(token, validateTokenTag)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetLoginFromContext retrieves the user ID from the request context
func GetLoginFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userKey).(string)
	return userID, ok
}

// GetLoggerFromContext retrieves the *slog.Logger from the request context
func GetLoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	lg, ok := ctx.Value(loggerKey).(*slog.Logger)
	return lg, ok
}
