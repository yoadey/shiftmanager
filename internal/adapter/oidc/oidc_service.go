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
}

// Service implements port.OIDCService using the coreos go-oidc provider and the
// standard oauth2 authorization-code flow.
type Service struct {
	provider     *coreoidc.Provider
	verifier     *coreoidc.IDTokenVerifier
	oauthConfig  *oauth2.Config
	providerName string
}

var _ port.OIDCService = (*Service)(nil)

// New constructs a Service by discovering the provider metadata from the issuer.
func New(ctx context.Context, cfg Config) (*Service, error) {
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

// VerifyIDToken validates the raw id_token and extracts standard claims.
func (s *Service) VerifyIDToken(ctx context.Context, rawIDToken string) (*port.OIDCClaims, error) {
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
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

	return &port.OIDCClaims{
		Subject:  idToken.Subject,
		Email:    claims.Email,
		Name:     claims.Name,
		Provider: s.providerName,
	}, nil
}
