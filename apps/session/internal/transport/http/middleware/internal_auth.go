package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const (
	InternalServiceHeader = "X-Internal-Service"
	InternalTokenHeader   = "X-Internal-Token"
)

func InternalAuth(expectedService string, token string) func(http.Handler) http.Handler {
	expectedService = strings.ToLower(strings.TrimSpace(expectedService))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			service := strings.ToLower(strings.TrimSpace(r.Header.Get(InternalServiceHeader)))
			gotToken := strings.TrimSpace(r.Header.Get(InternalTokenHeader))

			if service != expectedService {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if subtle.ConstantTimeCompare([]byte(gotToken), []byte(token)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
