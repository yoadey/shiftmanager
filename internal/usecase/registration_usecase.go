package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// ShiftUnderstaffedNotifier is notified when a shift may have dropped below its
// minimum helper count (SC-007). Optional.
type ShiftUnderstaffedNotifier interface {
	NotifyShiftUnderstaffed(ctx context.Context, shiftID uuid.UUID) error
}

// RegistrationUsecase handles shift registration business logic.
type RegistrationUsecase struct {
	registrations port.RegistrationRepository
	shifts        port.ShiftRepository
	events        port.EventRepository
	members       port.MemberRepository
	email         port.EmailService
	audit         port.AuditRepository
	settings      port.SettingsRepository
	understaffed  ShiftUnderstaffedNotifier
}

// SetUnderstaffedNotifier wires the understaffed-shift notifier (SC-007). It is
// optional and set after construction to avoid a circular usecase dependency.
func (uc *RegistrationUsecase) SetUnderstaffedNotifier(n ShiftUnderstaffedNotifier) {
	uc.understaffed = n
}

// NewRegistrationUsecase creates a new RegistrationUsecase.
func NewRegistrationUsecase(
	registrations port.RegistrationRepository,
	shifts port.ShiftRepository,
	events port.EventRepository,
	members port.MemberRepository,
	email port.EmailService,
	audit port.AuditRepository,
	settings port.SettingsRepository,
) *RegistrationUsecase {
	return &RegistrationUsecase{
		registrations: registrations,
		shifts:        shifts,
		events:        events,
		members:       members,
		email:         email,
		audit:         audit,
		settings:      settings,
	}
}

// RegisterInput holds the data needed to register for a shift.
type RegisterInput struct {
	ShiftID    uuid.UUID
	MemberID   *uuid.UUID // nil for kiosk / guest registrations
	GuestEmail *string
	Comment    string
}

// Register creates a new registration for a shift.
// For kiosk registrations MemberID is nil and GuestEmail must be set.
func (uc *RegistrationUsecase) Register(ctx context.Context, actorID *uuid.UUID, input RegisterInput) (*domain.Registration, error) {
	shift, err := uc.shifts.GetByID(ctx, input.ShiftID)
	if err != nil {
		return nil, err
	}

	event, err := uc.events.GetByID(ctx, shift.EventID)
	if err != nil {
		return nil, err
	}

	if event.Status != domain.EventStatusPublished {
		return nil, fmt.Errorf("event is not published")
	}

	// Check capacity.
	count, err := uc.registrations.CountActiveByShift(ctx, input.ShiftID)
	if err != nil {
		return nil, fmt.Errorf("count registrations: %w", err)
	}

	now := time.Now().UTC()
	state := domain.RegistrationStateRegistered
	var reservedUntil *time.Time

	if shift.MaxHelpers > 0 && count >= shift.MaxHelpers {
		// Put on reserve list with expiry.
		resHours, err := getSettingInt(ctx, uc.settings, domain.SettingKeyReservationHours, 48)
		if err != nil {
			resHours = 48
		}
		expiry := now.Add(time.Duration(resHours) * time.Hour)
		state = domain.RegistrationStateReserved
		reservedUntil = &expiry
	}

	// Duplicate check.
	if input.MemberID != nil {
		if existing, err := uc.registrations.FindByMemberAndShift(ctx, *input.MemberID, input.ShiftID); err == nil && existing != nil {
			return nil, domain.ErrAlreadyRegistered
		}
	} else if input.GuestEmail != nil {
		if existing, err := uc.registrations.FindByGuestEmailAndShift(ctx, *input.GuestEmail, input.ShiftID); err == nil && existing != nil {
			return nil, domain.ErrAlreadyRegistered
		}
	}

	token := uuid.New()
	reg := &domain.Registration{
		ID:                uuid.New(),
		ShiftID:           input.ShiftID,
		MemberID:          input.MemberID,
		GuestEmail:        input.GuestEmail,
		State:             state,
		Comment:           input.Comment,
		ReservedUntil:     reservedUntil,
		ConfirmationToken: &token,
		CreatedAt:         now,
	}

	if err := uc.registrations.Create(ctx, reg); err != nil {
		return nil, fmt.Errorf("create registration: %w", err)
	}

	// Send confirmation email.
	var emailTo string
	if input.GuestEmail != nil {
		emailTo = *input.GuestEmail
	} else if input.MemberID != nil {
		if m, err := uc.members.GetByID(ctx, *input.MemberID); err == nil {
			emailTo = m.Email
		}
	}
	if emailTo != "" && uc.email != nil {
		if input.GuestEmail != nil {
			// Kiosk flow: send double-opt-in confirmation link.
			baseURL := "http://localhost:8080"
			confirmURL := fmt.Sprintf("%s/api/v1/kiosk/confirm/%s", baseURL, token.String())
			_ = uc.email.SendKioskConfirmation(ctx, emailTo, confirmURL, shift, event)
		} else if input.MemberID != nil {
			// Member flow: send standard registration confirmation.
			_ = uc.email.SendConfirmation(ctx, emailTo, reg, shift, event)
		}
	}

	_ = uc.writeAudit(ctx, actorID, domain.AuditActionRegister, domain.AuditEntityRegistration, reg.ID.String(), nil, reg)

	return reg, nil
}

