package handler

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
)

// DevAuthHandler provides a token endpoint for test/dev mode only.
// It must never be registered in production (guarded by TestMode flag in router).
type DevAuthHandler struct {
	jwtSecret string
}

// NewDevAuthHandler creates a handler that issues JWTs for a given role without OIDC.
func NewDevAuthHandler(jwtSecret string) *DevAuthHandler {
	return &DevAuthHandler{jwtSecret: jwtSecret}
}

// Token issues a signed JWT for the requested role and member ID.
// GET /dev/token?role=vorstand&id=<uuid>&email=admin@test.local
// All parameters are optional; defaults to role=vorstand, a fixed admin UUID.
func (h *DevAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	if role == "" {
		role = "vorstand"
	}
	idStr := r.URL.Query().Get("id")
	memberID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // default admin seed UUID
	if idStr != "" {
		parsed, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}
		memberID = parsed
	}
	email := r.URL.Query().Get("email")
	if email == "" {
		email = "admin@test.local"
	}

	claims := middleware.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   memberID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		Role:  role,
		Email: email,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(h.jwtSecret))
	if err != nil {
		http.Error(w, `{"error":"token signing failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, http.StatusOK, map[string]string{
		"token": signed,
		"role":  role,
		"id":    memberID.String(),
		"email": email,
	})
}
