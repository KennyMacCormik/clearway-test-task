package http

import (
	"clearway-test-task/internal/net/http/handlers/assetH"
	"clearway-test-task/internal/net/http/handlers/authH"
	"clearway-test-task/internal/net/http/middleware/authM"
	"clearway-test-task/internal/storage"
	"context"
	"net/http"
	"time"
)

const httpDefaultTimeout = 1 * time.Second

type HttpServer struct {
	svr     *http.Server
	timeout time.Duration
}

func NewHttpServer(host, port string, auth storage.Auth, db storage.Db) *HttpServer {
	svr := &HttpServer{
		svr: &http.Server{
			Addr:         host + ":" + port,
			ReadTimeout:  httpDefaultTimeout,
			WriteTimeout: httpDefaultTimeout,
			IdleTimeout:  httpDefaultTimeout,
		},
		timeout: httpDefaultTimeout,
	}

	http.HandleFunc("POST /auth", authH.AuthPost(auth))
	http.Handle("GET /asset/{assetName}", authM.WithAuth(assetH.AssetGet(db), auth))
	http.Handle("POST /asset/{assetName}", authM.WithAuth(assetH.AssetPost(db), auth))

	return svr
}

func (s *HttpServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.svr.Shutdown(ctx)
}

func (s *HttpServer) Listen() error {
	return s.svr.ListenAndServe()
}
