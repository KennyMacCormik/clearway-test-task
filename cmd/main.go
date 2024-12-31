package main

import (
	"clearway-test-task/internal/config"
	myinit "clearway-test-task/internal/init"
	"os"
	"os/signal"
	"syscall"
)

const errExit = 1

func main() {
	cfg, err := config.New()
	if err != nil {
		lg := myinit.Logger(cfg)
		lg.Error("config init error", "error", err.Error())
		os.Exit(errExit)
	}

	lg := myinit.Logger(cfg)
	lg.Debug("logger init success")

	db, auth, err := myinit.Storage(cfg, lg)
	if err != nil {
		lg.Error("db init error", "error", err.Error())
		os.Exit(errExit)
	}
	lg.Debug("db init success")
	defer func() { _ = db.Close(lg) }()
	defer func() { _ = auth.Close() }()

	svr := myinit.Net(cfg, auth.ValidateToken, auth.GetToken, db, lg)
	lg.Debug("http init success")
	defer func() { _ = svr.Close(lg) }()

	go func() { _ = svr.Listen() }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
