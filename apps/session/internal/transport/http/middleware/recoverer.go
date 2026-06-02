package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := recover()
				if err == nil {
					return
				}

				log.ErrorContext(
					r.Context(),
					"panic recovered",
					"request_id", RequestIDFromContext(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"error", err,
					"stack", string(debug.Stack()),
				)

				response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
			}()

			next.ServeHTTP(w, r)
		})
	}
}
