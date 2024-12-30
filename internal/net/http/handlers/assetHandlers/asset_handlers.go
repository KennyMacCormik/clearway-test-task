package assetHandlers

import (
	"clearway-test-task/internal/net/http/middleware/authMiddleware"
	"clearway-test-task/internal/storage"
	"clearway-test-task/pkg"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
)

const assetNameValidationTag = "alphanum"

func AssetPost(db storage.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		lg, ok := authMiddleware.GetLoggerFromContext(r.Context())
		if !ok {
			lg = pkg.DefaultLogger()
		}
		assetName, login, err := getAssetNameAndLogin(r)
		if err != nil {
			lg.Error("assetPost error getting asset name and login", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			lg.Error("assetPost error reading body", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		if err = db.SetDataByAssetName(r.Context(), assetName, login, r.Header.Get("Content-Type"), b); err != nil {
			lg.Error("assetPost rror setting data", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if _, err = w.Write([]byte("{\"status\":\"ok\"}")); err != nil {
			lg.Error("assetPost error writing response", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	})
}

func AssetGet(db storage.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lg, ok := authMiddleware.GetLoggerFromContext(r.Context())
		if !ok {
			lg = pkg.DefaultLogger()
		}
		assetName, login, err := getAssetNameAndLogin(r)
		if err != nil {
			lg.Error("assetGet error getting asset name and login", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		d, ct, err := db.GetDataByAssetName(r.Context(), assetName, login)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				lg.Error("assetGet asset name does not exist", "error", err)
				http.Error(w, "", http.StatusNotFound)
				return
			}
			lg.Error("assetGet error getting asset", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		switch ct {
		case "application/text":
			w.Header().Set("Content-Type", "application/text; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		if _, err = w.Write(d); err != nil {
			lg.Error("assetGet error writing response", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	})
}

func getAssetNameAndLogin(r *http.Request) (string, string, error) {
	assetName := r.PathValue("assetName")
	if err := validateAssetName(assetName); err != nil {
		return "", "", fmt.Errorf("invalid asset name: %w", err)
	}
	login, ok := authMiddleware.GetLoginFromContext(r.Context())
	if !ok {
		return "", "", pkg.NewNotFoundError("login not found")
	}
	return assetName, login, nil
}

func validateAssetName(assetName string) error {
	val := validator.New(validator.WithRequiredStructEnabled())
	if err := val.Var(assetName, assetNameValidationTag); err != nil {
		return fmt.Errorf("assetName validation failed: %w", errors.New("expected match regex '/(\\w+)/g'"))
	}
	return nil
}
