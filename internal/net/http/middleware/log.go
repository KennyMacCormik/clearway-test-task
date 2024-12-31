package middleware

import (
	"clearway-test-task/pkg"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

func WithHttpData(r *http.Request, lg *slog.Logger) *slog.Logger {
	lg.Debug("request received",
		"Proto", r.Proto,
		"Header", r.Header,
		"RemoteAddr", r.RemoteAddr,
		"RequestURI", r.RequestURI,
		"ContentLength", r.ContentLength,
		"Method", r.Method,
		"Host", r.Host,
		"UrlPath", r.URL.Path,
	)
	return lg.With("Method", r.Method, "UrlPath", r.URL.Path)
}

func SetupLoggerFromContext(fn string, r *http.Request) *slog.Logger {
	lg, ok := GetLoggerFromContext(r.Context())
	if !ok {
		lg = WithHttpData(r, pkg.DefaultLogger().With("ID", uuid.New()))
	}
	return lg.With("func", fn)
}

func SetupLoggerFromFunc(fn string, r *http.Request, loggerForHandlers func() *slog.Logger) *slog.Logger {
	return WithHttpData(r, loggerForHandlers()).With("func", fn)
}
