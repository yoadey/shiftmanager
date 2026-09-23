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

// RecurrenceFrequency describes how often a recurring event repeats (V-007).
type RecurrenceFrequency string

const (
	RecurrenceFrequencyNone    RecurrenceFrequency = ""
	RecurrenceFrequencyWeekly  RecurrenceFrequency = "weekly"
	RecurrenceFrequencyMonthly RecurrenceFrequency = "monthly"
)

// Valid reports whether f is a frequency occurrences can actually be generated for.
func (f RecurrenceFrequency) Valid() bool {
	return f == RecurrenceFrequencyWeekly || f == RecurrenceFrequencyMonthly
}

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
	// Recurrence (V-007): set on every occurrence generated as part of a
	// series, including the source event that started it, so the UI can show
	// them as a series. Zero values mean the event is not part of a series.
	RecurrenceFrequency RecurrenceFrequency `json:"recurrenceFrequency,omitempty"`
	RecurrenceUntil     *time.Time          `json:"recurrenceUntil,omitempty"`
	RecurrenceGroupID   *uuid.UUID          `json:"recurrenceGroupId,omitempty"`
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

// CanRecur returns true when the event may become the head of a recurring
// series (V-007). A cancelled or already-completed event should not spawn
// fresh occurrences.
func (e *Event) CanRecur() bool {
	return e.Status == EventStatusDraft || e.Status == EventStatusPublished
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

// OccupancyStatus is the human-facing fill state of a shift.
type OccupancyStatus string

const (
	// OccupancyOffen: no helpers registered yet.
	OccupancyOffen OccupancyStatus = "offen"
	// OccupancyTeilweise: at least one helper but below the minimum.
	OccupancyTeilweise OccupancyStatus = "teilweise"
	// OccupancyBesetzt: minimum reached but capacity not yet exhausted.
	OccupancyBesetzt OccupancyStatus = "besetzt"
	// OccupancyAusgebucht: capacity (max helpers) reached.
	OccupancyAusgebucht OccupancyStatus = "ausgebucht"
)

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

// FreeSlots returns how many more helpers fit before the shift is full.
// When MaxHelpers is 0 (unbounded) it returns -1 to signal "unlimited".
func (o ShiftOccupancy) FreeSlots() int {
	if o.MaxHelpers <= 0 {
		return -1
	}
	free := o.MaxHelpers - o.Registered
	if free < 0 {
		return 0
	}
	return free
}

// NeedsMore returns how many more helpers are required to reach the minimum.
// It is 0 once the minimum is satisfied.
func (o ShiftOccupancy) NeedsMore() int {
	need := o.MinHelpers - o.Registered
	if need < 0 {
		return 0
	}
	return need
}

// Status classifies the shift fill state into offen/teilweise/besetzt/ausgebucht.
func (o ShiftOccupancy) Status() OccupancyStatus {
	if o.MaxHelpers > 0 && o.Registered >= o.MaxHelpers {
		return OccupancyAusgebucht
	}
	if o.Registered == 0 {
		return OccupancyOffen
	}
	if o.Registered < o.MinHelpers {
		return OccupancyTeilweise
	}
	return OccupancyBesetzt
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

// EventAttachment is a file (image or document) attached to an event (V-008).
// Whether it's an image (vs. a generic document, e.g. a PDF) is derived by
// callers from ContentType's "image/" prefix rather than tracked separately.
type EventAttachment struct {
	ID          uuid.UUID `json:"id"`
	EventID     uuid.UUID `json:"eventId"`
	FileName    string    `json:"fileName"`
	URL         string    `json:"url"`
	ContentType string    `json:"contentType"`
	SizeBytes   int64     `json:"sizeBytes"`
	UploadedAt  time.Time `json:"uploadedAt"`
}

// Errors for event operations.
var (
	ErrEventNotFound            = fmt.Errorf("event not found")
	ErrEventNotPublishable      = fmt.Errorf("event cannot be published in current state")
	ErrShiftNotFound            = fmt.Errorf("shift not found")
	ErrShiftFull                = fmt.Errorf("shift is fully booked")
	ErrAlreadyRegistered        = fmt.Errorf("already registered for this shift")
	ErrRegistrationNotFound     = fmt.Errorf("registration not found")
	ErrDeregisterDeadlinePassed = fmt.Errorf("deregistration deadline has passed")
	ErrInvalidToken             = fmt.Errorf("invalid or expired confirmation token")
	ErrEventAttachmentNotFound  = fmt.Errorf("event attachment not found")
	ErrInvalidRecurrence        = fmt.Errorf("invalid recurrence configuration")
	ErrAlreadyRecurring         = fmt.Errorf("event is already part of a recurrence series")
)
