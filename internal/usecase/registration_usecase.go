package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	publicBaseURL string // origin of the app for e-mailed links (BASE_URL)
}

// SetPublicBaseURL sets the app origin used in e-mailed confirmation links.
func (uc *RegistrationUsecase) SetPublicBaseURL(url string) {
	uc.publicBaseURL = strings.TrimRight(url, "/")
}

// confirmURL is the app page that confirms a registration from its e-mailed
// token (it calls GET /api/v1/kiosk/confirm/{token}).
func (uc *RegistrationUsecase) confirmURL(token uuid.UUID) string {
	base := uc.publicBaseURL
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/shifts/confirm/%s", base, token)
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
	// GuestName is the name entered at the kiosk (K-014, K-006/K-007).
	GuestName *string
	// PendingMemberID is the member whose e-mail was entered at the kiosk;
	// the confirmation link goes to them (K-006).
	PendingMemberID *uuid.UUID
	// SelfRegistered marks a kiosk entry still to be assigned to a member.
	SelfRegistered bool
	Comment        string
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
	if input.PendingMemberID != nil {
		if existing, err := uc.registrations.FindByMemberAndShift(ctx, *input.PendingMemberID, input.ShiftID); err == nil && existing != nil {
			return nil, domain.ErrAlreadyRegistered
		}
		// The same member's e-mail entered twice: one pending entry is enough.
		if regs, err := uc.registrations.FindByShiftID(ctx, input.ShiftID); err == nil {
			for _, r := range regs {
				if r.PendingMemberID != nil && *r.PendingMemberID == *input.PendingMemberID {
					return nil, domain.ErrAlreadyRegistered
				}
			}
		}
	}

	token := uuid.New()
	reg := &domain.Registration{
		ID:                uuid.New(),
		ShiftID:           input.ShiftID,
		MemberID:          input.MemberID,
		GuestName:         input.GuestName,
		GuestEmail:        input.GuestEmail,
		State:             state,
		Comment:           input.Comment,
		ReservedUntil:     reservedUntil,
		ConfirmationToken: &token,
		CreatedAt:         now,
		SelfRegistered:    input.SelfRegistered,
		PendingMemberID:   input.PendingMemberID,
	}
	if input.MemberID == nil && input.GuestEmail == nil && input.PendingMemberID == nil {
		reg.ConfirmationToken = nil // name only: no link to send
	}

	if err := uc.registrations.Create(ctx, reg); err != nil {
		return nil, fmt.Errorf("create registration: %w", err)
	}

	// Send confirmation email.
	if uc.email != nil {
		switch {
		case input.PendingMemberID != nil:
			// Kiosk entry with a member's e-mail: ask that member to confirm (K-006).
			if m, err := uc.members.GetByID(ctx, *input.PendingMemberID); err == nil {
				_ = uc.email.SendKioskConfirmation(ctx, m.Email, uc.confirmURL(token), shift, event)
			}
		case input.GuestEmail != nil:
			// Kiosk guest: double-opt-in confirmation link (K-007).
			_ = uc.email.SendKioskConfirmation(ctx, *input.GuestEmail, uc.confirmURL(token), shift, event)
		case input.MemberID != nil:
			// Member flow: send standard registration confirmation.
			if m, err := uc.members.GetByID(ctx, *input.MemberID); err == nil {
				_ = uc.email.SendConfirmation(ctx, m.Email, reg, shift, event)
			}
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

// ConfirmByLink validates a one-time confirmation token. The link confirms
// the e-mail address (or, for a pending kiosk entry, the member's identity)
// and makes a reservation binding; it never confirms attendance, which only
// an organiser does (K-011).
func (uc *RegistrationUsecase) ConfirmByLink(ctx context.Context, token uuid.UUID) (*domain.Registration, error) {
	reg, err := uc.registrations.GetByToken(ctx, token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if reg.IsExpired(time.Now().UTC()) {
		return nil, domain.ErrInvalidToken
	}

	// A kiosk entry with a member's e-mail: confirming means "that was me";
	// it assigns the entry to that member (K-006).
	if reg.PendingMemberID != nil {
		return uc.assign(ctx, nil, reg, *reg.PendingMemberID, false, true)
	}

	before := *reg
	bindReservation(reg)
	reg.ConfirmationToken = nil // consume the token

	if err := uc.registrations.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("confirm registration: %w", err)
	}

	_ = uc.writeAudit(ctx, nil, domain.AuditActionConfirm, domain.AuditEntityRegistration, reg.ID.String(), before, reg)

	return reg, nil
}

// bindReservation turns a reservation into a binding registration that no
// longer expires (K-007, K-009); other states are left alone.
func bindReservation(reg *domain.Registration) {
	if reg.State == domain.RegistrationStateReserved {
		reg.State = domain.RegistrationStateRegistered
		reg.ReservedUntil = nil
	}
}

// KioskRegisterInput is a self-service kiosk registration: a member picked
// via the member search (MemberID), or a typed full name with an optional
// e-mail address (K-014, K-006, K-007).
type KioskRegisterInput struct {
	ShiftID  uuid.UUID
	MemberID *uuid.UUID
	Name     string
	Email    string
	Comment  string
}

// Bounds for a name typed at the kiosk.
const (
	minKioskNameLen = 2
	maxKioskNameLen = 100
)

// ErrInvalidKioskName is returned for a missing or overlong kiosk name.
var ErrInvalidKioskName = fmt.Errorf("name must be %d to %d characters", minKioskNameLen, maxKioskNameLen)

// KioskRegister registers someone from the kiosk. A typed name becomes an
// entry for an organiser to assign; with an e-mail that belongs to an active
// member, that member is asked to confirm instead. The result never reveals
// whether the e-mail was known (K-005).
func (uc *RegistrationUsecase) KioskRegister(ctx context.Context, in KioskRegisterInput) (*domain.Registration, error) {
	if in.MemberID != nil {
		return uc.Register(ctx, nil, RegisterInput{ShiftID: in.ShiftID, MemberID: in.MemberID, Comment: in.Comment})
	}
	name := strings.Join(strings.Fields(in.Name), " ")
	if n := len([]rune(name)); n < minKioskNameLen || n > maxKioskNameLen {
		return nil, ErrInvalidKioskName
	}
	input := RegisterInput{ShiftID: in.ShiftID, GuestName: &name, SelfRegistered: true, Comment: in.Comment}
	if email := strings.ToLower(strings.TrimSpace(in.Email)); email != "" {
		if m, err := uc.members.GetByEmail(ctx, email); err == nil && m.IsActive {
			input.PendingMemberID = &m.ID
		} else {
			input.GuestEmail = &email
		}
	}
	reg, err := uc.Register(ctx, nil, input)
	if errors.Is(err, domain.ErrAlreadyRegistered) {
		// Nothing to add. Every duplicate entered with an e-mail gets the
		// same answer as a new entry: whether the address's member is already
		// on the shift, already pending, or the address is an unknown guest's
		// must not be told apart, not even by entering it twice (K-005).
		return nil, nil
	}
	return reg, err
}

// AssignRegistration assigns an unassigned kiosk or guest entry to a member
// and tells the member by e-mail (K-015).
func (uc *RegistrationUsecase) AssignRegistration(ctx context.Context, actorID, regID, memberID uuid.UUID) (*domain.Registration, error) {
	reg, err := uc.registrations.GetByID(ctx, regID)
	if err != nil {
		return nil, err
	}
	aid := actorID
	return uc.assign(ctx, &aid, reg, memberID, true, false)
}

// assign links reg to memberID, clearing the kiosk/guest fields. notify sends
// the member an e-mail (not needed when they confirmed themselves); bind makes
// a reservation binding (only the member's own link confirmation does).
func (uc *RegistrationUsecase) assign(ctx context.Context, actorID *uuid.UUID, reg *domain.Registration, memberID uuid.UUID, notify, bind bool) (*domain.Registration, error) {
	if reg.MemberID != nil {
		return nil, domain.ErrRegistrationAssigned
	}
	member, err := uc.members.GetByID(ctx, memberID)
	if err != nil {
		return nil, domain.ErrMemberNotFound
	}
	if !member.IsActive {
		return nil, domain.ErrMemberNotFound
	}
	if existing, err := uc.registrations.FindByMemberAndShift(ctx, memberID, reg.ShiftID); err == nil && existing != nil && existing.ID != reg.ID {
		return nil, domain.ErrAlreadyRegistered
	}

	before := *reg
	reg.MemberID = &memberID
	reg.GuestName = nil
	reg.GuestEmail = nil
	reg.PendingMemberID = nil
	reg.SelfRegistered = false
	reg.ConfirmationToken = nil
	if bind {
		bindReservation(reg)
	}
	if err := uc.registrations.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("assign registration: %w", err)
	}
	_ = uc.writeAudit(ctx, actorID, domain.AuditActionAssign, domain.AuditEntityRegistration, reg.ID.String(), before, reg)

	if notify && uc.email != nil {
		if shift, err := uc.shifts.GetByID(ctx, reg.ShiftID); err == nil {
			event, _ := uc.events.GetByID(ctx, shift.EventID)
			_ = uc.email.SendKioskAssigned(ctx, member.Email, member, shift, event)
		}
	}
	return reg, nil
}

// Suggestion bounds for assigning kiosk entries (K-015).
const (
	suggestionMinScore = 0.5
	suggestionLimit    = 5
)

// SuggestMembers proposes active members for an unassigned entry by name
// similarity, including near misses; the member whose e-mail was entered
// comes first. Members already on the shift are left out.
func (uc *RegistrationUsecase) SuggestMembers(ctx context.Context, regID uuid.UUID) ([]domain.MemberSuggestion, error) {
	reg, err := uc.registrations.GetByID(ctx, regID)
	if err != nil {
		return nil, err
	}
	if reg.MemberID != nil {
		return nil, domain.ErrRegistrationAssigned
	}
	query := ""
	switch {
	case reg.GuestName != nil:
		query = *reg.GuestName
	case reg.GuestEmail != nil:
		// No name: the local part often holds one ("max.mustermann@...").
		local, _, _ := strings.Cut(*reg.GuestEmail, "@")
		query = strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(local)
	}

	onShift := map[uuid.UUID]bool{}
	if regs, err := uc.registrations.FindByShiftID(ctx, reg.ShiftID); err == nil {
		for _, r := range regs {
			if r.MemberID != nil {
				onShift[*r.MemberID] = true
			}
		}
	}
	members, err := uc.members.List(ctx, port.MemberFilter{Status: port.MemberStatusActive})
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	var out []domain.MemberSuggestion
	for _, m := range members {
		if onShift[m.ID] {
			continue
		}
		pending := reg.PendingMemberID != nil && *reg.PendingMemberID == m.ID
		score := domain.NameSimilarity(query, m.FullName())
		if score < suggestionMinScore && !pending {
			continue
		}
		out = append(out, domain.MemberSuggestion{ID: m.ID, Name: m.FullName(), Score: score, Pending: pending})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Pending != out[j].Pending {
			return out[i].Pending
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > suggestionLimit {
		out = out[:suggestionLimit]
	}
	return out, nil
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
