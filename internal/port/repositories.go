package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
)

// MemberRepository defines persistence operations for members.
type MemberRepository interface {
	Create(ctx context.Context, m *domain.Member) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error)
	GetByEmail(ctx context.Context, email string) (*domain.Member, error)
	GetByOIDCSubject(ctx context.Context, provider, subject string) (*domain.Member, error)
	List(ctx context.Context, filter MemberFilter) ([]*domain.Member, error)
	Update(ctx context.Context, m *domain.Member) error
	Deactivate(ctx context.Context, id uuid.UUID, leftAt time.Time) error
	LinkOIDC(ctx context.Context, link *domain.OIDCLink) error
	GetOIDCLinks(ctx context.Context, memberID uuid.UUID) ([]*domain.OIDCLink, error)
	Count(ctx context.Context) (int, error)
}

// MemberFilter holds optional filters when listing members.
type MemberFilter struct {
	IsActive *bool
	Search   string // matches first/last name or email
	Limit    int
	Offset   int
}

// EventRepository defines persistence operations for events.
type EventRepository interface {
	Create(ctx context.Context, e *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	List(ctx context.Context, filter EventFilter) ([]*domain.Event, error)
	Update(ctx context.Context, e *domain.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EventStatus) error
}

// EventFilter holds optional filters when listing events.
type EventFilter struct {
	Status     *domain.EventStatus
	Visibility *domain.EventVisibility
	FromDate   *time.Time
	ToDate     *time.Time
	Limit      int
	Offset     int
}

// ShiftRepository defines persistence operations for shifts.
type ShiftRepository interface {
	Create(ctx context.Context, s *domain.Shift) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Shift, error)
	FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain.Shift, error)
	Update(ctx context.Context, s *domain.Shift) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RegistrationRepository defines persistence operations for shift registrations.
type RegistrationRepository interface {
	Create(ctx context.Context, r *domain.Registration) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Registration, error)
	GetByToken(ctx context.Context, token uuid.UUID) (*domain.Registration, error)
	FindByShiftID(ctx context.Context, shiftID uuid.UUID) ([]*domain.Registration, error)
	FindByMemberID(ctx context.Context, memberID uuid.UUID) ([]*domain.Registration, error)
	FindByMemberAndShift(ctx context.Context, memberID uuid.UUID, shiftID uuid.UUID) (*domain.Registration, error)
	FindByGuestEmailAndShift(ctx context.Context, guestEmail string, shiftID uuid.UUID) (*domain.Registration, error)
	CountActiveByShift(ctx context.Context, shiftID uuid.UUID) (int, error)
	Update(ctx context.Context, r *domain.Registration) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListUnconfirmedExpiredReservations(ctx context.Context, before time.Time) ([]*domain.Registration, error)
}

// HourRepository defines persistence operations for hour entries and club years.
type HourRepository interface {
	CreateEntry(ctx context.Context, e *domain.HourEntry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.HourEntry, error)
	FindEntriesByMemberAndYear(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.HourEntry, error)
	FindEntriesByYear(ctx context.Context, clubYearID uuid.UUID) ([]*domain.HourEntry, error)
	UpdateEntry(ctx context.Context, e *domain.HourEntry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error

	CreateClubYear(ctx context.Context, y *domain.ClubYear) error
	GetActiveClubYear(ctx context.Context) (*domain.ClubYear, error)
	GetClubYearByID(ctx context.Context, id uuid.UUID) (*domain.ClubYear, error)
	ListClubYears(ctx context.Context) ([]*domain.ClubYear, error)

	GetHourTarget(ctx context.Context, memberID, clubYearID uuid.UUID) (*domain.HourTarget, error)
	UpsertHourTarget(ctx context.Context, t *domain.HourTarget) error
}

// AuditRepository defines persistence operations for the audit log.
// Entries are append-only; update and delete operations are intentionally absent.
type AuditRepository interface {
	Insert(ctx context.Context, e *domain.AuditEntry) error
	List(ctx context.Context, filter AuditFilter) ([]*domain.AuditEntry, error)
}

// AuditFilter holds optional filters when querying the audit log.
type AuditFilter struct {
	ActorID  *uuid.UUID
	Entity   string
	EntityID string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

// SettingsRepository defines persistence operations for application settings.
type SettingsRepository interface {
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
	GetAllSettings(ctx context.Context) (map[string]string, error)

	GetBranding(ctx context.Context) (*domain.BrandingConfig, error)
	UpdateBranding(ctx context.Context, b *domain.BrandingConfig) error

	GetFeeTiers(ctx context.Context, clubYearID uuid.UUID) ([]*domain.FeeTier, error)
	ReplaceFeeTiers(ctx context.Context, clubYearID uuid.UUID, tiers []*domain.FeeTier) error
}
