package http

import (
	"clearway-test-task/internal/net/http/handlers/assetHandlers"
	"clearway-test-task/internal/net/http/handlers/authHandlers"
	"clearway-test-task/internal/net/http/middleware/authMiddleware"
	"clearway-test-task/internal/storage"
	"context"
	"log/slog"
	"net/http"
	"time"
)

type HttpServer struct {
	svr     *http.Server
	timeout time.Duration
}

func NewHttpServer(host, port string, ReadTimeout, WriteTimeout, IdleTimeout time.Duration, auth storage.Auth, db storage.Db, loggerForHandlers func() *slog.Logger) *HttpServer {
	svr := &HttpServer{
		svr: &http.Server{
			Addr:         host + ":" + port,
			ReadTimeout:  ReadTimeout,
			WriteTimeout: WriteTimeout,
			IdleTimeout:  IdleTimeout,
		},
		timeout: ReadTimeout,
	}

	http.HandleFunc("POST /auth", authHandlers.BasicAuthPost(auth, loggerForHandlers))
	http.Handle("GET /asset/{assetName}",
		authMiddleware.WithAuth(
			assetHandlers.AssetGet(db),
			auth.ValidateToken,
			loggerForHandlers,
		),
	)
	http.Handle("POST /asset/{assetName}",
		authMiddleware.WithAuth(
			assetHandlers.AssetPost(db),
			auth.ValidateToken,
			loggerForHandlers,
		),
	)

	return svr
}

func (s *HttpServer) Close(lg *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	err := s.svr.Shutdown(ctx)
	if err != nil {
		lg.Error("http server shutdown error", "error", err)
		return err
	}
	lg.Debug("http server shutdown successfully")
	return nil
}

func (s *HttpServer) Listen() error {
	return s.svr.ListenAndServe()
}
