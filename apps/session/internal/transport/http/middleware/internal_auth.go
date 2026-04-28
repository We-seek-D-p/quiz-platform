package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
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
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.Token)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
