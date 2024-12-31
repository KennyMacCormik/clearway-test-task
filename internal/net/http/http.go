package http

import (
	"clearway-test-task/internal/net/http/handlers/assetHandlers"
	"clearway-test-task/internal/net/http/handlers/authHandlers"
	"clearway-test-task/internal/net/http/middleware/authMiddleware"
	"clearway-test-task/internal/net/http/middleware/logMiddleware"
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

func NewHttpServer(host, port string, ReadTimeout, WriteTimeout, IdleTimeout time.Duration,
	ValidateToken func(token string) (string, error),
	GetToken func(ctx context.Context, login string, password string) (string, int64, error),
	loggerForHandlers func() *slog.Logger,
	db storage.Db) *HttpServer {
	svr := &HttpServer{
		svr: &http.Server{
			Addr:         host + ":" + port,
			ReadTimeout:  ReadTimeout,
			WriteTimeout: WriteTimeout,
			IdleTimeout:  IdleTimeout,
		},
		timeout: ReadTimeout,
	}

	assetH := assetHandlers.NewAssetHandler(db)
	authH := authHandlers.NewAuthHandler(GetToken)

	authM := authMiddleware.NewAuthMiddleware(ValidateToken)

	AssetGet := logMiddleware.NewLoggerMiddleware(loggerForHandlers, authM.WithBasicAuth(assetH.AssetGet()))
	AssetPost := logMiddleware.NewLoggerMiddleware(loggerForHandlers, authM.WithBasicAuth(assetH.AssetPost()))
	AssetDelete := logMiddleware.NewLoggerMiddleware(loggerForHandlers, authM.WithBasicAuth(assetH.AssetDelete()))

	AuthPost := logMiddleware.NewLoggerMiddleware(loggerForHandlers, authH.AuthPost())

	assetHandlers.RegAssetHandlers(AssetGet.Handle(), AssetPost.Handle(), AssetDelete.Handle())
	authHandlers.RegAuthHandlers(AuthPost.Handle())

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
