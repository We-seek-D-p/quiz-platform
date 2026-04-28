package httptransport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/handler"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/middleware"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

func NewRouter(cfg *config.Config, log *slog.Logger, internalSessionHandler *handler.InternalSessionHandler) http.Handler {
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

	r.Route("/internal/v1", func(r chi.Router) {
		r.Use(middleware.InternalAuth(cfg.Internal))

		r.Put("/sessions/{session_id}", internalSessionHandler.InitSession)
		r.Get("/sessions/{session_id}", internalSessionHandler.GetSessionRuntime)
		r.Delete("/sessions/{session_id}", internalSessionHandler.DeleteSessionRuntime)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Warn(
			"route not found",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
		)

		response.Error(w, http.StatusNotFound, "not_found", "route not found")
	})

	return r
}
