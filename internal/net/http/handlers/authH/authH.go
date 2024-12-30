package authH

import (
	"clearway-test-task/internal/storage"
	"encoding/json"
	"net/http"
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func AuthPost(auth storage.Auth) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var res token

		l, p, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		t, exp, err := auth.GetToken(l, p)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		res.AccessToken = t
		res.ExpiresIn = exp
		res.TokenType = "Bearer"

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
