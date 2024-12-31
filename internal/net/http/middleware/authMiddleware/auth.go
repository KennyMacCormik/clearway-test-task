package authMiddleware

import (
	"clearway-test-task/internal/net/http/middleware/logMiddleware"
	"clearway-test-task/pkg"
	"clearway-test-task/pkg/validator"
	"context"
	"errors"
	"net/http"
	"strings"
)

const validateTokenTag string = "jwt"
const UserKey string = "user"

type AuthMiddleware struct {
	validateToken func(token string) (string, error)
}

func NewAuthMiddleware(validateToken func(token string) (string, error)) *AuthMiddleware {
	return &AuthMiddleware{validateToken: validateToken}
}

func (a *AuthMiddleware) WithBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const fn string = "WithAuth"
		lg := logMiddleware.SetupLoggerFromContext(fn, r)

		token, err := getTokenFromRequest(r)
		if err != nil {
			lg.Error("token error", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		uid, err := a.validateToken(token)
		if err != nil {
			lg.Error("authorization error", "error", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, uid)

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

	err := validator.ValInstance.ValidateWithTag(token, validateTokenTag)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetLoginFromContext retrieves the user ID from the request context
func GetLoginFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserKey).(string)
	return userID, ok
}
