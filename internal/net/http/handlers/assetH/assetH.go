package assetH

import (
	"clearway-test-task/internal/helpers"
	"clearway-test-task/internal/storage"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"strings"
)

const hmacSecret = "secret"

func AssetPost(db storage.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		assetName := r.PathValue("assetName")
		if assetName == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = db.SetDataByAssetName(
			assetName,
			getLoginFromToken(strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]),
			r.Header.Get("Content-Type"),
			b); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if _, err = w.Write([]byte("{\"status\":\"ok\"}")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
func AssetGet(db storage.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assetName := r.PathValue("assetName")
		if assetName == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		d, ct, err := db.GetDataByAssetName(
			assetName,
			getLoginFromToken(strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]),
		)
		if err != nil {
			if errors.As(err, &helpers.NotFound{}) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		switch ct {
		case "application/text":
			w.Header().Set("Content-Type", "application/text; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		if _, err := w.Write(d); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

// getLoginFromToken ignores all possible errors.
// It is possible because we use authentication middleware that thoroughly checks token for errors.
// At this point, we're sure this token is safe and valid.
func getLoginFromToken(tkn string) string {
	t, _ := jwt.Parse(helpers.Base64Decode(tkn), func(tkn *jwt.Token) (interface{}, error) {
		return helpers.ConvertStrToBytes(hmacSecret), nil
	})
	c := t.Claims.(jwt.MapClaims)
	return c["login"].(string)
}
