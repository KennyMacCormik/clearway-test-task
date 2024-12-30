package myinit

import (
	"clearway-test-task/internal/config"
	"clearway-test-task/internal/net/http"
	"clearway-test-task/internal/storage"
	"github.com/google/uuid"
	"log/slog"
	"strconv"
)

func Net(cfg config.Config, auth storage.Auth, db storage.Db, lg *slog.Logger) *http.HttpServer {
	return http.NewHttpServer(cfg.Http.Host,
		strconv.Itoa(cfg.Http.Port),
		cfg.Http.ReadTimeout,
		cfg.Http.WriteTimeout,
		cfg.Http.IdleTimeout,
		auth,
		db,
		loggerForHandlers(lg))
}

func loggerForHandlers(lg *slog.Logger) func() *slog.Logger {
	return func() *slog.Logger {
		uid := uuid.New()
		newLg := lg.With("ID", uid)
		return newLg
	}
}
