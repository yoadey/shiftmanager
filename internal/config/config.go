package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port    string
	BaseURL string

	// Database
	DatabaseURL string

	// Redis (optional)
	RedisURL string

	// OIDC (single-provider — kept for backward compatibility; see
	// OIDCProviderNames/OIDCProvider below for the multi-provider path, A-005).
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	// OIDCProviderNames lists additional named providers (A-005), read from
	// OIDC_PROVIDERS (comma-separated). Empty unless multi-provider login is
	// configured — see Providers().
	OIDCProviderNames []string
	// LoginRedirectURL is the SPA route the backend redirects to after the
	// OIDC callback, carrying the JWT (or an error code) in the URL fragment.
	// Relative paths resolve against the server's own origin.
	LoginRedirectURL string
	// BootstrapAdminEmail, when set, auto-provisions/promotes the member with
	// this e-mail to an active administrator on login (first-admin bootstrap).
	BootstrapAdminEmail string

	// SMTP
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// Feature flags
	KioskEnabled bool
	TestMode     bool // TEST_MODE=true: in-memory repos, no DB required, /dev/token active

	// Logging
	LogLevel string

	// Business rules
	ReservationHours    int
	DeregisterDeadlineH int

	// Uploads (logo storage B-004, event attachments V-008)
	UploadDir string

	// MediaStorage selects where uploaded media is written: "local" (default,
	// UploadDir on disk) or "s3" (T-013, S3-compatible object storage).
	MediaStorage      string
	S3Endpoint        string // only needed for non-AWS providers (MinIO, Hetzner, …); empty = AWS S3
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3ForcePathStyle  bool
	// S3PublicBaseURL overrides the URL objects are served from (e.g. a CDN
	// in front of the bucket). Optional.
	S3PublicBaseURL string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		BaseURL:             getEnv("BASE_URL", "http://localhost:8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		RedisURL:            getEnv("REDIS_URL", ""),
		OIDCIssuer:          getEnv("OIDC_ISSUER", ""),
		OIDCClientID:        getEnv("OIDC_CLIENT_ID", ""),
		OIDCClientSecret:    getEnv("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:     getEnv("OIDC_REDIRECT_URL", ""),
		LoginRedirectURL:    getEnv("LOGIN_REDIRECT_URL", "/auth/callback"),
		BootstrapAdminEmail: getEnv("BOOTSTRAP_ADMIN_EMAIL", ""),
		SMTPHost:            getEnv("SMTP_HOST", "localhost"),
		SMTPUser:            getEnv("SMTP_USER", ""),
		SMTPPass:            getEnv("SMTP_PASS", ""),
		SMTPFrom:            getEnv("SMTP_FROM", "noreply@shiftmanager.local"),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTExpiration:       24 * time.Hour,
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		KioskEnabled:        getEnvBool("KIOSK_ENABLED", true),
		TestMode:            getEnvBool("TEST_MODE", false),
		ReservationHours:    getEnvInt("RESERVATION_HOURS", 48),
		DeregisterDeadlineH: getEnvInt("DEREGISTER_DEADLINE_H", 24),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		MediaStorage:        getEnv("MEDIA_STORAGE", "local"),
		S3Endpoint:          getEnv("S3_ENDPOINT", ""),
		S3Region:            getEnv("S3_REGION", ""),
		S3Bucket:            getEnv("S3_BUCKET", ""),
		S3AccessKeyID:       getEnv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey:   getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle:    getEnvBool("S3_FORCE_PATH_STYLE", false),
		S3PublicBaseURL:     getEnv("S3_PUBLIC_BASE_URL", ""),
		OIDCProviderNames:   getEnvList("OIDC_PROVIDERS"),
	}

	var err error
	cfg.SMTPPort, err = strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	if !cfg.TestMode && cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required (or set TEST_MODE=true)")
	}
	if cfg.JWTSecret == "" {
		if cfg.TestMode {
			cfg.JWTSecret = "test-secret-do-not-use-in-production"
		} else {
			return nil, fmt.Errorf("JWT_SECRET is required")
		}
	}

	switch cfg.MediaStorage {
	case "local":
	case "s3":
		if cfg.S3Bucket == "" {
			return nil, fmt.Errorf("S3_BUCKET is required when MEDIA_STORAGE=s3")
		}
	default:
		return nil, fmt.Errorf("invalid MEDIA_STORAGE %q: must be \"local\" or \"s3\"", cfg.MediaStorage)
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

func getEnvInt(key string, defaultVal int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

// getEnvList splits a comma-separated environment variable into trimmed,
// non-empty entries. Returns nil if the variable is unset or empty.
func getEnvList(key string) []string {
	val := getEnv(key, "")
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// OIDCProviderConfig configures one named OIDC provider (A-005).
type OIDCProviderConfig struct {
	Name         string
	Label        string
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Providers returns the configured OIDC providers, in the order login
// buttons should be displayed.
//
// With OIDC_PROVIDERS unset (the common case), this is the single legacy
// provider named "default", built from OIDC_ISSUER/OIDC_CLIENT_ID/
// OIDC_CLIENT_SECRET/OIDC_REDIRECT_URL — existing single-provider
// deployments are unaffected by A-005.
//
// With OIDC_PROVIDERS set to a comma-separated list of short names (e.g.
// "verein,google"), each name N is looked up as OIDC_<N>_ISSUER/_CLIENT_ID/
// _CLIENT_SECRET/_LABEL (N upper-cased), falling back to OIDC_REDIRECT_URL
// for _REDIRECT_URL (the callback endpoint registered with every IdP is
// normally the same one) and to N itself for _LABEL. This reads the
// environment directly rather than from pre-loaded Config fields, since the
// variable names are only known once OIDC_PROVIDERS itself is read — there
// is no fixed field to load them into ahead of time.
func (c *Config) Providers() []OIDCProviderConfig {
	if len(c.OIDCProviderNames) == 0 {
		return []OIDCProviderConfig{{
			Name:         "default",
			Label:        getEnv("OIDC_LABEL", "Vereinskonto"),
			Issuer:       c.OIDCIssuer,
			ClientID:     c.OIDCClientID,
			ClientSecret: c.OIDCClientSecret,
			RedirectURL:  c.OIDCRedirectURL,
		}}
	}

	providers := make([]OIDCProviderConfig, 0, len(c.OIDCProviderNames))
	seen := make(map[string]bool, len(c.OIDCProviderNames))
	for _, name := range c.OIDCProviderNames {
		// Names are looked up case-insensitively via strings.ToUpper below
		// (and used as the ?provider= value / map key elsewhere), so two
		// entries differing only in case would silently read the same
		// OIDC_<NAME>_* variables and render as duplicate login buttons.
		// Normalize to lower-case and drop repeats, keeping the first.
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true

		prefix := "OIDC_" + strings.ToUpper(name) + "_"
		providers = append(providers, OIDCProviderConfig{
			Name:         key,
			Label:        getEnv(prefix+"LABEL", name),
			Issuer:       getEnv(prefix+"ISSUER", ""),
			ClientID:     getEnv(prefix+"CLIENT_ID", ""),
			ClientSecret: getEnv(prefix+"CLIENT_SECRET", ""),
			RedirectURL:  getEnv(prefix+"REDIRECT_URL", c.OIDCRedirectURL),
		})
	}
	return providers
}
