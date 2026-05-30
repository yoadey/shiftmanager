package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	// ContextKeyUserID is the context key for the authenticated user's ID.
	ContextKeyUserID contextKey = "userID"
	// ContextKeyUserRole is the context key for the authenticated user's role.
	ContextKeyUserRole contextKey = "userRole"
	// ContextKeyUserEmail is the context key for the authenticated user's email.
	ContextKeyUserEmail contextKey = "userEmail"
)

// Claims represents the JWT payload used in ShiftManager tokens.
type Claims struct {
	jwt.RegisteredClaims
	Role  string `json:"role"`
	Email string `json:"email"`
}

// JWTAuth returns a middleware that validates Bearer JWTs.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			sub := claims.Subject
			userID, err := uuid.Parse(sub)
			if err != nil {
				http.Error(w, `{"error":"invalid token subject"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, ContextKeyUserRole, claims.Role)
			ctx = context.WithValue(ctx, ContextKeyUserEmail, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the authenticated user's UUID from context.
// Returns uuid.Nil if not present.
func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ContextKeyUserID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetUserRole extracts the authenticated user's role from context.
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(ContextKeyUserRole).(string); ok {
		return role
	}
	return ""
}

// GetUserEmail extracts the authenticated user's email from context.
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(ContextKeyUserEmail).(string); ok {
		return email
	}
	return ""
}
