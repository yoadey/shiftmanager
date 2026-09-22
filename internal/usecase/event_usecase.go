package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// EventUsecase handles event and shift business logic.
type EventUsecase struct {
	events        port.EventRepository
	shifts        port.ShiftRepository
	registrations port.RegistrationRepository
	audit         port.AuditRepository
	email         port.EmailService
}

// NewEventUsecase creates a new EventUsecase.
func NewEventUsecase(
	events port.EventRepository,
	shifts port.ShiftRepository,
	registrations port.RegistrationRepository,
	audit port.AuditRepository,
	emailSvc port.EmailService,
) *EventUsecase {
	return &EventUsecase{
		events:        events,
		shifts:        shifts,
		registrations: registrations,
		audit:         audit,
		email:         emailSvc,
	}
}

// CreateEventInput holds the fields needed to create an event.
type CreateEventInput struct {
	Name        string
	Description string
	Location    string
	Category    string
	StartDate   time.Time
	EndDate     time.Time
	Visibility  domain.EventVisibility
	Status      domain.EventStatus
}

// CreateEvent validates and persists a new event. Status defaults to draft.
func (uc *EventUsecase) CreateEvent(ctx context.Context, actorID uuid.UUID, input CreateEventInput) (*domain.Event, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("event name is required")
	}
	if input.EndDate.Before(input.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}
	if input.Visibility == "" {
		input.Visibility = domain.EventVisibilityPublic
	}
	if input.Status == "" {
		input.Status = domain.EventStatusDraft
	}

	now := time.Now().UTC()
	e := &domain.Event{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Location:    input.Location,
		Category:    input.Category,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Status:      input.Status,
		Visibility:  input.Visibility,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.events.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityEvent, e.ID.String(), nil, e)

	return e, nil
}

// UpdateEventInput holds the fields that may be changed on an event.
type UpdateEventInput struct {
	Name        string
	Description string
	Location    string
	Category    string
	StartDate   *time.Time
	EndDate     *time.Time
	Visibility  domain.EventVisibility
	Status      domain.EventStatus // optional; only "cancelled" is acted upon here
}

// UpdateEvent applies changes to an existing event.
func (uc *EventUsecase) UpdateEvent(ctx context.Context, actorID uuid.UUID, id uuid.UUID, input UpdateEventInput) (*domain.Event, error) {
	e, err := uc.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	before := *e

	if input.Name != "" {
		e.Name = input.Name
	}
	if input.Description != "" {
		e.Description = input.Description
	}
	if input.Location != "" {
		e.Location = input.Location
	}
	if input.Category != "" {
		e.Category = input.Category
	}
	if input.StartDate != nil {
		e.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		e.EndDate = *input.EndDate
	}
	if input.Visibility != "" {
		e.Visibility = input.Visibility
	}
	if input.Status != "" {
		e.Status = input.Status
	}
	e.UpdatedAt = time.Now().UTC()

	if e.EndDate.Before(e.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	if err := uc.events.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityEvent, e.ID.String(), before, e)

	// If the event was just cancelled, notify all registered helpers (best-effort).
	if e.Status == domain.EventStatusCancelled && before.Status != domain.EventStatusCancelled && uc.email != nil {
		shifts, err := uc.shifts.FindByEventID(ctx, e.ID)
		if err == nil {
			for _, s := range shifts {
				regs, err := uc.registrations.FindByShiftID(ctx, s.ID)
				if err != nil {
					continue
				}
				for _, reg := range regs {
					var emailTo string
					if reg.GuestEmail != nil {
						emailTo = *reg.GuestEmail
					}
					// For member registrations we don't have the email here; best-effort only.
					if emailTo != "" {
						_ = uc.email.SendCancellation(ctx, emailTo, reg, s, e)
					}
				}
			}
		}
	}

	return e, nil
}

// DeleteEvent removes an event and all its shifts.
func (uc *EventUsecase) DeleteEvent(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	e, err := uc.events.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.events.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDelete, domain.AuditEntityEvent, id.String(), e, nil)

	return nil
}