// DeregisterByShift looks up the calling member's registration for the given shift
// and then delegates to Deregister. This avoids requiring the client to know the
// registration ID.
func (uc *RegistrationUsecase) DeregisterByShift(ctx context.Context, memberID uuid.UUID, shiftID uuid.UUID, force bool) error {
	reg, err := uc.registrations.FindByMemberAndShift(ctx, memberID, shiftID)
	if err != nil {
		return fmt.Errorf("keine Anmeldung für diese Schicht gefunden")
	}
	return uc.Deregister(ctx, &memberID, reg.ID, force)
}

// Deregister removes a registration if the deregistration deadline has not passed.
func (uc *RegistrationUsecase) Deregister(ctx context.Context, actorID *uuid.UUID, registrationID uuid.UUID, force bool) error {
	reg, err := uc.registrations.GetByID(ctx, registrationID)
	if err != nil {
		return err
	}

	shift, err := uc.shifts.GetByID(ctx, reg.ShiftID)
	if err != nil {
		return err
	}

	if !force && !shift.StartAt.IsZero() {
		deadlineH, err := getSettingInt(ctx, uc.settings, domain.SettingKeyDeregisterDeadlineH, 24)
		if err != nil {
			deadlineH = 24
		}
		deadline := shift.StartAt.Add(-time.Duration(deadlineH) * time.Hour)
		if time.Now().UTC().After(deadline) {
			return domain.ErrDeregisterDeadlinePassed
		}
	}

	if err := uc.registrations.Delete(ctx, registrationID); err != nil {
		return fmt.Errorf("delete registration: %w", err)
	}

	event, _ := uc.events.GetByID(ctx, shift.EventID)

	// Notify the registrant.
	var emailTo string
	if reg.GuestEmail != nil {
		emailTo = *reg.GuestEmail
	} else if reg.MemberID != nil {
		if m, err := uc.members.GetByID(ctx, *reg.MemberID); err == nil {
			emailTo = m.Email
		}
	}
	if emailTo != "" && uc.email != nil && event != nil {
		_ = uc.email.SendCancellation(ctx, emailTo, reg, shift, event)
	}

	_ = uc.writeAudit(ctx, actorID, domain.AuditActionDeregister, domain.AuditEntityRegistration, registrationID.String(), reg, nil)

	// SC-007: a deregistration may push the shift below its minimum; notify the
	// organizer. Best-effort and non-blocking on the caller's behalf.
	if uc.understaffed != nil {
		_ = uc.understaffed.NotifyShiftUnderstaffed(ctx, shift.ID)
	}

	return nil
}

// ConfirmByLink validates a one-time confirmation token and transitions the registration
// to confirmed state.
func (uc *RegistrationUsecase) ConfirmByLink(ctx context.Context, token uuid.UUID) (*domain.Registration, error) {
	reg, err := uc.registrations.GetByToken(ctx, token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if reg.State == domain.RegistrationStateConfirmed {
		return reg, nil // idempotent
	}

	if reg.IsExpired(time.Now().UTC()) {
		return nil, domain.ErrInvalidToken
	}

	reg.State = domain.RegistrationStateConfirmed
	reg.ConfirmationToken = nil // consume the token

	if err := uc.registrations.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("confirm registration: %w", err)
	}

	_ = uc.writeAudit(ctx, nil, domain.AuditActionConfirm, domain.AuditEntityRegistration, reg.ID.String(), nil, reg)

	return reg, nil
}

// ExpireReservations finds and deletes all reserved registrations that have passed
// their reservation deadline. Intended to be called by the background scheduler.
func (uc *RegistrationUsecase) ExpireReservations(ctx context.Context) (int, error) {
	expired, err := uc.registrations.ListUnconfirmedExpiredReservations(ctx, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("list expired reservations: %w", err)
	}

	count := 0
	for _, reg := range expired {
		if err := uc.registrations.Delete(ctx, reg.ID); err != nil {
			continue
		}
		count++
		_ = uc.writeAudit(ctx, nil, domain.AuditActionDeregister, domain.AuditEntityRegistration,
			reg.ID.String(), reg, map[string]string{"reason": "reservation_expired"})
	}

	return count, nil
}

