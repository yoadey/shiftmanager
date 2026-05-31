package config

import (
	"fmt"
	"os"
	"strconv"
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

	// OIDC
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string

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

	// Logging
	LogLevel string

	// Business rules
	ReservationHours    int
	DeregisterDeadlineH int

	// Uploads (logo storage, B-004)
	UploadDir string
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
		SMTPHost:            getEnv("SMTP_HOST", "localhost"),
		SMTPUser:            getEnv("SMTP_USER", ""),
		SMTPPass:            getEnv("SMTP_PASS", ""),
		SMTPFrom:            getEnv("SMTP_FROM", "noreply@shiftmanager.local"),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTExpiration:       24 * time.Hour,
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		KioskEnabled:        getEnvBool("KIOSK_ENABLED", true),
		ReservationHours:    getEnvInt("RESERVATION_HOURS", 48),
		DeregisterDeadlineH: getEnvInt("DEREGISTER_DEADLINE_H", 24),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
	}

	var err error
	cfg.SMTPPort, err = strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
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