// PublishEvent transitions an event from draft to published.
func (uc *EventUsecase) PublishEvent(ctx context.Context, actorID uuid.UUID, id uuid.UUID) (*domain.Event, error) {
	e, err := uc.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !e.CanPublish() {
		return nil, domain.ErrEventNotPublishable
	}

	if err := uc.events.UpdateStatus(ctx, id, domain.EventStatusPublished); err != nil {
		return nil, fmt.Errorf("publish event: %w", err)
	}
	e.Status = domain.EventStatusPublished

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityEvent, id.String(), nil, map[string]string{"status": "published"})

	return e, nil
}

// ListEvents returns events matching the filter.
func (uc *EventUsecase) ListEvents(ctx context.Context, filter port.EventFilter) ([]*domain.Event, error) {
	return uc.events.List(ctx, filter)
}

// GetEvent returns a single event by ID.
func (uc *EventUsecase) GetEvent(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	return uc.events.GetByID(ctx, id)
}

// CreateShiftInput holds the fields needed to create a shift within an event.
type CreateShiftInput struct {
	EventID               uuid.UUID
	Name                  string
	StartAt               time.Time
	EndAt                 time.Time
	MinHelpers            int
	MaxHelpers            int
	RequiredQualification string
}

// CreateShift adds a new shift to an event.
func (uc *EventUsecase) CreateShift(ctx context.Context, actorID uuid.UUID, input CreateShiftInput) (*domain.Shift, error) {
	event, err := uc.events.GetByID(ctx, input.EventID)
	if err != nil {
		return nil, err
	}

	if input.EndAt.Before(input.StartAt) || input.EndAt.Equal(input.StartAt) {
		return nil, fmt.Errorf("shift end must be after start")
	}
	if input.MinHelpers < 0 {
		input.MinHelpers = 0
	}
	if input.MaxHelpers < input.MinHelpers {
		input.MaxHelpers = input.MinHelpers
	}

	_ = event // used for validation context
	s := &domain.Shift{
		ID:                    uuid.New(),
		EventID:               input.EventID,
		Name:                  input.Name,
		StartAt:               input.StartAt,
		EndAt:                 input.EndAt,
		MinHelpers:            input.MinHelpers,
		MaxHelpers:            input.MaxHelpers,
		RequiredQualification: input.RequiredQualification,
		Date:                  input.StartAt.Truncate(24 * time.Hour),
	}

	if err := uc.shifts.Create(ctx, s); err != nil {
		return nil, fmt.Errorf("create shift: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityShift, s.ID.String(), nil, s)

	return s, nil
}

// UpdateShiftInput holds the fields that may be changed on a shift.
type UpdateShiftInput struct {
	Name                  string
	StartAt               *time.Time
	EndAt                 *time.Time
	MinHelpers            *int
	MaxHelpers            *int
	RequiredQualification string
}

// UpdateShift applies changes to an existing shift.
func (uc *EventUsecase) UpdateShift(ctx context.Context, actorID uuid.UUID, id uuid.UUID, input UpdateShiftInput) (*domain.Shift, error) {
	s, err := uc.shifts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	before := *s

	if input.Name != "" {
		s.Name = input.Name
	}
	if input.StartAt != nil {
		s.StartAt = *input.StartAt
		s.Date = input.StartAt.Truncate(24 * time.Hour)
	}
	if input.EndAt != nil {
		s.EndAt = *input.EndAt
	}
	if input.MinHelpers != nil {
		s.MinHelpers = *input.MinHelpers
	}
	if input.MaxHelpers != nil {
		s.MaxHelpers = *input.MaxHelpers
	}
	if input.RequiredQualification != "" {
		s.RequiredQualification = input.RequiredQualification
	}

	if s.EndAt.Before(s.StartAt) || s.EndAt.Equal(s.StartAt) {
		return nil, fmt.Errorf("shift end must be after start")
	}

	if err := uc.shifts.Update(ctx, s); err != nil {
		return nil, fmt.Errorf("update shift: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityShift, s.ID.String(), before, s)

	return s, nil
}

// DeleteShift removes a shift and all its registrations.
func (uc *EventUsecase) DeleteShift(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	s, err := uc.shifts.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.shifts.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete shift: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDelete, domain.AuditEntityShift, id.String(), s, nil)

	return nil
}

// GetEventTimeline returns an event with all its shifts grouped by day, each with
// registrations and occupancy data.
func (uc *EventUsecase) GetEventTimeline(ctx context.Context, eventID uuid.UUID) (*domain.EventTimeline, error) {
	event, err := uc.events.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	shifts, err := uc.shifts.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("load shifts: %w", err)
	}

	// Group shifts by calendar date.
	dayMap := make(map[string]*domain.TimelineDay)
	var dayOrder []string

	for _, s := range shifts {
		dateKey := s.Date.Format("2006-01-02")
		if _, exists := dayMap[dateKey]; !exists {
			dayMap[dateKey] = &domain.TimelineDay{Date: s.Date}
			dayOrder = append(dayOrder, dateKey)
		}

		regs, err := uc.registrations.FindByShiftID(ctx, s.ID)
		if err != nil {
			return nil, fmt.Errorf("load registrations for shift %s: %w", s.ID, err)
		}

		occ := computeOccupancy(s, regs)

		dayMap[dateKey].Shifts = append(dayMap[dateKey].Shifts, domain.ShiftWithDetails{
			Shift:         *s,
			Registrations: derefRegistrations(regs),
			Occupancy:     occ,
		})
	}

	sort.Strings(dayOrder)
	days := make([]domain.TimelineDay, 0, len(dayOrder))
	for _, key := range dayOrder {
		days = append(days, *dayMap[key])
	}

	return &domain.EventTimeline{Event: *event, Days: days}, nil
}

// computeOccupancy calculates occupancy stats for a shift.
func computeOccupancy(s *domain.Shift, regs []*domain.Registration) domain.ShiftOccupancy {
	occ := domain.ShiftOccupancy{
		ShiftID:    s.ID,
		MinHelpers: s.MinHelpers,
		MaxHelpers: s.MaxHelpers,
	}
	for _, r := range regs {
		switch r.State {
		case domain.RegistrationStateRegistered, domain.RegistrationStateReserved:
			occ.Registered++
		case domain.RegistrationStateConfirmed:
			occ.Confirmed++
			occ.Registered++
		}
	}
	occ.IsFull = s.MaxHelpers > 0 && occ.Registered >= s.MaxHelpers
	occ.IsUnder = occ.Registered < s.MinHelpers
	return occ
}

func derefRegistrations(in []*domain.Registration) []domain.Registration {
	out := make([]domain.Registration, len(in))
	for i, r := range in {
		out[i] = *r
	}
	return out
}

// CopyEvent creates a new event as a copy of an existing one (including all shifts).
// The copy gets status draft and the name "Kopie von <original name>".
func (uc *EventUsecase) CopyEvent(ctx context.Context, actorID uuid.UUID, id uuid.UUID) (*domain.Event, error) {
	orig, err := uc.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	shifts, err := uc.shifts.FindByEventID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load shifts: %w", err)
	}

	now := time.Now().UTC()
	newEvent := &domain.Event{
		ID:          uuid.New(),
		Name:        "Kopie von " + orig.Name,
		Description: orig.Description,
		Location:    orig.Location,
		Category:    orig.Category,
		StartDate:   orig.StartDate,
		EndDate:     orig.EndDate,
		Status:      domain.EventStatusDraft,
		Visibility:  orig.Visibility,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.events.Create(ctx, newEvent); err != nil {
		return nil, fmt.Errorf("create copy event: %w", err)
	}

	for _, s := range shifts {
		newShift := &domain.Shift{
			ID:                    uuid.New(),
			EventID:               newEvent.ID,
			Name:                  s.Name,
			StartAt:               s.StartAt,
			EndAt:                 s.EndAt,
			MinHelpers:            s.MinHelpers,
			MaxHelpers:            s.MaxHelpers,
			RequiredQualification: s.RequiredQualification,
			Date:                  s.Date,
		}
		if err := uc.shifts.Create(ctx, newShift); err != nil {
			return nil, fmt.Errorf("create copy shift: %w", err)
		}
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityEvent, newEvent.ID.String(), nil, map[string]string{"copiedFrom": id.String()})

	return newEvent, nil
}

func (uc *EventUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}
