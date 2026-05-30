package port

import (
	"context"
	"time"

	"github.com/yoadey/shiftmanager/internal/domain"
)

// EmailService defines the contract for sending transactional emails.
type EmailService interface {
	SendConfirmation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error
	SendReminder(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event, daysUntil int) error
	SendCancellation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error
	SendKioskConfirmation(ctx context.Context, to string, confirmURL string, shift *domain.Shift, event *domain.Event) error
	SendMissingHoursWarning(ctx context.Context, to string, member *domain.Member, missingHours float64, year *domain.ClubYear) error
}

// OIDCService defines the contract for OIDC provider interactions.
type OIDCService interface {
	// GetAuthURL returns the provider URL to redirect the user to for login.
	GetAuthURL(state string) string
	// Exchange converts an authorization code into an ID token and access token.
	Exchange(ctx context.Context, code string) (*OIDCTokens, error)
	// VerifyIDToken validates the id_token and returns the claims.
	VerifyIDToken(ctx context.Context, rawIDToken string) (*OIDCClaims, error)
}

// OIDCTokens holds the tokens returned after a successful OIDC exchange.
type OIDCTokens struct {
	IDToken      string
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// OIDCClaims holds the standard claims extracted from an OIDC ID token.
type OIDCClaims struct {
	Subject  string
	Email    string
	Name     string
	Provider string
}

// CacheService defines a simple get/set/delete cache contract.
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}

// ErrCacheMiss is returned when the key is not present in the cache.
var ErrCacheMiss = &CacheError{Code: "MISS", Message: "cache miss"}

// CacheError wraps cache-specific error conditions.
type CacheError struct {
	Code    string
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}