// ConfirmShiftRegistration sets a registration to confirmed state (by veranstaltungsleiter).
func (uc *RegistrationUsecase) ConfirmShiftRegistration(ctx context.Context, actorID uuid.UUID, registrationID uuid.UUID) (*domain.Registration, error) {
	reg, err := uc.registrations.GetByID(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	before := *reg
	reg.State = domain.RegistrationStateConfirmed

	if err := uc.registrations.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("confirm registration: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionConfirm, domain.AuditEntityRegistration, reg.ID.String(), before, reg)

	return reg, nil
}

// UpdateRegistration changes the state and/or booked hours of a registration.
// Only the non-nil fields are applied. Intended for use by veranstaltungsleiter+.
func (uc *RegistrationUsecase) UpdateRegistration(ctx context.Context, actorID uuid.UUID, regID uuid.UUID, state *domain.RegistrationState, bookedHours *float64) (*domain.Registration, error) {
	reg, err := uc.registrations.GetByID(ctx, regID)
	if err != nil {
		return nil, err
	}
	before := *reg

	if state != nil {
		reg.State = *state
	}
	if bookedHours != nil {
		reg.BookedHours = bookedHours
	}

	if err := uc.registrations.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("update registration: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityRegistration, reg.ID.String(), before, reg)

	return reg, nil
}

// ForceDeleteRegistration removes a registration regardless of deadline or event status.
// Intended for use by veranstaltungsleiter+ to fix up attendance after the fact.
func (uc *RegistrationUsecase) ForceDeleteRegistration(ctx context.Context, actorID uuid.UUID, regID uuid.UUID) error {
	reg, err := uc.registrations.GetByID(ctx, regID)
	if err != nil {
		return err
	}

	if err := uc.registrations.Delete(ctx, regID); err != nil {
		return fmt.Errorf("delete registration: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDeregister, domain.AuditEntityRegistration, regID.String(), reg, nil)

	return nil
}

// OrganizerRegister adds a member to a shift on behalf of an organizer.
// Unlike the regular Register path it bypasses event-status and capacity checks
// so organisers can fix up attendance on completed events or override capacity.
func (uc *RegistrationUsecase) OrganizerRegister(ctx context.Context, actorID uuid.UUID, shiftID uuid.UUID, memberID uuid.UUID) (*domain.Registration, error) {
	if _, err := uc.shifts.GetByID(ctx, shiftID); err != nil {
		return nil, err
	}

	if _, err := uc.members.GetByID(ctx, memberID); err != nil {
		return nil, fmt.Errorf("member not found")
	}

	if _, err := uc.registrations.FindByMemberAndShift(ctx, memberID, shiftID); err == nil {
		return nil, domain.ErrAlreadyRegistered
	}

	reg := &domain.Registration{
		ID:        uuid.New(),
		ShiftID:   shiftID,
		MemberID:  &memberID,
		State:     domain.RegistrationStateRegistered,
		CreatedAt: time.Now().UTC(),
	}

	if err := uc.registrations.Create(ctx, reg); err != nil {
		return nil, fmt.Errorf("create registration: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionRegister, domain.AuditEntityRegistration, reg.ID.String(), nil, reg)

	return reg, nil
}

// OrganizerAddGuest adds a non-member helper to a shift by name (with an
// optional email), bypassing event-status and capacity checks like
// OrganizerRegister. Unlike the self-service kiosk guest flow (tracked by
// email only, via Register/GuestEmail), this covers a helper an organizer
// knows by name but who has no account and may not have an email at all.
func (uc *RegistrationUsecase) OrganizerAddGuest(ctx context.Context, actorID uuid.UUID, shiftID uuid.UUID, guestName string, guestEmail *string) (*domain.Registration, error) {
	guestName = strings.TrimSpace(guestName)
	if guestName == "" {
		return nil, fmt.Errorf("guest name is required")
	}
	// Normalized the same way as the kiosk self-service guest path
	// (kiosk_handler.go), since FindByGuestEmailAndShift's duplicate lookup
	// lowercases only its query argument: a mismatched case here would let
	// the same person end up registered twice for the same shift, once via
	// this path and once via kiosk self-service.
	if guestEmail != nil {
		e := strings.ToLower(strings.TrimSpace(*guestEmail))
		guestEmail = &e
	}

	if _, err := uc.shifts.GetByID(ctx, shiftID); err != nil {
		return nil, err
	}

	// Same duplicate guard as the kiosk self-service path and
	// OrganizerRegister's member path: without an email there's nothing
	// reliable to dedupe on (two different people can share a name), so this
	// only catches the case where one is given.
	if guestEmail != nil {
		if existing, err := uc.registrations.FindByGuestEmailAndShift(ctx, *guestEmail, shiftID); err == nil && existing != nil {
			return nil, domain.ErrAlreadyRegistered
		}
	}

	reg := &domain.Registration{
		ID:         uuid.New(),
		ShiftID:    shiftID,
		GuestName:  &guestName,
		GuestEmail: guestEmail,
		State:      domain.RegistrationStateRegistered,
		CreatedAt:  time.Now().UTC(),
	}

	if err := uc.registrations.Create(ctx, reg); err != nil {
		return nil, fmt.Errorf("create registration: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionRegister, domain.AuditEntityRegistration, reg.ID.String(), nil, reg)

	return reg, nil
}

func (uc *RegistrationUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}

// getSettingInt is a helper to read an integer setting with a fallback default.
func getSettingInt(ctx context.Context, repo port.SettingsRepository, key string, def int) (int, error) {
	val, err := repo.GetSetting(ctx, key)
	if err != nil {
		return def, err
	}
	var i int
	if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
		return def, nil
	}
	return i, nil
}
