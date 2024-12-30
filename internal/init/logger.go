package myinit

import (
	"clearway-test-task/internal/config"
	"clearway-test-task/pkg"
	"log/slog"
	"os"
)

var logLevelMap = map[string]slog.Level{
	"debug": -4,
	"info":  0,
	"warn":  4,
	"error": 8,
}

func Logger(conf config.Config) *slog.Logger {
	if validateLoggingConf(conf) {
		lg := loggerWithConf(conf)
		return lg
	} else {
		lg := pkg.DefaultLogger()
		lg.Info("config validation failed, running logger with default values level=info format=text")
		return lg
	}
}

func validateLoggingConf(conf config.Config) bool {
	if conf.Log.Level != "debug" &&
		conf.Log.Level != "info" &&
		conf.Log.Level != "warn" &&
		conf.Log.Level != "error" {
		return false
	}
	if conf.Log.Format != "text" &&
		conf.Log.Format != "json" {
		return false
	}
	return true
}

func loggerWithConf(conf config.Config) *slog.Logger {
	var logLevel = new(slog.LevelVar)
	logLevel.Set(logLevelMap[conf.Log.Level])

	if conf.Log.Format == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}
