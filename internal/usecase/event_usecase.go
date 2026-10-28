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
	members       port.MemberRepository
	attachments   port.EventAttachmentRepository
}

// NewEventUsecase creates a new EventUsecase.
func NewEventUsecase(
	events port.EventRepository,
	shifts port.ShiftRepository,
	registrations port.RegistrationRepository,
	audit port.AuditRepository,
	emailSvc port.EmailService,
	members port.MemberRepository,
	attachments port.EventAttachmentRepository,
) *EventUsecase {
	return &EventUsecase{
		events:        events,
		shifts:        shifts,
		registrations: registrations,
		audit:         audit,
		email:         emailSvc,
		members:       members,
		attachments:   attachments,
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

// DeleteEvent removes an event and all its shifts, and returns its
// attachments (if any) so the caller can also remove their underlying files.
// The event itself is deleted first: on Postgres the migration's ON DELETE
// CASCADE removes the attachment rows atomically with it, while GORM's
// AutoMigrate (SQLite) defines no such FK, so the explicit cleanup below
// (best-effort, since the event is already gone at that point and its
// failure must not undo that or block returning the files to remove) covers
// that case too.
func (uc *EventUsecase) DeleteEvent(ctx context.Context, actorID uuid.UUID, id uuid.UUID) ([]*domain.EventAttachment, error) {
	e, err := uc.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	attachments, err := uc.attachments.ListByEvent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}

	if err := uc.events.Delete(ctx, id); err != nil {
		return nil, fmt.Errorf("delete event: %w", err)
	}

	_ = uc.attachments.DeleteByEvent(ctx, id)

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDelete, domain.AuditEntityEvent, id.String(), e, nil)

	return attachments, nil
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

	uc.notifyNewEvent(ctx, e)

	return e, nil
}

// CompleteEvent transitions a published event to completed and bulk-confirms all
// registrations that are still in the "registered" state across all its shifts.
func (uc *EventUsecase) CompleteEvent(ctx context.Context, actorID uuid.UUID, id uuid.UUID) (*domain.Event, error) {
	e, err := uc.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !e.CanComplete() {
		return nil, domain.ErrEventNotCompletable
	}

	if err := uc.events.UpdateStatus(ctx, id, domain.EventStatusCompleted); err != nil {
		return nil, fmt.Errorf("complete event: %w", err)
	}
	e.Status = domain.EventStatusCompleted

	shifts, err := uc.shifts.FindByEventID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load shifts: %w", err)
	}
	for _, s := range shifts {
		if _, err := uc.registrations.ConfirmRegisteredByShift(ctx, s.ID); err != nil {
			return nil, fmt.Errorf("confirm registrations for shift %s: %w", s.ID, err)
		}
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityEvent, id.String(), nil, map[string]string{"status": "completed"})

	return e, nil
}

