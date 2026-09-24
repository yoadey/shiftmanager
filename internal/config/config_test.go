package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviders_LegacySingleProvider(t *testing.T) {
	cfg := &Config{
		OIDCIssuer:       "https://idp.example/realm",
		OIDCClientID:     "shiftmanager",
		OIDCClientSecret: "secret",
		OIDCRedirectURL:  "https://app.example/api/v1/auth/callback",
	}

	providers := cfg.Providers()

	require.Len(t, providers, 1)
	assert.Equal(t, "default", providers[0].Name)
	assert.Equal(t, "https://idp.example/realm", providers[0].Issuer)
	assert.Equal(t, "shiftmanager", providers[0].ClientID)
	assert.Equal(t, "secret", providers[0].ClientSecret)
	assert.Equal(t, "https://app.example/api/v1/auth/callback", providers[0].RedirectURL)
}

func TestProviders_MultipleNamedProviders(t *testing.T) {
	t.Setenv("OIDC_VEREIN_ISSUER", "https://verein.example")
	t.Setenv("OIDC_VEREIN_CLIENT_ID", "verein-client")
	t.Setenv("OIDC_VEREIN_CLIENT_SECRET", "verein-secret")
	t.Setenv("OIDC_VEREIN_LABEL", "Vereins-SSO")
	t.Setenv("OIDC_GOOGLE_ISSUER", "https://accounts.google.com")
	t.Setenv("OIDC_GOOGLE_CLIENT_ID", "google-client")
	t.Setenv("OIDC_GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("OIDC_GOOGLE_LABEL", "Google")
	t.Setenv("OIDC_REDIRECT_URL", "https://app.example/api/v1/auth/callback")

	cfg := &Config{
		OIDCRedirectURL:   "https://app.example/api/v1/auth/callback",
		OIDCProviderNames: []string{"verein", "google"},
	}

	providers := cfg.Providers()

	require.Len(t, providers, 2)
	assert.Equal(t, "verein", providers[0].Name)
	assert.Equal(t, "Vereins-SSO", providers[0].Label)
	assert.Equal(t, "https://verein.example", providers[0].Issuer)
	assert.Equal(t, "https://app.example/api/v1/auth/callback", providers[0].RedirectURL, "falls back to the shared OIDC_REDIRECT_URL")
	assert.Equal(t, "google", providers[1].Name)
	assert.Equal(t, "Google", providers[1].Label)
	assert.Equal(t, "https://accounts.google.com", providers[1].Issuer)
}

func TestProviders_PerProviderRedirectURLOverridesShared(t *testing.T) {
	t.Setenv("OIDC_GOOGLE_ISSUER", "https://accounts.google.com")
	t.Setenv("OIDC_GOOGLE_REDIRECT_URL", "https://app.example/api/v1/auth/callback/google")

	cfg := &Config{
		OIDCRedirectURL:   "https://app.example/api/v1/auth/callback",
		OIDCProviderNames: []string{"google"},
	}

	providers := cfg.Providers()

	require.Len(t, providers, 1)
	assert.Equal(t, "https://app.example/api/v1/auth/callback/google", providers[0].RedirectURL)
}

func TestProviders_LabelDefaultsToName(t *testing.T) {
	cfg := &Config{OIDCProviderNames: []string{"verein"}}

	providers := cfg.Providers()

	require.Len(t, providers, 1)
	assert.Equal(t, "verein", providers[0].Label)
}

func TestProviders_DeduplicatesCaseInsensitiveNames(t *testing.T) {
	t.Setenv("OIDC_GOOGLE_LABEL", "Google")
	cfg := &Config{OIDCProviderNames: []string{"google", "Google", "GOOGLE"}}

	providers := cfg.Providers()

	require.Len(t, providers, 1, "case-variant repeats of the same name must collapse to one provider")
	assert.Equal(t, "google", providers[0].Name)
}

func TestGetEnvList(t *testing.T) {
	t.Setenv("TEST_LIST", " a, b ,,c ")
	assert.Equal(t, []string{"a", "b", "c"}, getEnvList("TEST_LIST"))

	_ = os.Unsetenv("TEST_LIST_UNSET")
	assert.Nil(t, getEnvList("TEST_LIST_UNSET"))
}
