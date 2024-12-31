package assetHandlers

import (
	myerrors "clearway-test-task/internal/errors"
	"clearway-test-task/internal/net/http/middleware"
	"clearway-test-task/internal/storage"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
)

const assetNameValidationTag = "alphanum"

type AssetHandler struct {
	db storage.Db
}

func NewAssetHandler(db storage.Db) *AssetHandler {
	return &AssetHandler{
		db: db,
	}
}

func (a *AssetHandler) AssetDelete() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		const fn string = "AssetDelete"
		lg := middleware.SetupLoggerFromContext(fn, r)

		assetName, login, err := getAssetNameAndLogin(r)
		if err != nil {
			lg.Error("error getting asset name and login", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		if err = a.db.DeleteDataByAssetName(r.Context(), assetName, login); err != nil {
			lg.Error("error deleting data", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if _, err = w.Write([]byte("{\"status\":\"ok\"}")); err != nil {
			lg.Error("error writing response", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		lg.Info("success", "Login", login, "AssetName", assetName)
	})
}

func (a *AssetHandler) AssetPost() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		const fn string = "AssetPost"
		lg := middleware.SetupLoggerFromContext(fn, r)

		assetName, login, err := getAssetNameAndLogin(r)
		if err != nil {
			lg.Error("error getting asset name and login", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			lg.Error("error reading body", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		if err = a.db.SetDataByAssetName(r.Context(), assetName, login, r.Header.Get("Content-Type"), b); err != nil {
			lg.Error("error setting data", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if _, err = w.Write([]byte("{\"status\":\"ok\"}")); err != nil {
			lg.Error("error writing response", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		lg.Info("success", "Login", login, "AssetName", assetName)
	})
}

func (a *AssetHandler) AssetGet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const fn string = "AssetGet"
		lg := middleware.SetupLoggerFromContext(fn, r)

		assetName, login, err := getAssetNameAndLogin(r)
		if err != nil {
			lg.Error("error getting asset name and login", "error", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		d, ct, err := a.db.GetDataByAssetName(r.Context(), assetName, login)
		if err != nil {
			var assetErr myerrors.ErrAssetNotFound
			if errors.As(err, &assetErr) {
				lg.Error("asset name does not exist",
					"error", err,
					"AssetId", assetErr.AssetID,
					"Login", assetErr.Login,
				)
				http.Error(w, "", http.StatusNotFound)
				return
			}
			lg.Error("error getting asset", "error", err)
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
			lg.Error("error writing response", "error", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		lg.Info("success", "Login", login, "AssetName", assetName)
	})
}

func getAssetNameAndLogin(r *http.Request) (string, string, error) {
	assetName := r.PathValue("assetName")
	if err := validateAssetName(assetName); err != nil {
		return "", "", fmt.Errorf("invalid asset name: %w", err)
	}
	login, ok := middleware.GetLoginFromContext(r.Context())
	if !ok {
		return "", "", myerrors.NewNotFoundError("login not found")
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