// notifyNewEvent sends the opt-in "new event published" mail (KANN) to every
// active member who opted in. Best-effort: a failed lookup or send never
// fails the publish.
func (uc *EventUsecase) notifyNewEvent(ctx context.Context, e *domain.Event) {
	if uc.email == nil || uc.members == nil {
		return
	}
	active := true
	members, err := uc.members.List(ctx, port.MemberFilter{IsActive: &active})
	if err != nil {
		return
	}
	data := map[string]any{
		"EventName": e.Name,
		"Location":  e.Location,
		"StartAt":   e.StartDate.Format("02.01.2006 15:04"),
	}
	for _, m := range members {
		if !m.NotifyNewEvents {
			continue
		}
		_ = uc.email.SendByTemplate(ctx, m.Email, domain.EmailTemplateNewEvent, data)
	}
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

	if err := uc.copyShifts(ctx, shifts, newEvent.ID, 0); err != nil {
		return nil, fmt.Errorf("create copy shift: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityEvent, newEvent.ID.String(), nil, map[string]string{"copiedFrom": id.String()})

	return newEvent, nil
}

// copyShifts creates copies of shifts under a new event, shifting each
// shift's start/end by delta (0 for an exact copy, e.g. CopyEvent; the
// per-occurrence offset for GenerateRecurrence). Date is always recomputed
// from the (possibly shifted) start rather than copied, since a shifted
// occurrence generally falls on a different calendar day.
func (uc *EventUsecase) copyShifts(ctx context.Context, shifts []*domain.Shift, newEventID uuid.UUID, delta time.Duration) error {
	for _, s := range shifts {
		newShift := &domain.Shift{
			ID:                    uuid.New(),
			EventID:               newEventID,
			Name:                  s.Name,
			StartAt:               s.StartAt.Add(delta),
			EndAt:                 s.EndAt.Add(delta),
			MinHelpers:            s.MinHelpers,
			MaxHelpers:            s.MaxHelpers,
			RequiredQualification: s.RequiredQualification,
		}
		newShift.Date = newShift.StartAt.Truncate(24 * time.Hour)
		if err := uc.shifts.Create(ctx, newShift); err != nil {
			return err
		}
	}
	return nil
}

// cleanupFailedOccurrences best-effort removes occurrence events (and their
// shifts) already created earlier in a GenerateRecurrence call that failed
// partway through, so a partial failure doesn't leave orphaned draft events
// with no source event tracking their RecurrenceGroupID. There is no
// database transaction around the batch (matching the rest of this usecase,
// e.g. DeleteEvent's attachment cleanup), so this is a best-effort cleanup,
// not a guarantee; its own errors are intentionally swallowed so a cleanup
// failure doesn't mask the original error being returned to the caller.
func (uc *EventUsecase) cleanupFailedOccurrences(ctx context.Context, occurrences []*domain.Event) {
	for _, occ := range occurrences {
		if shifts, err := uc.shifts.FindByEventID(ctx, occ.ID); err == nil {
			for _, s := range shifts {
				_ = uc.shifts.Delete(ctx, s.ID)
			}
		}
		_ = uc.events.Delete(ctx, occ.ID)
	}
}

// maxRecurrenceOccurrences caps how many follow-up events GenerateRecurrence
// creates in one call, so a caller can't trigger unbounded event/shift
// creation (e.g. a weekly recurrence with an "until" decades out).
const maxRecurrenceOccurrences = 104

// GenerateRecurrence turns an existing event into a recurring series (V-007):
// it stamps the event itself with the recurrence config and creates follow-up
// occurrences (each a full copy of the event, including its shifts, offset by
// one week/month per step) up to and including `until`. All occurrences,
// including the source event, share a RecurrenceGroupID so the UI can
// present them as a series. Returns the updated source event followed by the
// newly created occurrences, in chronological order.
func (uc *EventUsecase) GenerateRecurrence(ctx context.Context, actorID uuid.UUID, id uuid.UUID, frequency domain.RecurrenceFrequency, until time.Time) ([]*domain.Event, error) {
	if !frequency.Valid() {
		return nil, domain.ErrInvalidRecurrence
	}

	orig, err := uc.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !orig.CanRecur() {
		return nil, domain.ErrInvalidRecurrence
	}
	if orig.RecurrenceGroupID != nil {
		return nil, domain.ErrAlreadyRecurring
	}
	if !until.After(orig.StartDate) {
		return nil, domain.ErrInvalidRecurrence
	}

	starts := recurrenceOccurrenceStarts(orig.StartDate, until, frequency)
	if len(starts) == 0 {
		return nil, domain.ErrInvalidRecurrence
	}
	if len(starts) > maxRecurrenceOccurrences {
		return nil, fmt.Errorf("recurrence range too large: max %d occurrences", maxRecurrenceOccurrences)
	}

	shifts, err := uc.shifts.FindByEventID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load shifts: %w", err)
	}

	// Occurrences are created before the source event is stamped as
	// recurring (below), so a failure partway through never leaves the
	// source event permanently locked out of a retry by the
	// ErrAlreadyRecurring guard above.
	groupID := uuid.New()
	duration := orig.EndDate.Sub(orig.StartDate)
	aid := actorID
	var occurrences []*domain.Event

	for _, start := range starts {
		now := time.Now().UTC()
		occurrence := &domain.Event{
			ID:                  uuid.New(),
			Name:                orig.Name,
			Description:         orig.Description,
			Location:            orig.Location,
			Category:            orig.Category,
			StartDate:           start,
			EndDate:             start.Add(duration),
			Status:              domain.EventStatusDraft,
			Visibility:          orig.Visibility,
			RecurrenceFrequency: frequency,
			RecurrenceUntil:     &until,
			RecurrenceGroupID:   &groupID,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
		if err := uc.events.Create(ctx, occurrence); err != nil {
			uc.cleanupFailedOccurrences(ctx, occurrences)
			return nil, fmt.Errorf("create recurring occurrence: %w", err)
		}

		if err := uc.copyShifts(ctx, shifts, occurrence.ID, start.Sub(orig.StartDate)); err != nil {
			uc.cleanupFailedOccurrences(ctx, append(occurrences, occurrence))
			return nil, fmt.Errorf("create recurring shift: %w", err)
		}

		occurrences = append(occurrences, occurrence)
	}

	// Atomic, conditional on the source event still not being part of a
	// series: closes the race where two concurrent calls both passed the
	// RecurrenceGroupID nil check above and would otherwise both proceed to
	// stamp the source event (last write wins), leaving the loser's
	// occurrences orphaned with no source event pointing at their group.
	// Here, only the winner's occurrences survive; the loser's are rolled
	// back via cleanupFailedOccurrences, same as any other failure above.
	applied, err := uc.events.MarkRecurring(ctx, id, frequency, until, groupID)
	if err != nil {
		uc.cleanupFailedOccurrences(ctx, occurrences)
		return nil, fmt.Errorf("update source event: %w", err)
	}
	if !applied {
		uc.cleanupFailedOccurrences(ctx, occurrences)
		return nil, domain.ErrAlreadyRecurring
	}
	before := *orig
	orig.RecurrenceFrequency = frequency
	orig.RecurrenceUntil = &until
	orig.RecurrenceGroupID = &groupID
	orig.UpdatedAt = time.Now().UTC()

	// Audit entries are only written now, once the whole batch (occurrences
	// and the source event's own stamp) has actually succeeded — a partial
	// failure rolls occurrences back via cleanupFailedOccurrences above, and
	// the audit log is append-only (no compensating delete entries), so
	// writing "create" audit entries any earlier would leave permanent, stale
	// records pointing at events that no longer exist.
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityEvent, id.String(), before, orig)
	for _, occ := range occurrences {
		_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityEvent, occ.ID.String(), nil, map[string]string{"recurrenceOf": id.String()})
	}

	return append([]*domain.Event{orig}, occurrences...), nil
}

