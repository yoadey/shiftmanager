package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// --- fakes -----------------------------------------------------------------

type fakeOIDC struct {
	claims *port.OIDCClaims
}

func (f *fakeOIDC) GetAuthURL(state string) string { return "https://idp.example/auth?state=" + state }
func (f *fakeOIDC) Exchange(_ context.Context, _ string) (*port.OIDCTokens, error) {
	return &port.OIDCTokens{IDToken: "id-token"}, nil
}
func (f *fakeOIDC) VerifyIDToken(_ context.Context, _ string) (*port.OIDCClaims, error) {
	return f.claims, nil
}

type fakeMemberStore struct {
	byEmail   map[string]*domain.Member
	bySubject map[string]*domain.Member
	created   []*domain.Member
	updated   []*domain.Member
	links     []*domain.OIDCLink
}

func newFakeStore() *fakeMemberStore {
	return &fakeMemberStore{byEmail: map[string]*domain.Member{}, bySubject: map[string]*domain.Member{}}
}

func (s *fakeMemberStore) GetByEmail(_ *http.Request, email string) (*domain.Member, error) {
	if m, ok := s.byEmail[strings.ToLower(email)]; ok {
		return m, nil
	}
	return nil, domain.ErrMemberNotFound
}
func (s *fakeMemberStore) GetByID(_ *http.Request, id uuid.UUID) (*domain.Member, error) {
	for _, m := range s.byEmail {
		if m.ID == id {
			return m, nil
		}
	}
	for _, m := range s.bySubject {
		if m.ID == id {
			return m, nil
		}
	}
	return nil, domain.ErrMemberNotFound
}
func (s *fakeMemberStore) GetByOIDCSubject(_ *http.Request, _, subject string) (*domain.Member, error) {
	if m, ok := s.bySubject[subject]; ok {
		return m, nil
	}
	return nil, domain.ErrMemberNotFound
}
func (s *fakeMemberStore) Create(_ *http.Request, m *domain.Member) error {
	s.created = append(s.created, m)
	s.byEmail[strings.ToLower(m.Email)] = m
	return nil
}
func (s *fakeMemberStore) Update(_ *http.Request, m *domain.Member) error {
	s.updated = append(s.updated, m)
	return nil
}
func (s *fakeMemberStore) LinkOIDC(_ *http.Request, link *domain.OIDCLink) error {
	s.links = append(s.links, link)
	return nil
}

type fakeAudit struct{ actions []string }

func (a *fakeAudit) WriteAudit(_ *http.Request, _ *uuid.UUID, action, _, _ string, _, _ interface{}) error {
	a.actions = append(a.actions, action)
	return nil
}

// --- helpers ---------------------------------------------------------------

func newHandler(store *fakeMemberStore, claims *port.OIDCClaims, bootstrapEmail string) (*AuthHandler, *fakeAudit) {
	audit := &fakeAudit{}
	return &AuthHandler{
		oidc:                map[string]port.OIDCService{"default": &fakeOIDC{claims: claims}},
		providers:           []OIDCProviderInfo{{Name: "default", Label: "Vereinskonto"}},
		members:             store,
		audit:               audit,
		jwtSecret:           "test-secret",
		jwtExpiry:           time.Hour,
		loginRedirect:       "/auth/callback",
		bootstrapAdminEmail: bootstrapEmail,
	}, audit
}

// callback drives Callback with a matching state cookie and returns the
// fragment after '#' in the redirect Location.
func callback(t *testing.T, h *AuthHandler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?code=c&state=s", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "s" + oidcStateSeparator + "default"})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	_, frag, ok := strings.Cut(loc, "#")
	if !ok {
		t.Fatalf("location %q has no fragment", loc)
	}
	return frag
}

// --- tests -----------------------------------------------------------------

func TestSplitName(t *testing.T) {
	cases := []struct{ name, email, wantFirst, wantLast string }{
		{"Maximilian Müller", "x@y.de", "Maximilian", "Müller"},
		{"Anna Maria Schmidt", "x@y.de", "Anna", "Maria Schmidt"},
		{"Cher", "x@y.de", "Cher", ""},
		{"", "john.doe@club.de", "john.doe", ""},
		{"  ", "solo@club.de", "solo", ""},
	}
	for _, c := range cases {
		first, last := splitName(c.name, c.email)
		if first != c.wantFirst || last != c.wantLast {
			t.Errorf("splitName(%q,%q) = %q/%q, want %q/%q", c.name, c.email, first, last, c.wantFirst, c.wantLast)
		}
	}
}

func TestCallback_UnknownUser_AutoRegisteredInactive(t *testing.T) {
	store := newFakeStore()
	claims := &port.OIDCClaims{Subject: "sub-1", Email: "new@club.de", Name: "Neu Mitglied", Provider: "kc"}
	h, audit := newHandler(store, claims, "")

	frag := callback(t, h)

	if frag != "error=registered" {
		t.Fatalf("want error=registered, got %q", frag)
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 created member, got %d", len(store.created))
	}
	m := store.created[0]
	if m.IsActive {
		t.Error("auto-registered member must be inactive")
	}
	if m.Role != domain.RoleMitglied {
		t.Errorf("role = %q, want mitglied", m.Role)
	}
	if len(store.links) != 1 {
		t.Errorf("expected OIDC link to be written, got %d", len(store.links))
	}
	if len(audit.actions) == 0 || audit.actions[0] != "member.autoregister" {
		t.Errorf("expected autoregister audit, got %v", audit.actions)
	}
}

