package middleware

import (
	"net/http"

	"github.com/yoadey/shiftmanager/internal/domain"
)

// RequireRole returns a middleware that rejects requests from users who do not
// hold at least one of the specified roles. It relies on the JWTAuth middleware
// having already stored the role in the request context.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if userRole == "" {
				http.Error(w, `{"error":"unauthenticated"}`, http.StatusUnauthorized)
				return
			}

			for _, required := range roles {
				if domain.HasRole(userRole, required) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
		})
	}
}

// RequireExactRole returns a middleware that only allows the exact listed roles
// (no hierarchy elevation). Useful when a vorstand should NOT access an admin-only endpoint.
func RequireExactRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if _, ok := allowed[userRole]; !ok {
				http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
