package logMiddleware

import (
	"clearway-test-task/pkg"
	"context"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

const LoggerKey string = "logger"

type LogMiddleware struct {
	loggerForHandlers func() *slog.Logger
	next              http.Handler
}

func NewLoggerMiddleware(loggerForHandlers func() *slog.Logger, next http.Handler) *LogMiddleware {
	return &LogMiddleware{loggerForHandlers: loggerForHandlers, next: next}
}

func (a *LogMiddleware) Handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lg := WithHttpData(r, a.loggerForHandlers())
		ctx := context.WithValue(r.Context(), LoggerKey, lg)

		a.next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetLoggerFromContext retrieves the *slog.Logger from the request context
func GetLoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	lg, ok := ctx.Value(LoggerKey).(*slog.Logger)
	return lg, ok
}

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