func TestCallback_BootstrapAdmin_Unknown_CreatedActiveAdmin(t *testing.T) {
	store := newFakeStore()
	claims := &port.OIDCClaims{Subject: "sub-2", Email: "boss@club.de", Name: "Chef", Provider: "kc"}
	h, _ := newHandler(store, claims, "BOSS@club.de") // case-insensitive match

	frag := callback(t, h)

	if !strings.HasPrefix(frag, "token=") {
		t.Fatalf("bootstrap admin should be logged in, got %q", frag)
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 created member, got %d", len(store.created))
	}
	m := store.created[0]
	if !m.IsActive || m.Role != domain.RoleAdmin {
		t.Errorf("bootstrap admin must be active admin, got active=%v role=%q", m.IsActive, m.Role)
	}
}

func TestCallback_ExistingInactive_NotRegisteredMessage(t *testing.T) {
	store := newFakeStore()
	existing := &domain.Member{ID: uuid.New(), Email: "old@club.de", Role: domain.RoleMitglied, IsActive: false}
	store.byEmail["old@club.de"] = existing
	claims := &port.OIDCClaims{Subject: "sub-3", Email: "old@club.de", Provider: "kc"}
	h, _ := newHandler(store, claims, "")

	frag := callback(t, h)

	if frag != "error=inactive" {
		t.Fatalf("want error=inactive, got %q", frag)
	}
	if len(store.created) != 0 {
		t.Error("must not create a member for an existing one")
	}
}

func TestCallback_ExistingActive_MatchedByEmail_LinksSubject(t *testing.T) {
	store := newFakeStore()
	existing := &domain.Member{ID: uuid.New(), Email: "ok@club.de", Role: domain.RoleMitglied, IsActive: true}
	store.byEmail["ok@club.de"] = existing
	claims := &port.OIDCClaims{Subject: "sub-4", Email: "ok@club.de", Provider: "kc"}
	h, _ := newHandler(store, claims, "")

	frag := callback(t, h)

	if !strings.HasPrefix(frag, "token=") {
		t.Fatalf("active member should log in, got %q", frag)
	}
	if len(store.links) != 1 {
		t.Errorf("subject should be linked on first email match, got %d links", len(store.links))
	}
}

func TestCallback_MatchedBySubject_NoDuplicateLink(t *testing.T) {
	store := newFakeStore()
	existing := &domain.Member{ID: uuid.New(), Email: "sub@club.de", Role: domain.RoleVorstand, IsActive: true}
	store.bySubject["sub-5"] = existing
	claims := &port.OIDCClaims{Subject: "sub-5", Email: "sub@club.de", Provider: "kc"}
	h, _ := newHandler(store, claims, "")

	frag := callback(t, h)

	if !strings.HasPrefix(frag, "token=") {
		t.Fatalf("want token, got %q", frag)
	}
	if len(store.links) != 0 {
		t.Errorf("must not re-link an already-linked subject, got %d links", len(store.links))
	}
}

func TestCallback_BootstrapAdmin_PromotesExisting(t *testing.T) {
	store := newFakeStore()
	existing := &domain.Member{ID: uuid.New(), Email: "promote@club.de", Role: domain.RoleMitglied, IsActive: false}
	store.byEmail["promote@club.de"] = existing
	claims := &port.OIDCClaims{Subject: "sub-6", Email: "promote@club.de", Provider: "kc"}
	h, _ := newHandler(store, claims, "promote@club.de")

	frag := callback(t, h)

	if !strings.HasPrefix(frag, "token=") {
		t.Fatalf("promoted admin should log in, got %q", frag)
	}
	if existing.Role != domain.RoleAdmin || !existing.IsActive {
		t.Errorf("existing member should be promoted to active admin, got role=%q active=%v", existing.Role, existing.IsActive)
	}
	if len(store.updated) == 0 {
		t.Error("promotion should persist via Update")
	}
}

func TestRefresh_ActiveMember_IssuesNewToken(t *testing.T) {
	store := newFakeStore()
	member := &domain.Member{ID: uuid.New(), Email: "x@club.de", Role: domain.RoleMitglied, IsActive: true}
	store.byEmail["x@club.de"] = member
	h, _ := newHandler(store, &port.OIDCClaims{}, "")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, member.ID))
	rec := httptest.NewRecorder()
	h.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) || !strings.Contains(rec.Body.String(), `"expiresIn"`) {
		t.Errorf("expected token+expiresIn in response, got %s", rec.Body.String())
	}
}

