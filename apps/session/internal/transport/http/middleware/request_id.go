package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get(RequestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		w.Header().Set(RequestIDHeader, requestID)

		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}

	return requestID
}
