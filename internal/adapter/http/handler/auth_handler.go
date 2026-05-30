package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// AuthHandler handles OIDC-based authentication flows.
type AuthHandler struct {
	oidc      port.OIDCService
	members   memberGetter
	audit     auditWriter
	jwtSecret string
	jwtExpiry time.Duration
}

type memberGetter interface {
	GetByEmail(r *http.Request, email string) (*domain.Member, error)
	GetByOIDCSubject(r *http.Request, provider, subject string) (*domain.Member, error)
}

type auditWriter interface {
	WriteAudit(r *http.Request, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	oidc port.OIDCService,
	members memberRepoAdapter,
	audit auditRepoAdapter,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthHandler {
	return &AuthHandler{
		oidc:      oidc,
		members:   members,
		audit:     audit,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Login initiates the OIDC authorization code flow.
// GET /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()
	// In production, store state in a session cookie to prevent CSRF.
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	url := h.oidc.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

// Callback handles the OIDC provider redirect with an authorization code.
// GET /api/v1/auth/callback
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oidc_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, `{"error":"invalid state parameter"}`, http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"missing code parameter"}`, http.StatusBadRequest)
		return
	}

	tokens, err := h.oidc.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"token exchange failed"}`, http.StatusUnauthorized)
		return
	}

	claims, err := h.oidc.VerifyIDToken(r.Context(), tokens.IDToken)
	if err != nil {
		http.Error(w, `{"error":"id token verification failed"}`, http.StatusUnauthorized)
		return
	}

	// Look up member by OIDC subject, then fall back to email.
	var member *domain.Member
	member, err = h.members.GetByOIDCSubject(r, claims.Provider, claims.Subject)
	if err != nil || member == nil {
		member, err = h.members.GetByEmail(r, claims.Email)
		if err != nil || member == nil {
			// Auto-provision a basic member record for first-time OIDC users.
			// A real implementation might redirect to a registration page instead.
			http.Error(w, `{"error":"no member account found"}`, http.StatusForbidden)
			return
		}
	}

	if !member.IsActive {
		http.Error(w, `{"error":"account is inactive"}`, http.StatusForbidden)
		return
	}

	jwtToken, err := h.issueJWT(member)
	if err != nil {
		http.Error(w, `{"error":"token issuance failed"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":     jwtToken,
		"expiresIn": int(h.jwtExpiry.Seconds()),
		"member": map[string]interface{}{
			"id":    member.ID,
			"email": member.Email,
			"role":  member.Role,
		},
	})
}

// Me returns the currently authenticated user's profile.
// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	email := middleware.GetUserEmail(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":    userID,
		"role":  role,
		"email": email,
	})
}

// Logout clears any session state.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	aid := userID
	_ = h.audit.WriteAudit(r, &aid, domain.AuditActionLogout, domain.AuditEntityAuth, userID.String(), nil, nil)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) issueJWT(member *domain.Member) (string, error) {
	now := time.Now().UTC()
	claims := middleware.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   member.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(h.jwtExpiry)),
			Issuer:    "shiftmanager",
		},
		Role:  member.Role,
		Email: member.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// memberRepoAdapter and auditRepoAdapter provide thin adapter shims for the
// handler's dependencies so that the handler only depends on its own interface.
type memberRepoAdapter struct {
	repo interface {
		GetByEmail(ctx interface{}, email string) (*domain.Member, error)
		GetByOIDCSubject(ctx interface{}, provider, subject string) (*domain.Member, error)
	}
}

func (a memberRepoAdapter) GetByEmail(r *http.Request, email string) (*domain.Member, error) {
	return a.repo.GetByEmail(r.Context(), email)
}

func (a memberRepoAdapter) GetByOIDCSubject(r *http.Request, provider, subject string) (*domain.Member, error) {
	return a.repo.GetByOIDCSubject(r.Context(), provider, subject)
}

type auditRepoAdapter struct {
	repo interface {
		Insert(ctx interface{}, e *domain.AuditEntry) error
	}
}

func (a auditRepoAdapter) WriteAudit(r *http.Request, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return a.repo.Insert(r.Context(), entry)
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON decodes the request body into v.
func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
