package httptransport

import (
	"log/slog"
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.Recoverer(log))
	r.Use(middleware.GatewayUserMiddleware)

	r.Get("/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Warn(
			"route not found",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
		)

		http.Error(w, "not found", http.StatusNotFound)
	})

	return r
}