// recurrenceOccurrenceStarts returns the start times of the follow-up
// occurrences (i.e. excluding start itself) stepped weekly/monthly from
// start, up to and including until. It stops early (without erroring) once
// more than maxRecurrenceOccurrences+1 dates have been produced, leaving the
// caller to reject the oversized result.
func recurrenceOccurrenceStarts(start, until time.Time, frequency domain.RecurrenceFrequency) []time.Time {
	var starts []time.Time
	for n := 1; ; n++ {
		var next time.Time
		switch frequency {
		case domain.RecurrenceFrequencyWeekly:
			next = start.AddDate(0, 0, 7*n)
		case domain.RecurrenceFrequencyMonthly:
			next = addMonthsClamped(start, n)
		default:
			return starts
		}
		if next.After(until) {
			return starts
		}
		starts = append(starts, next)
		if len(starts) > maxRecurrenceOccurrences {
			return starts
		}
	}
}

// addMonthsClamped adds n calendar months to t. Unlike time.Time.AddDate,
// which overflows a day-of-month that doesn't exist in the target month into
// the following month (e.g. Jan 31 + 1 month rolls over to Mar 3, not Feb),
// this clamps to the target month's last day instead (Jan 31 + 1 month ->
// Feb 28/29). Each occurrence is computed straight from the original start
// date (not cumulatively from the previous occurrence), so a clamped
// occurrence never drags later ones off course: Jan 31 + 2 months lands on
// Mar 31, not Mar 28.
func addMonthsClamped(t time.Time, n int) time.Time {
	firstOfTarget := time.Date(t.Year(), t.Month()+time.Month(n), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	lastDayOfTarget := firstOfTarget.AddDate(0, 1, -1).Day()
	day := t.Day()
	if day > lastDayOfTarget {
		day = lastDayOfTarget
	}
	return time.Date(firstOfTarget.Year(), firstOfTarget.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// AddAttachment records a file (image or document) uploaded against an event
// (V-008). The handler owns actually writing the file to storage; this only
// persists the resulting URL and metadata.
func (uc *EventUsecase) AddAttachment(ctx context.Context, actorID uuid.UUID, eventID uuid.UUID, fileName, url, contentType string, sizeBytes int64) (*domain.EventAttachment, error) {
	if _, err := uc.events.GetByID(ctx, eventID); err != nil {
		return nil, err
	}

	a := &domain.EventAttachment{
		ID:          uuid.New(),
		EventID:     eventID,
		FileName:    fileName,
		URL:         url,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		UploadedAt:  time.Now().UTC(),
	}
	if err := uc.attachments.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("create attachment: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityEvent, eventID.String(), nil, map[string]string{"attachmentAdded": fileName})

	return a, nil
}

// SetHeaderImage stores a dedicated header image for an event, separate from
// the general attachments list (V-010). The handler owns writing the file to
// storage; this only persists the resulting URL. Returns the previous URL
// (possibly empty, if none was set) so the handler can best-effort delete the
// file it replaces, the same way ClearHeaderImage already does for removal.
func (uc *EventUsecase) SetHeaderImage(ctx context.Context, actorID uuid.UUID, eventID uuid.UUID, url string) (ev *domain.Event, prevURL string, err error) {
	e, err := uc.events.GetByID(ctx, eventID)
	if err != nil {
		return nil, "", err
	}
	prevURL = e.HeaderImageURL
	e.HeaderImageURL = url
	e.UpdatedAt = time.Now().UTC()
	if err := uc.events.Update(ctx, e); err != nil {
		return nil, "", fmt.Errorf("update event: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityEvent, eventID.String(), map[string]string{"headerImageUrl": prevURL}, map[string]string{"headerImageUrl": url})

	return e, prevURL, nil
}

// ClearHeaderImage removes an event's header image (V-010). Returns the
// previous URL (possibly empty, if none was set) so the handler can
// best-effort delete the underlying file.
func (uc *EventUsecase) ClearHeaderImage(ctx context.Context, actorID uuid.UUID, eventID uuid.UUID) (prevURL string, err error) {
	e, err := uc.events.GetByID(ctx, eventID)
	if err != nil {
		return "", err
	}
	prevURL = e.HeaderImageURL
	e.HeaderImageURL = ""
	e.UpdatedAt = time.Now().UTC()
	if err := uc.events.Update(ctx, e); err != nil {
		return "", fmt.Errorf("update event: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityEvent, eventID.String(), nil, map[string]string{"headerImageCleared": prevURL})

	return prevURL, nil
}

// ListAttachments returns the files attached to an event (V-008).
func (uc *EventUsecase) ListAttachments(ctx context.Context, eventID uuid.UUID) ([]*domain.EventAttachment, error) {
	if _, err := uc.events.GetByID(ctx, eventID); err != nil {
		return nil, err
	}
	return uc.attachments.ListByEvent(ctx, eventID)
}

// DeleteAttachment removes an attachment, scoped to the given event so a
// caller can't delete another event's attachment by guessing its ID. Returns
// the deleted attachment so the handler can also remove the underlying file.
func (uc *EventUsecase) DeleteAttachment(ctx context.Context, actorID uuid.UUID, eventID, attachmentID uuid.UUID) (*domain.EventAttachment, error) {
	a, err := uc.attachments.GetByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}
	if a.EventID != eventID {
		return nil, domain.ErrEventAttachmentNotFound
	}
	if err := uc.attachments.Delete(ctx, attachmentID); err != nil {
		return nil, err
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDelete, domain.AuditEntityEvent, eventID.String(), a, nil)

	return a, nil
}

func (uc *EventUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}
