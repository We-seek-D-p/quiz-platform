package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

const (
	InternalServiceHeader = "X-Internal-Service"
	InternalTokenHeader   = "X-Internal-Token"
)

func InternalAuth(cfg config.Internal) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			service := strings.TrimSpace(r.Header.Get(InternalServiceHeader))
			token := strings.TrimSpace(r.Header.Get(InternalTokenHeader))

			if !cfg.Allows(service) {
				response.Error(w, http.StatusForbidden, "forbidden", "forbidden")
				return
			}

			if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.Token)) != 1 {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
