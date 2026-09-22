package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	members   memberStore
	audit     auditWriter
	jwtSecret string
	jwtExpiry time.Duration
	// loginRedirect is the SPA route to which the callback redirects after a
	// successful (or failed) exchange, passing the result in the URL fragment.
	loginRedirect string
	// bootstrapAdminEmail, when configured (BOOTSTRAP_ADMIN_EMAIL), causes the
	// member with this e-mail to be auto-provisioned and/or promoted to an
	// active administrator on login. This solves the first-admin bootstrap
	// problem without manual database access.
	bootstrapAdminEmail string
}

// memberStore is the subset of member persistence the auth flow needs: looking
// up members and auto-registering / linking OIDC identities on first login.
type memberStore interface {
	GetByEmail(r *http.Request, email string) (*domain.Member, error)
	GetByOIDCSubject(r *http.Request, provider, subject string) (*domain.Member, error)
	Create(r *http.Request, m *domain.Member) error
	Update(r *http.Request, m *domain.Member) error
	LinkOIDC(r *http.Request, link *domain.OIDCLink) error
}

type auditWriter interface {
	WriteAudit(r *http.Request, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error
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
	// Clear the short-lived state cookie regardless of outcome.
	defer http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	stateCookie, err := r.Cookie("oidc_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		h.redirectError(w, r, "invalid_state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectError(w, r, "missing_code")
		return
	}

	tokens, err := h.oidc.Exchange(r.Context(), code)
	if err != nil {
		h.redirectError(w, r, "exchange_failed")
		return
	}

	claims, err := h.oidc.VerifyIDToken(r.Context(), tokens.IDToken)
	if err != nil {
		h.redirectError(w, r, "token_invalid")
		return
	}

	isBootstrapAdmin := h.bootstrapAdminEmail != "" &&
		strings.EqualFold(strings.TrimSpace(claims.Email), strings.TrimSpace(h.bootstrapAdminEmail))

	// Resolve the member: match by OIDC subject (oidc_links), then fall back to
	// e-mail. The members list is maintained independently of the IdP (ML-004),
	// so a known member's first login is matched by their e-mail address.
	member, _ := h.members.GetByOIDCSubject(r, claims.Provider, claims.Subject)
	matchedBySubject := member != nil
	if member == nil {
		member, _ = h.members.GetByEmail(r, claims.Email)
	}

	justRegistered := false
	switch {
	case member == nil:
		// Unknown identity: auto-register a new member. Per policy they remain
		// INACTIVE (pending admin activation) — except the configured bootstrap
		// admin, who is created as an active administrator.
		member, err = h.provisionMember(r, claims, isBootstrapAdmin)
		if err != nil {
			h.redirectError(w, r, "server_error")
			return
		}
		justRegistered = true

	default:
		// Known member: ensure the OIDC subject is linked for future logins and
		// apply bootstrap-admin promotion if applicable.
		changed := false
		if !matchedBySubject {
			// Matched by e-mail; no link for this subject exists yet (otherwise
			// GetByOIDCSubject would have found it), so linking is safe.
			h.linkOIDC(r, member.ID, claims)
			if member.OIDCSubject == nil {
				member.OIDCSubject = &claims.Subject
				changed = true
			}
		}
		if isBootstrapAdmin && (member.Role != domain.RoleAdmin || !member.IsActive) {
			before := *member
			member.Role = domain.RoleAdmin
			member.IsActive = true
			changed = true
			_ = h.audit.WriteAudit(r, &member.ID, "member.bootstrap_admin", "member", member.ID.String(), before, *member)
		}
		if changed {
			if err := h.members.Update(r, member); err != nil {
				h.redirectError(w, r, "server_error")
				return
			}
		}
	}

	if !member.IsActive {
		// Newly self-registered accounts wait for admin activation; existing
		// inactive accounts have been deactivated.
		if justRegistered {
			h.redirectError(w, r, "registered")
		} else {
			h.redirectError(w, r, "inactive")
		}
		return
	}

	jwtToken, err := h.issueJWT(member)
	if err != nil {
		h.redirectError(w, r, "server_error")
		return
	}

	h.redirectSuccess(w, r, jwtToken)
}

// provisionMember creates a member record for a first-time OIDC user. Regular
// users are created inactive (pending admin activation); the configured
// bootstrap admin is created as an active administrator.
func (h *AuthHandler) provisionMember(r *http.Request, claims *port.OIDCClaims, bootstrapAdmin bool) (*domain.Member, error) {
	first, last := splitName(claims.Name, claims.Email)
	member := &domain.Member{
		ID:          uuid.New(),
		FirstName:   first,
		LastName:    last,
		Email:       claims.Email,
		JoinedAt:    time.Now().UTC(),
		IsActive:    bootstrapAdmin,
		Role:        domain.RoleMitglied,
		OIDCSubject: &claims.Subject,
	}
	if bootstrapAdmin {
		member.Role = domain.RoleAdmin
	}
	if err := h.members.Create(r, member); err != nil {
		return nil, err
	}
	// Best-effort link so subsequent logins match by subject; e-mail fallback
	// still works if this fails.
	h.linkOIDC(r, member.ID, claims)

	action := "member.autoregister"
	if bootstrapAdmin {
		action = "member.bootstrap_admin"
	}
	_ = h.audit.WriteAudit(r, &member.ID, action, "member", member.ID.String(), nil, *member)
	return member, nil
}

// linkOIDC records an oidc_links row, ignoring errors (the e-mail fallback
// keeps login working even if the link cannot be written).
func (h *AuthHandler) linkOIDC(r *http.Request, memberID uuid.UUID, claims *port.OIDCClaims) {
	_ = h.members.LinkOIDC(r, &domain.OIDCLink{
		ID:       uuid.New(),
		MemberID: memberID,
		Provider: claims.Provider,
		Subject:  claims.Subject,
		LinkedAt: time.Now().UTC(),
	})
}

// splitName derives a first/last name from the OIDC "name" claim, falling back
// to the local part of the e-mail address when no name is provided.
func splitName(name, email string) (first, last string) {
	name = strings.TrimSpace(name)
	if name == "" {
		local := email
		if at := strings.IndexByte(email, '@'); at > 0 {
			local = email[:at]
		}
		return local, ""
	}
	parts := strings.Fields(name)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

// redirectToSPA sends the browser back to the single-page app's callback
// route, passing the result in the URL fragment. The fragment is never sent
// to the server, so the token stays out of access logs and the Referer header;
// the SPA reads it client-side and then strips it from the URL.
func (h *AuthHandler) redirectToSPA(w http.ResponseWriter, r *http.Request, fragment string) {
	target := h.loginRedirect
	if target == "" {
		target = "/auth/callback"
	}
	http.Redirect(w, r, target+"#"+fragment, http.StatusFound)
}

// redirectError redirects to the SPA with a machine-readable error code.
func (h *AuthHandler) redirectError(w http.ResponseWriter, r *http.Request, code string) {
	h.redirectToSPA(w, r, "error="+url.QueryEscape(code))
}

// redirectSuccess redirects to the SPA with the freshly issued JWT.
func (h *AuthHandler) redirectSuccess(w http.ResponseWriter, r *http.Request, token string) {
	h.redirectToSPA(w, r, fmt.Sprintf("token=%s&expiresIn=%d", url.QueryEscape(token), int(h.jwtExpiry.Seconds())))
}

// Me returns the currently authenticated user's profile.
// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	email := middleware.GetUserEmail(r.Context())
	firstName := middleware.GetUserFirstName(r.Context())
	lastName := middleware.GetUserLastName(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":        userID,
		"role":      role,
		"email":     email,
		"firstName": firstName,
		"lastName":  lastName,
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
		Role:      member.Role,
		Email:     member.Email,
		FirstName: member.FirstName,
		LastName:  member.LastName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
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
