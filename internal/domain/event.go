package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventStatus represents the lifecycle status of an event.
type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCompleted EventStatus = "completed"
	EventStatusCancelled EventStatus = "cancelled"
)

// EventVisibility controls who can see the event.
type EventVisibility string

const (
	EventVisibilityPublic  EventVisibility = "public"
	EventVisibilityPrivate EventVisibility = "private"
)

// Event represents a club event that has one or more shifts.
type Event struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Location    string          `json:"location"`
	Category    string          `json:"category"`
	StartDate   time.Time       `json:"startDate"`
	EndDate     time.Time       `json:"endDate"`
	Status      EventStatus     `json:"status"`
	Visibility  EventVisibility `json:"visibility"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// IsMultiDay returns true when the event spans more than one calendar day.
func (e *Event) IsMultiDay() bool {
	return e.EndDate.Format("2006-01-02") != e.StartDate.Format("2006-01-02")
}

// CanPublish returns true when the event can transition to published state.
func (e *Event) CanPublish() bool {
	return e.Status == EventStatusDraft
}

// CanDelete returns true when the event may be deleted.
func (e *Event) CanDelete() bool {
	return e.Status == EventStatusDraft || e.Status == EventStatusCancelled
}

// Shift represents a single work block within an event.
type Shift struct {
	ID                    uuid.UUID `json:"id"`
	EventID               uuid.UUID `json:"eventId"`
	Name                  string    `json:"name"`
	StartAt               time.Time `json:"startAt"`
	EndAt                 time.Time `json:"endAt"`
	MinHelpers            int       `json:"minHelpers"`
	MaxHelpers            int       `json:"maxHelpers"`
	RequiredQualification string    `json:"requiredQualification"`
	Date                  time.Time `json:"date"` // calendar date for multi-day grouping
}

// DurationHours returns the shift length in hours.
func (s *Shift) DurationHours() float64 {
	return s.EndAt.Sub(s.StartAt).Hours()
}

// RegistrationState represents the state of a registration.
type RegistrationState string

const (
	RegistrationStateRegistered RegistrationState = "registered"
	RegistrationStateReserved   RegistrationState = "reserved"
	RegistrationStateConfirmed  RegistrationState = "confirmed"
	RegistrationStateNoShow     RegistrationState = "no_show"
)

// Registration links a member (or guest) to a shift.
type Registration struct {
	ID                uuid.UUID         `json:"id"`
	ShiftID           uuid.UUID         `json:"shiftId"`
	MemberID          *uuid.UUID        `json:"memberId,omitempty"`
	GuestEmail        *string           `json:"guestEmail,omitempty"`
	State             RegistrationState `json:"state"`
	Comment           string            `json:"comment"`
	ReservedUntil     *time.Time        `json:"reservedUntil,omitempty"`
	BookedHours       *float64          `json:"bookedHours,omitempty"`
	ConfirmationToken *uuid.UUID        `json:"confirmationToken,omitempty"`
	CreatedAt         time.Time         `json:"createdAt"`
}

// IsExpired returns true when a reserved registration has passed its deadline.
func (r *Registration) IsExpired(now time.Time) bool {
	if r.State != RegistrationStateReserved {
		return false
	}
	if r.ReservedUntil == nil {
		return false
	}
	return now.After(*r.ReservedUntil)
}

// ShiftOccupancy describes the current fill state of a shift.
type ShiftOccupancy struct {
	ShiftID    uuid.UUID `json:"shiftId"`
	Registered int       `json:"registered"`
	Confirmed  int       `json:"confirmed"`
	MinHelpers int       `json:"minHelpers"`
	MaxHelpers int       `json:"maxHelpers"`
	IsFull     bool      `json:"isFull"`
	IsUnder    bool      `json:"isUnder"`
}

// TimelineDay groups shifts for a single calendar day.
type TimelineDay struct {
	Date   time.Time          `json:"date"`
	Shifts []ShiftWithDetails `json:"shifts"`
}

// ShiftWithDetails combines a shift with its registrations and occupancy.
type ShiftWithDetails struct {
	Shift         Shift          `json:"shift"`
	Registrations []Registration `json:"registrations"`
	Occupancy     ShiftOccupancy `json:"occupancy"`
}

// EventTimeline is the full view of an event including all days and shifts.
type EventTimeline struct {
	Event Event         `json:"event"`
	Days  []TimelineDay `json:"days"`
}

// Errors for event operations.
var (
	ErrEventNotFound      = fmt.Errorf("event not found")
	ErrEventNotPublishable = fmt.Errorf("event cannot be published in current state")
	ErrShiftNotFound      = fmt.Errorf("shift not found")
	ErrShiftFull          = fmt.Errorf("shift is fully booked")
	ErrAlreadyRegistered  = fmt.Errorf("already registered for this shift")
	ErrRegistrationNotFound = fmt.Errorf("registration not found")
	ErrDeregisterDeadlinePassed = fmt.Errorf("deregistration deadline has passed")
	ErrInvalidToken       = fmt.Errorf("invalid or expired confirmation token")
)
