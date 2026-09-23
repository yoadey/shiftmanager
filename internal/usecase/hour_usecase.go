package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// HourUsecase handles hour tracking and confirmation business logic.
type HourUsecase struct {
	hours   port.HourRepository
	members port.MemberRepository
	shifts  port.ShiftRepository
	audit   port.AuditRepository
	email   port.EmailService
	events  port.EventRepository
}

// NewHourUsecase creates a new HourUsecase.
func NewHourUsecase(
	hours port.HourRepository,
	members port.MemberRepository,
	shifts port.ShiftRepository,
	audit port.AuditRepository,
	emailSvc port.EmailService,
	events port.EventRepository,
) *HourUsecase {
	return &HourUsecase{
		hours:   hours,
		members: members,
		shifts:  shifts,
		audit:   audit,
		email:   emailSvc,
		events:  events,
	}
}

// ConfirmShiftHours creates a confirmed hour entry for a member after their shift.
func (uc *HourUsecase) ConfirmShiftHours(ctx context.Context, actorID uuid.UUID, memberID, shiftID, clubYearID uuid.UUID, hours float64, description string) (*domain.HourEntry, error) {
	if hours <= 0 {
		return nil, fmt.Errorf("hours must be positive")
	}

	if _, err := uc.members.GetByID(ctx, memberID); err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}
	if _, err := uc.shifts.GetByID(ctx, shiftID); err != nil {
		return nil, fmt.Errorf("shift not found: %w", err)
	}
	if _, err := uc.hours.GetClubYearByID(ctx, clubYearID); err != nil {
		return nil, fmt.Errorf("club year not found: %w", err)
	}

	aid := actorID
	entry := &domain.HourEntry{
		ID:          uuid.New(),
		MemberID:    memberID,
		ShiftID:     &shiftID,
		ClubYearID:  clubYearID,
		Hours:       hours,
		Type:        domain.HourEntryTypeShift,
		Status:      domain.HourEntryStatusConfirmed,
		BookedBy:    &aid,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}

	if err := uc.hours.CreateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("create hour entry: %w", err)
	}

	_ = uc.writeAudit(ctx, &aid, domain.AuditActionConfirm, domain.AuditEntityHourEntry, entry.ID.String(), nil, entry)

	// Send hours-confirmed email (best-effort).
	if uc.email != nil {
		member, _ := uc.members.GetByID(ctx, memberID)
		shift, _ := uc.shifts.GetByID(ctx, shiftID)
		var event *domain.Event
		if shift != nil && uc.events != nil {
			event, _ = uc.events.GetByID(ctx, shift.EventID)
		}
		if member != nil && member.Email != "" {
			_ = uc.email.SendHoursConfirmed(ctx, member.Email, member, shift, event, hours)
		}
	}

	return entry, nil
}

// ManualBookingInput holds the data needed for a manual hour booking.
type ManualBookingInput struct {
	MemberID    uuid.UUID
	ClubYearID  uuid.UUID
	Hours       float64
	Description string
	Status      domain.HourEntryStatus
}

// ManualBooking creates a manual hour entry (board use only).
func (uc *HourUsecase) ManualBooking(ctx context.Context, actorID uuid.UUID, input ManualBookingInput) (*domain.HourEntry, error) {
	if input.Hours == 0 {
		return nil, fmt.Errorf("hours must be non-zero")
	}
	if input.Description == "" {
		return nil, fmt.Errorf("description is required for manual bookings")
	}

	if _, err := uc.members.GetByID(ctx, input.MemberID); err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}
	if _, err := uc.hours.GetClubYearByID(ctx, input.ClubYearID); err != nil {
		return nil, fmt.Errorf("club year not found: %w", err)
	}

	status := input.Status
	if status == "" {
		status = domain.HourEntryStatusConfirmed
	}

	aid := actorID
	entry := &domain.HourEntry{
		ID:          uuid.New(),
		MemberID:    input.MemberID,
		ClubYearID:  input.ClubYearID,
		Hours:       input.Hours,
		Type:        domain.HourEntryTypeManual,
		Status:      status,
		BookedBy:    &aid,
		Description: input.Description,
		CreatedAt:   time.Now().UTC(),
	}

	if err := uc.hours.CreateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("create manual hour entry: %w", err)
	}

	_ = uc.writeAudit(ctx, &aid, domain.AuditActionManualBook, domain.AuditEntityHourEntry, entry.ID.String(), nil, entry)

	return entry, nil
}

