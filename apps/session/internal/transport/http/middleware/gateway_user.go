package middleware

import (
	"context"
	"net/http"
	"strings"
)

const (
	UserIDHeader   = "X-User-ID"
	UserRoleHeader = "X-User-Role"
)

type GatewayUser struct {
	ID   string
	Role string
}

type gatewayUserKey struct{}

func GatewayUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.Header.Get(UserIDHeader))
		userRole := strings.TrimSpace(r.Header.Get(UserRoleHeader))

		if userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		user := GatewayUser{
			ID:   userID,
			Role: userRole,
		}

		ctx := context.WithValue(r.Context(), gatewayUserKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GatewayUserFromContext(ctx context.Context) (GatewayUser, bool) {
	user, ok := ctx.Value(gatewayUserKey{}).(GatewayUser)
	return user, ok
}
