package oidc

import (
	"context"
	"fmt"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/yoadey/shiftmanager/internal/port"
	"golang.org/x/oauth2"
)

// Config holds the OIDC provider configuration.
type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	// RoleAttributePath is an optional JMESPath expression mapping the
	// provider's claims to an application role (see RoleMapper).
	RoleAttributePath string
}

// Service implements port.OIDCService using the coreos go-oidc provider and the
// standard oauth2 authorization-code flow.
type Service struct {
	provider     *coreoidc.Provider
	verifier     *coreoidc.IDTokenVerifier
	oauthConfig  *oauth2.Config
	providerName string
	roles        *RoleMapper
}

var _ port.OIDCService = (*Service)(nil)

// New constructs a Service by discovering the provider metadata from the issuer.
func New(ctx context.Context, cfg Config) (*Service, error) {
	roles, err := NewRoleMapper(cfg.RoleAttributePath)
	if err != nil {
		return nil, err
	}

	provider, err := coreoidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{coreoidc.ScopeOpenID, "profile", "email"}
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	return &Service{
		provider:     provider,
		verifier:     provider.Verifier(&coreoidc.Config{ClientID: cfg.ClientID}),
		oauthConfig:  oauthConfig,
		providerName: cfg.Issuer,
		roles:        roles,
	}, nil
}

// GetAuthURL returns the provider authorization URL for the given state.
func (s *Service) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state)
}

// Exchange converts an authorization code into tokens.
func (s *Service) Exchange(ctx context.Context, code string) (*port.OIDCTokens, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oauth exchange: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token missing from token response")
	}

	return &port.OIDCTokens{
		IDToken:      rawIDToken,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}, nil
}

// VerifyIDToken validates the id_token, extracts the standard claims and
// resolves the application role. Like Grafana, the role expression is first
// evaluated against the id_token claims and, if that yields no role, against
// the UserInfo response (many IdPs only put groups/roles there).
func (s *Service) VerifyIDToken(ctx context.Context, tokens *port.OIDCTokens) (*port.OIDCClaims, error) {
	idToken, err := s.verifier.Verify(ctx, tokens.IDToken)
	if err != nil {
		return nil, fmt.Errorf("verify id token: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("decode id token claims: %w", err)
	}

	out := &port.OIDCClaims{
		Subject:  idToken.Subject,
		Email:    claims.Email,
		Name:     claims.Name,
		Provider: s.providerName,
	}
	if s.roles != nil {
		var raw map[string]any
		if err := idToken.Claims(&raw); err == nil {
			out.Role = s.roles.Resolve(raw)
		}
		if out.Role == "" && tokens.AccessToken != "" {
			out.Role = s.roleFromUserInfo(ctx, tokens.AccessToken)
		}
	}
	return out, nil
}

// roleFromUserInfo evaluates the role expression against the UserInfo
// response. Failures yield no role: the member then waits for manual
// activation, exactly as without a role expression.
func (s *Service) roleFromUserInfo(ctx context.Context, accessToken string) string {
	info, err := s.provider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
	if err != nil {
		return ""
	}
	var raw map[string]any
	if err := info.Claims(&raw); err != nil {
		return ""
	}
	return s.roles.Resolve(raw)
}
