package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// HourUsecase handles hour tracking and confirmation business logic.
type HourUsecase struct {
	hours         port.HourRepository
	members       port.MemberRepository
	shifts        port.ShiftRepository
	registrations port.RegistrationRepository
	audit         port.AuditRepository
	email         port.EmailService
	events        port.EventRepository
}

// NewHourUsecase creates a new HourUsecase.
func NewHourUsecase(
	hours port.HourRepository,
	members port.MemberRepository,
	shifts port.ShiftRepository,
	registrations port.RegistrationRepository,
	audit port.AuditRepository,
	emailSvc port.EmailService,
	events port.EventRepository,
) *HourUsecase {
	return &HourUsecase{
		hours:         hours,
		members:       members,
		shifts:        shifts,
		registrations: registrations,
		audit:         audit,
		email:         emailSvc,
		events:        events,
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

	// Determine target hours (individual override takes priority).
	targetHours := year.DefaultTargetHours
	if member.IndividualGoalHours != nil {
		targetHours = *member.IndividualGoalHours
	}
	if ht, err := uc.hours.GetHourTarget(ctx, memberID, clubYearID); err == nil && ht != nil {
		targetHours = ht.TargetHours
	}

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
}

// CreateClubYear creates a new club year and optionally marks it as active.
// When SetActive is true, all other years are deactivated first.
func (uc *HourUsecase) CreateClubYear(ctx context.Context, actorID uuid.UUID, input CreateClubYearInput) (*domain.ClubYear, error) {
	if input.SetActive {
		if err := uc.hours.DeactivateAllClubYears(ctx); err != nil {
			return nil, fmt.Errorf("deactivate existing years: %w", err)
		}
	}
	year := &domain.ClubYear{
		ID:                 uuid.New(),
		Label:              input.Label,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		DefaultTargetHours: input.DefaultTargetHours,
		IsActive:           input.SetActive,
	}
	if err := uc.hours.CreateClubYear(ctx, year); err != nil {
		return nil, fmt.Errorf("create club year: %w", err)
	}
	return year, nil
}

// UpdateClubYearInput holds the editable fields for a club year.
type UpdateClubYearInput struct {
	Label              string    `json:"label"`
	StartDate          time.Time `json:"startDate"`
	EndDate            time.Time `json:"endDate"`
	DefaultTargetHours float64   `json:"defaultTargetHours"`
	SetActive          *bool     `json:"setActive"`
}

// UpdateClubYear updates label, date range, default target hours and active flag of an existing club year.
func (uc *HourUsecase) UpdateClubYear(ctx context.Context, actorID uuid.UUID, id uuid.UUID, input UpdateClubYearInput) (*domain.ClubYear, error) {
	year, err := uc.hours.GetClubYearByID(ctx, id)
	if err != nil {
		return nil, err
	}
	before := *year
	year.Label = input.Label
	year.StartDate = input.StartDate
	year.EndDate = input.EndDate
	if input.DefaultTargetHours > 0 {
		year.DefaultTargetHours = input.DefaultTargetHours
	}
	if input.SetActive != nil {
		if *input.SetActive {
			// Deactivate all others before activating this one.
			if err := uc.hours.DeactivateAllClubYears(ctx); err != nil {
				return nil, fmt.Errorf("deactivate existing years: %w", err)
			}
		}
		year.IsActive = *input.SetActive
	}
	if err := uc.hours.UpdateClubYear(ctx, year); err != nil {
		return nil, fmt.Errorf("update club year: %w", err)
	}
	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityClubYear, id.String(), before, *year)
	return year, nil
}

// DeleteClubYear removes a club year, including active ones.
func (uc *HourUsecase) DeleteClubYear(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	year, err := uc.hours.GetClubYearByID(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.hours.DeleteClubYear(ctx, id); err != nil {
		return fmt.Errorf("delete club year: %w", err)
	}
	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDelete, domain.AuditEntityClubYear, id.String(), year, nil)
	return nil
}

// MemberHourEntry is the per-entry view returned in MemberAccountFull.
// It extends the core HourEntry fields with optional display-only enrichment.
type MemberHourEntry struct {
	ID          uuid.UUID              `json:"id"`
	MemberID    uuid.UUID              `json:"memberId"`
	ShiftID     *uuid.UUID             `json:"shiftId,omitempty"`
	ClubYearID  uuid.UUID              `json:"clubYearId"`
	Hours       float64                `json:"hours"`
	Type        domain.HourEntryType   `json:"type"`
	Status      domain.HourEntryStatus `json:"status"`
	BookedBy    *uuid.UUID             `json:"bookedBy,omitempty"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"createdAt"`
	Manual      bool                   `json:"manual,omitempty"`
	Desc        string                 `json:"desc,omitempty"`
	Date        string                 `json:"date,omitempty"`
	EventName   string                 `json:"eventName,omitempty"`
	ShiftName   string                 `json:"shiftName,omitempty"`
}

