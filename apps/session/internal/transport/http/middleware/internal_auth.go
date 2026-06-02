package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

const (
	InternalServiceHeader = "X-Internal-Service"
	InternalTokenHeader   = "X-Internal-Token"
)

func InternalAuth(cfg config.Internal, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			service := strings.TrimSpace(r.Header.Get(InternalServiceHeader))
			token := strings.TrimSpace(r.Header.Get(InternalTokenHeader))

			if !cfg.Allows(service) {
				log.WarnContext(
					r.Context(),
					"unauthorized service perimeter intrusion attempt blocked",
					"request_id", RequestIDFromContext(r.Context()),
					"attempted_service", service,
					"path", r.URL.Path,
				)
				response.Error(w, http.StatusForbidden, "forbidden", "access denied for this service in internal perimeter")
				return
			}

			if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.Token)) != 1 {
				log.WarnContext(
					r.Context(),
					"service auth rejected due to token mismatch",
					"request_id", RequestIDFromContext(r.Context()),
					"authorized_service", service,
					"path", r.URL.Path,
				)
				response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid internal token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
