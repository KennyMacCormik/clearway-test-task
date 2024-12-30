package authMiddleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const userKey string = "user"
const loggerKey string = "logger"

func WithAuth(next http.Handler, validateToken func(token string) (string, error), loggerForHandlers func() *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lg := loggerForHandlers()
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			lg.Error("authorization error", "error", errors.New("authorization header is empty").Error())
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
			lg.Error("authorization error", "error", errors.New("authorization header is not bearer").Error())
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		token := authHeaderParts[1]

		uid, err := validateToken(token)
		if err != nil {
			lg.Error("authorization error", "error", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, uid)
		ctx = context.WithValue(ctx, loggerKey, lg)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetLoginFromContext retrieves the user ID from the request context
func GetLoginFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userKey).(string)
	return userID, ok
}

func GetLoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	lg, ok := ctx.Value(loggerKey).(*slog.Logger)
	return lg, ok
}