// MemberAccountFull combines the hour account summary with enriched entries.
type MemberAccountFull struct {
	Confirmed      float64            `json:"confirmed"`
	Reserved       float64            `json:"reserved"`
	Goal           float64            `json:"goal"`
	ClubYearLabel  string             `json:"clubYearLabel"`
	Entries        []*MemberHourEntry `json:"entries"`
}

// GetMemberAccountFull returns the account summary and enriched entries for a member
// using the active club year. It counts both explicit HourEntry records and
// confirmed shift registrations so the total matches the billing computation.
func (uc *HourUsecase) GetMemberAccountFull(ctx context.Context, memberID uuid.UUID) (*MemberAccountFull, error) {
	empty := &MemberAccountFull{Entries: []*MemberHourEntry{}}

	year, err := uc.hours.GetActiveClubYear(ctx)
	if err != nil {
		return empty, nil
	}

	// Determine target hours for this member.
	targetHours := year.DefaultTargetHours
	if ht, err := uc.hours.GetHourTarget(ctx, memberID, year.ID); err == nil && ht != nil {
		targetHours = ht.TargetHours
	}

	// Collect explicit HourEntry records and track which shifts they cover.
	rawEntries, err := uc.hours.FindEntriesByMemberAndYear(ctx, memberID, year.ID)
	if err != nil {
		rawEntries = nil
	}

	type shiftKey = uuid.UUID
	coveredShifts := make(map[shiftKey]bool)
	var confirmed, pending float64
	result := make([]*MemberHourEntry, 0, len(rawEntries))

	for _, e := range rawEntries {
		switch e.Status {
		case domain.HourEntryStatusConfirmed:
			confirmed += e.Hours
			if e.ShiftID != nil {
				coveredShifts[*e.ShiftID] = true
			}
		case domain.HourEntryStatusPending:
			pending += e.Hours
		}
		me := &MemberHourEntry{
			ID:          e.ID,
			MemberID:    e.MemberID,
			ShiftID:     e.ShiftID,
			ClubYearID:  e.ClubYearID,
			Hours:       e.Hours,
			Type:        e.Type,
			Status:      e.Status,
			BookedBy:    e.BookedBy,
			Description: e.Description,
			CreatedAt:   e.CreatedAt,
			Manual:      e.Type == domain.HourEntryTypeManual,
			Desc:        e.Description,
			Date:        e.CreatedAt.UTC().Format("2006-01-02"),
		}
		result = append(result, me)
	}

	// Also count and display confirmed shift registrations not already covered by an entry.
	if uc.registrations != nil && uc.shifts != nil {
		memberRegs, regErr := uc.registrations.FindByMemberID(ctx, memberID)
		if regErr == nil {
			// Build a map of confirmed shift IDs from member's registrations.
			confirmedRegByShift := make(map[uuid.UUID]*domain.Registration)
			for _, reg := range memberRegs {
				if reg.State == domain.RegistrationStateConfirmed && reg.MemberID != nil {
					confirmedRegByShift[reg.ShiftID] = reg
				}
			}

			if len(confirmedRegByShift) > 0 {
				shiftsInYear, shErr := uc.shifts.FindShiftsStartingBetween(ctx, year.StartDate, year.EndDate.Add(time.Minute))
				if shErr == nil {
					for _, s := range shiftsInYear {
						if coveredShifts[s.ID] {
							continue
						}
						reg, ok := confirmedRegByShift[s.ID]
						if !ok {
							continue
						}
						hours := s.EndAt.Sub(s.StartAt).Hours()
						if hours <= 0 {
							continue
						}
						confirmed += hours

						eventName := ""
						if uc.events != nil {
							if ev, evErr := uc.events.GetByID(ctx, s.EventID); evErr == nil {
								eventName = ev.Name
							}
						}
						result = append(result, &MemberHourEntry{
							ID:          reg.ID,
							MemberID:    memberID,
							ShiftID:     &s.ID,
							ClubYearID:  year.ID,
							Hours:       hours,
							Type:        domain.HourEntryTypeShift,
							Status:      domain.HourEntryStatusConfirmed,
							Description: s.Name,
							CreatedAt:   s.StartAt,
							Date:        s.StartAt.UTC().Format("2006-01-02"),
							EventName:   eventName,
							ShiftName:   s.Name,
						})
					}
				}
			}
		}
	}

	return &MemberAccountFull{
		Confirmed:     confirmed,
		Reserved:      pending,
		Goal:          targetHours,
		ClubYearLabel: year.Label,
		Entries:       result,
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
		target := year.DefaultTargetHours
		if m.IndividualGoalHours != nil {
			target = *m.IndividualGoalHours
		}
		if ht, err := uc.hours.GetHourTarget(ctx, m.ID, clubYearID); err == nil && ht != nil {
			target = ht.TargetHours
		}

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