// CorrectEntry updates an existing hour entry (corrections by board).
func (uc *HourUsecase) CorrectEntry(ctx context.Context, actorID uuid.UUID, entryID uuid.UUID, hours float64, description string) (*domain.HourEntry, error) {
	entry, err := uc.hours.GetEntryByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	before := *entry
	if hours != 0 {
		entry.Hours = hours
	}
	if description != "" {
		entry.Description = description
	}

	if err := uc.hours.UpdateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("update hour entry: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCorrect, domain.AuditEntityHourEntry, entryID.String(), before, entry)

	return entry, nil
}

// DeleteEntry removes an hour entry.
func (uc *HourUsecase) DeleteEntry(ctx context.Context, actorID uuid.UUID, entryID uuid.UUID) error {
	entry, err := uc.hours.GetEntryByID(ctx, entryID)
	if err != nil {
		return err
	}

	if err := uc.hours.DeleteEntry(ctx, entryID); err != nil {
		return fmt.Errorf("delete hour entry: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDelete, domain.AuditEntityHourEntry, entryID.String(), entry, nil)

	return nil
}

// GetMemberAccount returns a summary of the member's hours for the given club year.
func (uc *HourUsecase) GetMemberAccount(ctx context.Context, memberID, clubYearID uuid.UUID) (*domain.MemberHourAccount, error) {
	year, err := uc.hours.GetClubYearByID(ctx, clubYearID)
	if err != nil {
		return nil, err
	}

	member, err := uc.members.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	targetHours := resolveTargetHours(ctx, uc.hours, member, year)

	entries, err := uc.hours.FindEntriesByMemberAndYear(ctx, memberID, clubYearID)
	if err != nil {
		return nil, fmt.Errorf("load hour entries: %w", err)
	}

	var confirmed, pending float64
	for _, e := range entries {
		switch e.Status {
		case domain.HourEntryStatusConfirmed:
			confirmed += e.Hours
		case domain.HourEntryStatusPending:
			pending += e.Hours
		}
	}

	missing := targetHours - confirmed
	if missing < 0 {
		missing = 0
	}

	return &domain.MemberHourAccount{
		MemberID:       memberID,
		ClubYearID:     clubYearID,
		TargetHours:    targetHours,
		ConfirmedHours: confirmed,
		PendingHours:   pending,
		MissingHours:   missing,
	}, nil
}

// GetActiveClubYear returns the currently active club year.
func (uc *HourUsecase) GetActiveClubYear(ctx context.Context) (*domain.ClubYear, error) {
	return uc.hours.GetActiveClubYear(ctx)
}

// ListClubYears returns all club years, newest first.
func (uc *HourUsecase) ListClubYears(ctx context.Context) ([]*domain.ClubYear, error) {
	return uc.hours.ListClubYears(ctx)
}

// CreateClubYearInput holds the fields for creating a new club year.
type CreateClubYearInput struct {
	Label              string    `json:"label"`
	StartDate          time.Time `json:"startDate"`
	EndDate            time.Time `json:"endDate"`
	DefaultTargetHours float64   `json:"defaultTargetHours"`
	SetActive          bool      `json:"setActive"`
	// CarryOverEnabled configures whether excess hours from THIS year (once
	// it's later superseded by a new active year) get carried over (S-006).
	CarryOverEnabled bool `json:"carryOverEnabled"`
}

// CreateClubYear creates a new club year and optionally marks it as active.
// If it's set active and the previously active year had carry-over enabled
// (S-006), each active member's excess confirmed hours from that year
// (confirmed beyond target) are credited to them in the new year.
func (uc *HourUsecase) CreateClubYear(ctx context.Context, actorID uuid.UUID, input CreateClubYearInput) (*domain.ClubYear, error) {
	var prevYear *domain.ClubYear
	var prevYearLookupErr error
	if input.SetActive {
		prevYear, prevYearLookupErr = uc.hours.GetActiveClubYear(ctx)
		if errors.Is(prevYearLookupErr, domain.ErrClubYearNotFound) {
			// Expected for the very first club year ever created; there's
			// simply nothing to carry over from, and nothing to report.
			prevYearLookupErr = nil
		}
	}

	year := &domain.ClubYear{
		ID:                 uuid.New(),
		Label:              input.Label,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		DefaultTargetHours: input.DefaultTargetHours,
		IsActive:           input.SetActive,
		CarryOverEnabled:   input.CarryOverEnabled,
	}
	// Creating it as *the* active year is a single atomic operation (see
	// CreateActiveClubYear) — deactivating every other year in the same
	// transaction is what keeps "the active club year" a single,
	// well-defined thing even under concurrent calls, without which
	// GetMemberAccountFull and the admin UI's "Aktiv" badge would both treat
	// every year ever marked active as still active.
	var err error
	if input.SetActive {
		err = uc.hours.CreateActiveClubYear(ctx, year)
	} else {
		err = uc.hours.CreateClubYear(ctx, year)
	}
	if err != nil {
		return nil, fmt.Errorf("create club year: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityClubYear, year.ID.String(), nil, year)

	if input.SetActive {
		if prevYearLookupErr != nil {
			// A real (non-not-found) failure reading the previous active
			// year: we can't know whether carry-over should have run, so
			// it's skipped below (prevYear is nil) — but that must not be
			// silent, since the year was still created and activated either way.
			_ = uc.writeAudit(ctx, &aid, domain.AuditActionCarryOver, domain.AuditEntityClubYear, year.ID.String(), nil,
				map[string]string{"error": "could not determine previous active club year, carry-over skipped: " + prevYearLookupErr.Error()})
		} else if prevYear != nil {
			before := *prevYear
			after := *prevYear
			after.IsActive = false
			_ = uc.writeAudit(ctx, &aid, domain.AuditActionDeactivate, domain.AuditEntityClubYear, prevYear.ID.String(), before, after)
		}

		if prevYear != nil && prevYear.CarryOverEnabled {
			uc.carryOverExcessHours(ctx, actorID, prevYear, year)
		}
	}

	return year, nil
}

// resolveTargetHours returns the effective hour target for a member in a
// given club year: an explicit per-member-per-year HourTarget overrides an
// individual goal override, which overrides the year's default.
func resolveTargetHours(ctx context.Context, hours port.HourRepository, member *domain.Member, year *domain.ClubYear) float64 {
	target := year.DefaultTargetHours
	if member.IndividualGoalHours != nil {
		target = *member.IndividualGoalHours
	}
	if ht, err := hours.GetHourTarget(ctx, member.ID, year.ID); err == nil && ht != nil {
		target = ht.TargetHours
	}
	return target
}

// carryOverExcessHours implements S-006: for each active member, any
// confirmed hours in `from` beyond their target for that year become an
// initial confirmed credit in `to`. Best-effort per member — the new club
// year already exists by the time this runs, so a failure here must not
// undo that; how many members were credited/failed is recorded on the audit
// entry so a board member can reconcile the rest manually via ManualBooking.
func (uc *HourUsecase) carryOverExcessHours(ctx context.Context, actorID uuid.UUID, from, to *domain.ClubYear) {
	active := true
	members, err := uc.members.List(ctx, port.MemberFilter{IsActive: &active})
	if err != nil {
		_ = uc.writeAudit(ctx, &actorID, domain.AuditActionCarryOver, domain.AuditEntityClubYear, to.ID.String(), nil,
			map[string]string{"error": "could not list members, carry-over skipped: " + err.Error()})
		return
	}

	entries, err := uc.hours.FindEntriesByYear(ctx, from.ID)
	if err != nil {
		_ = uc.writeAudit(ctx, &actorID, domain.AuditActionCarryOver, domain.AuditEntityClubYear, to.ID.String(), nil,
			map[string]string{"error": "could not load previous year's hour entries, carry-over skipped: " + err.Error()})
		return
	}
	confirmedByMember := make(map[uuid.UUID]float64)
	for _, e := range entries {
		if e.Status == domain.HourEntryStatusConfirmed {
			confirmedByMember[e.MemberID] += e.Hours
		}
	}

	var credited, failed int
	for _, m := range members {
		target := resolveTargetHours(ctx, uc.hours, m, from)
		excess := confirmedByMember[m.ID] - target
		if excess <= 0 {
			continue
		}

		aid := actorID
		entry := &domain.HourEntry{
			ID:          uuid.New(),
			MemberID:    m.ID,
			ClubYearID:  to.ID,
			Hours:       excess,
			Type:        domain.HourEntryTypeCarryOver,
			Status:      domain.HourEntryStatusConfirmed,
			BookedBy:    &aid,
			Description: fmt.Sprintf("Übertrag aus %s", from.Label),
			CreatedAt:   time.Now().UTC(),
		}
		if err := uc.hours.CreateEntry(ctx, entry); err != nil {
			failed++
			continue
		}
		credited++
	}

	_ = uc.writeAudit(ctx, &actorID, domain.AuditActionCarryOver, domain.AuditEntityClubYear, to.ID.String(), nil, map[string]any{
		"carryOverFromClubYearId": from.ID.String(),
		"membersCredited":         credited,
		"membersFailed":           failed,
	})
}

// MemberAccountFull combines the hour account summary with the raw entries.
type MemberAccountFull struct {
	Confirmed float64             `json:"confirmed"`
	Reserved  float64             `json:"reserved"`
	Goal      float64             `json:"goal"`
	Entries   []*domain.HourEntry `json:"entries"`
}

// GetMemberAccountFull returns the account summary and raw entries for a member
// using the active club year. Falls back gracefully when no active year exists.
func (uc *HourUsecase) GetMemberAccountFull(ctx context.Context, memberID uuid.UUID) (*MemberAccountFull, error) {
	year, err := uc.hours.GetActiveClubYear(ctx)
	if err != nil {
		return &MemberAccountFull{Entries: []*domain.HourEntry{}}, nil
	}
	account, err := uc.GetMemberAccount(ctx, memberID, year.ID)
	if err != nil {
		return &MemberAccountFull{Entries: []*domain.HourEntry{}}, nil
	}
	entries, err := uc.hours.FindEntriesByMemberAndYear(ctx, memberID, year.ID)
	if err != nil {
		entries = nil
	}
	if entries == nil {
		entries = []*domain.HourEntry{}
	}
	return &MemberAccountFull{
		Confirmed: account.ConfirmedHours,
		Reserved:  account.PendingHours,
		Goal:      account.TargetHours,
		Entries:   entries,
	}, nil
}

// GetYearSummary returns hour summaries for all active members for the given year.
func (uc *HourUsecase) GetYearSummary(ctx context.Context, clubYearID uuid.UUID) ([]domain.YearSummaryRow, error) {
	year, err := uc.hours.GetClubYearByID(ctx, clubYearID)
	if err != nil {
		return nil, err
	}

	active := true
	members, err := uc.members.List(ctx, port.MemberFilter{IsActive: &active})
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	allEntries, err := uc.hours.FindEntriesByYear(ctx, clubYearID)
	if err != nil {
		return nil, fmt.Errorf("load year entries: %w", err)
	}

	// Index entries by memberID.
	confirmedByMember := make(map[uuid.UUID]float64)
	for _, e := range allEntries {
		if e.Status == domain.HourEntryStatusConfirmed {
			confirmedByMember[e.MemberID] += e.Hours
		}
	}

	rows := make([]domain.YearSummaryRow, 0, len(members))
	for _, m := range members {
		target := resolveTargetHours(ctx, uc.hours, m, year)
		confirmed := confirmedByMember[m.ID]
		missing := target - confirmed
		if missing < 0 {
			missing = 0
		}

		rows = append(rows, domain.YearSummaryRow{
			Member:         *m,
			TargetHours:    target,
			ConfirmedHours: confirmed,
			MissingHours:   missing,
		})
	}

	return rows, nil
}

func (uc *HourUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}