func TestRefresh_InactiveMember_Unauthorized(t *testing.T) {
	store := newFakeStore()
	member := &domain.Member{ID: uuid.New(), Email: "x@club.de", Role: domain.RoleMitglied, IsActive: false}
	store.byEmail["x@club.de"] = member
	h, _ := newHandler(store, &port.OIDCClaims{}, "")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, member.ID))
	rec := httptest.NewRecorder()
	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRefresh_UnknownMember_Unauthorized(t *testing.T) {
	store := newFakeStore()
	h, _ := newHandler(store, &port.OIDCClaims{}, "")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uuid.New()))
	rec := httptest.NewRecorder()
	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCallback_InvalidState(t *testing.T) {
	store := newFakeStore()
	h, _ := newHandler(store, &port.OIDCClaims{}, "")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?code=c&state=s", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "DIFFERENT"})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Location"), "#error=invalid_state") {
		t.Fatalf("want invalid_state, got %q", rec.Header().Get("Location"))
	}
}

// --- A-005: multiple simultaneous OIDC providers ----------------------------

func TestLogin_SingleProvider_NoQueryParam_UsesIt(t *testing.T) {
	h, _ := newHandler(newFakeStore(), &port.OIDCClaims{}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || !strings.HasSuffix(cookies[0].Value, oidcStateSeparator+"default") {
		t.Fatalf("expected oidc_state cookie ending in %q, got %v", oidcStateSeparator+"default", cookies)
	}
}

func TestLogin_MultipleProviders_RequiresProviderParam(t *testing.T) {
	h, _ := newHandler(newFakeStore(), &port.OIDCClaims{}, "")
	h.oidc["google"] = &fakeOIDC{}
	h.providers = append(h.providers, OIDCProviderInfo{Name: "google", Label: "Google"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when provider is ambiguous, got %d", rec.Code)
	}
}

func TestLogin_UnknownProvider_Returns400(t *testing.T) {
	h, _ := newHandler(newFakeStore(), &port.OIDCClaims{}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login?provider=nope", nil)
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown provider, got %d", rec.Code)
	}
}

func TestLogin_MultipleProviders_SelectsRequestedProvider(t *testing.T) {
	h, _ := newHandler(newFakeStore(), &port.OIDCClaims{}, "")
	h.oidc["google"] = &fakeOIDC{}
	h.providers = append(h.providers, OIDCProviderInfo{Name: "google", Label: "Google"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login?provider=google", nil)
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || !strings.HasSuffix(cookies[0].Value, oidcStateSeparator+"google") {
		t.Fatalf("expected oidc_state cookie for provider google, got %v", cookies)
	}
}

// TestCallback_RoutesToCorrectProvider verifies the provider name carried in
// the state cookie (set by Login) determines which provider's Exchange/
// VerifyIDToken run in Callback — not any client-suppliable query parameter.
func TestCallback_RoutesToCorrectProvider(t *testing.T) {
	store := newFakeStore()
	defaultClaims := &port.OIDCClaims{Subject: "sub-default", Email: "a@club.de", Provider: "default-issuer"}
	googleClaims := &port.OIDCClaims{Subject: "sub-google", Email: "b@club.de", Provider: "google-issuer"}
	h, _ := newHandler(store, defaultClaims, "")
	h.oidc["google"] = &fakeOIDC{claims: googleClaims}
	h.providers = append(h.providers, OIDCProviderInfo{Name: "google", Label: "Google"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?code=c&state=s", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "s" + oidcStateSeparator + "google"})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(store.created) != 1 || store.created[0].Email != "b@club.de" {
		t.Fatalf("expected the google provider's claims to be used, got %+v", store.created)
	}
}

func TestCallback_UnknownProviderInCookie_InvalidState(t *testing.T) {
	h, _ := newHandler(newFakeStore(), &port.OIDCClaims{}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?code=c&state=s", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "s" + oidcStateSeparator + "removed-provider"})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if !strings.Contains(rec.Header().Get("Location"), "#error=invalid_state") {
		t.Fatalf("want invalid_state, got %q", rec.Header().Get("Location"))
	}
}

func TestProviders_ReturnsConfiguredList(t *testing.T) {
	h, _ := newHandler(newFakeStore(), &port.OIDCClaims{}, "")
	h.providers = append(h.providers, OIDCProviderInfo{Name: "google", Label: "Google"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/providers", nil)
	rec := httptest.NewRecorder()
	h.Providers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"name":"default"`) || !strings.Contains(body, `"name":"google"`) {
		t.Fatalf("expected both providers in response, got %s", body)
	}
}

// TestProviders_NilProvidersReturnsEmptyArray guards against json.Marshal's
// nil-slice-becomes-null behavior: api/openapi.yaml documents this endpoint
// as always an array, and internal/testmode/server.go passes a nil
// providers list, so a bare `null` here would be a real response for every
// TEST_MODE deployment, not just a theoretical edge case.
func TestProviders_NilProvidersReturnsEmptyArray(t *testing.T) {
	h := &AuthHandler{providers: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/providers", nil)
	rec := httptest.NewRecorder()
	h.Providers(rec, req)

	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}
