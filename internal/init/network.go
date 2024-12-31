package myinit

import (
	"clearway-test-task/internal/config"
	myhttp "clearway-test-task/internal/net/http"
	"clearway-test-task/internal/storage"
	"context"
	"github.com/google/uuid"
	"log/slog"
	"strconv"
)

func Net(cfg config.Config, ValidateToken func(token string) (string, error),
	GetToken func(ctx context.Context, login string, password string) (string, int64, error),
	db storage.Db, lg *slog.Logger) *myhttp.HttpServer {
	return myhttp.NewHttpServer(cfg.Http.Host,
		strconv.Itoa(cfg.Http.Port),
		cfg.Http.ReadTimeout,
		cfg.Http.WriteTimeout,
		cfg.Http.IdleTimeout,
		ValidateToken,
		GetToken,
		loggerForHandlers(lg),
		db,
	)
}

func loggerForHandlers(lg *slog.Logger) func() *slog.Logger {
	return func() *slog.Logger {
		newLg := lg.With("ID", uuid.New())
		return newLg
	}
}
