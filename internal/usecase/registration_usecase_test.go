package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
)

type regFixture struct {
	regs     *fakeRegistrationRepo
	shifts   *fakeShiftRepo
	events   *fakeEventRepo
	members  *fakeMemberRepo
	email    *fakeEmailService
	audit    *fakeAuditRepo
	settings *fakeSettingsRepo
	uc       *RegistrationUsecase
	event    *domain.Event
	shift    *domain.Shift
}

func newRegFixture(maxHelpers int) *regFixture {
	f := &regFixture{
		regs:     newFakeRegistrationRepo(),
		shifts:   newFakeShiftRepo(),
		events:   newFakeEventRepo(),
		members:  newFakeMemberRepo(),
		email:    &fakeEmailService{},
		audit:    newFakeAuditRepo(),
		settings: newFakeSettingsRepo(),
	}
	f.event = &domain.Event{ID: uuid.New(), Status: domain.EventStatusPublished, Name: "Fest"}
	f.events.add(f.event)
	f.shift = &domain.Shift{
		ID:         uuid.New(),
		EventID:    f.event.ID,
		Name:       "Bar",
		StartAt:    time.Now().UTC().Add(72 * time.Hour),
		EndAt:      time.Now().UTC().Add(76 * time.Hour),
		MaxHelpers: maxHelpers,
	}
	f.shifts.add(f.shift)
	f.uc = NewRegistrationUsecase(f.regs, f.shifts, f.events, f.members, f.email, f.audit, f.settings)
	return f
}

func TestRegister_Self(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "a@b.de", FirstName: "A", LastName: "B", IsActive: true})

	reg, err := f.uc.Register(context.Background(), &memberID, RegisterInput{ShiftID: f.shift.ID, MemberID: &memberID})
	require.NoError(t, err)
	assert.Equal(t, domain.RegistrationStateRegistered, reg.State)
	assert.Nil(t, reg.ReservedUntil)
	assert.True(t, f.audit.has(domain.AuditActionRegister, domain.AuditEntityRegistration))
	assert.Equal(t, 1, f.email.countKind("confirmation")) // member registration sends standard confirmation
}

func TestRegister_KioskGuestEmail(t *testing.T) {
	f := newRegFixture(5)
	guest := "guest@example.com"
	reg, err := f.uc.Register(context.Background(), nil, RegisterInput{ShiftID: f.shift.ID, GuestEmail: &guest})
	require.NoError(t, err)
	assert.Equal(t, domain.RegistrationStateRegistered, reg.State)
	assert.NotNil(t, reg.ConfirmationToken)
	assert.Equal(t, 1, f.email.countKind("kiosk"))
}

func TestRegister_FullShiftBecomesReserved(t *testing.T) {
	f := newRegFixture(1)
	// Fill the single slot.
	existing := uuid.New()
	f.members.add(&domain.Member{ID: existing, Email: "x@y.de", IsActive: true})
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, MemberID: &existing, State: domain.RegistrationStateRegistered})

	// Set reservation hours setting.
	_ = f.settings.SetSetting(context.Background(), domain.SettingKeyReservationHours, "48")

	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "new@y.de", IsActive: true})
	reg, err := f.uc.Register(context.Background(), &memberID, RegisterInput{ShiftID: f.shift.ID, MemberID: &memberID})
	require.NoError(t, err)
	assert.Equal(t, domain.RegistrationStateReserved, reg.State)
	require.NotNil(t, reg.ReservedUntil)
	assert.WithinDuration(t, time.Now().UTC().Add(48*time.Hour), *reg.ReservedUntil, time.Minute)
}

func TestRegister_DuplicateMember(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "a@b.de", IsActive: true})
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, MemberID: &memberID, State: domain.RegistrationStateRegistered})

	_, err := f.uc.Register(context.Background(), &memberID, RegisterInput{ShiftID: f.shift.ID, MemberID: &memberID})
	assert.ErrorIs(t, err, domain.ErrAlreadyRegistered)
}

func TestRegister_EventNotPublished(t *testing.T) {
	f := newRegFixture(5)
	f.event.Status = domain.EventStatusDraft
	f.events.add(f.event)
	memberID := uuid.New()
	_, err := f.uc.Register(context.Background(), &memberID, RegisterInput{ShiftID: f.shift.ID, MemberID: &memberID})
	assert.Error(t, err)
}

func TestDeregister_BeforeDeadline(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "a@b.de", IsActive: true})
	regID := uuid.New()
	f.regs.add(&domain.Registration{ID: regID, ShiftID: f.shift.ID, MemberID: &memberID, State: domain.RegistrationStateRegistered})

	err := f.uc.Deregister(context.Background(), &memberID, regID, false)
	require.NoError(t, err)
	_, getErr := f.regs.GetByID(context.Background(), regID)
	assert.ErrorIs(t, getErr, domain.ErrRegistrationNotFound)
	assert.True(t, f.audit.has(domain.AuditActionDeregister, domain.AuditEntityRegistration))
	assert.Equal(t, 1, f.email.countKind("cancellation"))
}

func TestDeregister_DeadlinePassed(t *testing.T) {
	f := newRegFixture(5)
	// Shift starts in 1 hour, deadline is 24h before start -> already passed.
	f.shift.StartAt = time.Now().UTC().Add(time.Hour)
	f.shifts.add(f.shift)
	_ = f.settings.SetSetting(context.Background(), domain.SettingKeyDeregisterDeadlineH, "24")

	memberID := uuid.New()
	regID := uuid.New()
	f.regs.add(&domain.Registration{ID: regID, ShiftID: f.shift.ID, MemberID: &memberID, State: domain.RegistrationStateRegistered})

	err := f.uc.Deregister(context.Background(), &memberID, regID, false)
	assert.ErrorIs(t, err, domain.ErrDeregisterDeadlinePassed)

	// Force overrides the deadline.
	err = f.uc.Deregister(context.Background(), &memberID, regID, true)
	assert.NoError(t, err)
}

func TestExpireReservations(t *testing.T) {
	f := newRegFixture(5)
	past := time.Now().UTC().Add(-time.Hour)
	future := time.Now().UTC().Add(time.Hour)

	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, State: domain.RegistrationStateReserved, ReservedUntil: &past})
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, State: domain.RegistrationStateReserved, ReservedUntil: &future})
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, State: domain.RegistrationStateRegistered})

	n, err := f.uc.ExpireReservations(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, f.audit.countActions(domain.AuditActionDeregister))
}

func TestConfirmByLink(t *testing.T) {
	f := newRegFixture(5)
	token := uuid.New()
	regID := uuid.New()
	f.regs.add(&domain.Registration{ID: regID, ShiftID: f.shift.ID, State: domain.RegistrationStateRegistered, ConfirmationToken: &token})

	reg, err := f.uc.ConfirmByLink(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, domain.RegistrationStateRegistered, reg.State, "a link never confirms attendance (K-011)")
	assert.Nil(t, reg.ConfirmationToken)
	assert.True(t, f.audit.has(domain.AuditActionConfirm, domain.AuditEntityRegistration))

	// Token consumed -> invalid now.
	_, err = f.uc.ConfirmByLink(context.Background(), token)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestConfirmByLink_MakesReservationBinding(t *testing.T) {
	f := newRegFixture(5)
	token := uuid.New()
	future := time.Now().UTC().Add(time.Hour)
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, State: domain.RegistrationStateReserved, ReservedUntil: &future, ConfirmationToken: &token})

	reg, err := f.uc.ConfirmByLink(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, domain.RegistrationStateRegistered, reg.State)
	assert.Nil(t, reg.ReservedUntil, "a binding registration no longer expires (K-009)")
}

func TestConfirmByLink_ExpiredReservation(t *testing.T) {
	f := newRegFixture(5)
	token := uuid.New()
	past := time.Now().UTC().Add(-time.Hour)
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, State: domain.RegistrationStateReserved, ReservedUntil: &past, ConfirmationToken: &token})

	_, err := f.uc.ConfirmByLink(context.Background(), token)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestConfirmShiftRegistration(t *testing.T) {
	f := newRegFixture(5)
	actor := uuid.New()
	regID := uuid.New()
	f.regs.add(&domain.Registration{ID: regID, ShiftID: f.shift.ID, State: domain.RegistrationStateRegistered})

	reg, err := f.uc.ConfirmShiftRegistration(context.Background(), actor, regID)
	require.NoError(t, err)
	assert.Equal(t, domain.RegistrationStateConfirmed, reg.State)
	assert.True(t, f.audit.has(domain.AuditActionConfirm, domain.AuditEntityRegistration))
}

func TestOrganizerAddGuest(t *testing.T) {
	f := newRegFixture(1) // maxHelpers=1: bypassing capacity is the point
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: f.shift.ID, State: domain.RegistrationStateRegistered})
	actor := uuid.New()
	email := "guest@example.com"

	reg, err := f.uc.OrganizerAddGuest(context.Background(), actor, f.shift.ID, "  Jane Doe  ", &email)
	require.NoError(t, err)
	assert.Nil(t, reg.MemberID)
	require.NotNil(t, reg.GuestName)
	assert.Equal(t, "Jane Doe", *reg.GuestName, "name is trimmed")
	require.NotNil(t, reg.GuestEmail)
	assert.Equal(t, email, *reg.GuestEmail)
	assert.Equal(t, domain.RegistrationStateRegistered, reg.State)
	assert.True(t, f.audit.has(domain.AuditActionRegister, domain.AuditEntityRegistration))
}

// The guest email must be normalized the same way the kiosk self-service
// path does, since FindByGuestEmailAndShift's duplicate check only
// lowercases its query argument, not the stored column.
func TestOrganizerAddGuest_NormalizesEmail(t *testing.T) {
	f := newRegFixture(5)

	rawEmail := "  Jane@Example.COM  "
	reg, err := f.uc.OrganizerAddGuest(context.Background(), uuid.New(), f.shift.ID, "Jane Doe", &rawEmail)
	require.NoError(t, err)
	require.NotNil(t, reg.GuestEmail)
	assert.Equal(t, "jane@example.com", *reg.GuestEmail)
}

func TestOrganizerAddGuest_EmailOptional(t *testing.T) {
	f := newRegFixture(5)

	reg, err := f.uc.OrganizerAddGuest(context.Background(), uuid.New(), f.shift.ID, "Jane Doe", nil)
	require.NoError(t, err)
	assert.Nil(t, reg.GuestEmail)
	require.NotNil(t, reg.GuestName)
	assert.Equal(t, "Jane Doe", *reg.GuestName)
}

func TestOrganizerAddGuest_RequiresName(t *testing.T) {
	f := newRegFixture(5)

	_, err := f.uc.OrganizerAddGuest(context.Background(), uuid.New(), f.shift.ID, "   ", nil)
	assert.Error(t, err)
}

func TestOrganizerAddGuest_RejectsDuplicateEmail(t *testing.T) {
	f := newRegFixture(5)
	email := "jane@example.com"

	_, err := f.uc.OrganizerAddGuest(context.Background(), uuid.New(), f.shift.ID, "Jane Doe", &email)
	require.NoError(t, err)

	// Same email, different case/whitespace and a different name -- still
	// the same person as far as the shift is concerned.
	dup := "  Jane@Example.com  "
	_, err = f.uc.OrganizerAddGuest(context.Background(), uuid.New(), f.shift.ID, "J. Doe", &dup)
	assert.ErrorIs(t, err, domain.ErrAlreadyRegistered)
}

func TestOrganizerAddGuest_SameNameDifferentEmailAllowed(t *testing.T) {
	f := newRegFixture(5)
	email1 := "jane1@example.com"
	email2 := "jane2@example.com"

	_, err := f.uc.OrganizerAddGuest(context.Background(), uuid.New(), f.shift.ID, "Jane Doe", &email1)
	require.NoError(t, err)
	_, err = f.uc.OrganizerAddGuest(context.Background(), uuid.New(), f.shift.ID, "Jane Doe", &email2)
	assert.NoError(t, err, "two different people can share a name")
}

func TestOrganizerAddGuest_UnknownShift(t *testing.T) {
	f := newRegFixture(5)

	_, err := f.uc.OrganizerAddGuest(context.Background(), uuid.New(), uuid.New(), "Jane Doe", nil)
	assert.ErrorIs(t, err, domain.ErrShiftNotFound)
}

func TestKioskRegister_NameOnly(t *testing.T) {
	f := newRegFixture(5)
	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "  Maximilian   Mustermann "})
	require.NoError(t, err)
	assert.Nil(t, reg.MemberID)
	require.NotNil(t, reg.GuestName)
	assert.Equal(t, "Maximilian Mustermann", *reg.GuestName)
	assert.True(t, reg.SelfRegistered)
	assert.Nil(t, reg.ConfirmationToken, "nothing to confirm without an e-mail")
	assert.Empty(t, f.email.sent)

	_, err = f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "M"})
	assert.ErrorIs(t, err, ErrInvalidKioskName)
}

func TestKioskRegister_KnownEmailAsksMemberAndAssignsOnConfirm(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "max@example.org", FirstName: "Max", LastName: "Mustermann", IsActive: true})

	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Maxi", Email: " MAX@example.org "})
	require.NoError(t, err)
	assert.Nil(t, reg.MemberID)
	require.NotNil(t, reg.PendingMemberID)
	assert.Equal(t, memberID, *reg.PendingMemberID)
	assert.Equal(t, "Maxi", *reg.GuestName)
	assert.Nil(t, reg.GuestEmail)
	require.Len(t, f.email.sent, 1)
	assert.Equal(t, sentEmail{kind: "kiosk", to: "max@example.org"}, f.email.sent[0])

	// Confirming assigns the entry without marking attendance (K-006).
	confirmed, err := f.uc.ConfirmByLink(context.Background(), *reg.ConfirmationToken)
	require.NoError(t, err)
	require.NotNil(t, confirmed.MemberID)
	assert.Equal(t, memberID, *confirmed.MemberID)
	assert.Equal(t, domain.RegistrationStateRegistered, confirmed.State)
	assert.Nil(t, confirmed.GuestName)
	assert.Nil(t, confirmed.PendingMemberID)
	assert.False(t, confirmed.SelfRegistered)
	assert.Equal(t, 0, f.email.countKind("kiosk_assigned"), "they confirmed themselves")

	// The link works once (K-011).
	_, err = f.uc.ConfirmByLink(context.Background(), *reg.ConfirmationToken)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestKioskRegister_UnknownEmailStaysGuest(t *testing.T) {
	f := newRegFixture(5)
	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Hans Gast", Email: "hans@example.org"})
	require.NoError(t, err)
	assert.Nil(t, reg.PendingMemberID)
	assert.Equal(t, "hans@example.org", *reg.GuestEmail)
	assert.True(t, reg.SelfRegistered)
	assert.Equal(t, []sentEmail{{kind: "kiosk", to: "hans@example.org"}}, f.email.sent)
}

func TestAssignRegistration(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "anna@example.org", FirstName: "Anna", LastName: "Bauer", IsActive: true})
	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Anna Baur"})
	require.NoError(t, err)

	actor := uuid.New()
	assigned, err := f.uc.AssignRegistration(context.Background(), actor, reg.ID, memberID)
	require.NoError(t, err)
	assert.Equal(t, memberID, *assigned.MemberID)
	assert.Nil(t, assigned.GuestName)
	assert.Equal(t, []sentEmail{{kind: "kiosk_assigned", to: "anna@example.org"}}, f.email.sent)
	assert.True(t, f.audit.has(domain.AuditActionAssign, domain.AuditEntityRegistration))

	_, err = f.uc.AssignRegistration(context.Background(), actor, reg.ID, memberID)
	assert.ErrorIs(t, err, domain.ErrRegistrationAssigned)

	// A member can't end up twice on the same shift.
	other, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Anna B."})
	require.NoError(t, err)
	_, err = f.uc.AssignRegistration(context.Background(), actor, other.ID, memberID)
	assert.ErrorIs(t, err, domain.ErrAlreadyRegistered)

	// Inactive members can't be assigned.
	gone := uuid.New()
	f.members.add(&domain.Member{ID: gone, FirstName: "Alt", LastName: "Glied", IsActive: false})
	_, err = f.uc.AssignRegistration(context.Background(), actor, other.ID, gone)
	assert.ErrorIs(t, err, domain.ErrMemberNotFound)
}

func TestSuggestMembers(t *testing.T) {
	f := newRegFixture(5)
	maxi := uuid.New()
	pending := uuid.New()
	onShift := uuid.New()
	f.members.add(&domain.Member{ID: maxi, FirstName: "Maximilian", LastName: "Müller", IsActive: true})
	f.members.add(&domain.Member{ID: uuid.New(), FirstName: "Erika", LastName: "Schmidt", IsActive: true})
	f.members.add(&domain.Member{ID: uuid.New(), FirstName: "Max", LastName: "Muller", IsActive: false})
	f.members.add(&domain.Member{ID: pending, Email: "zz@example.org", FirstName: "Zora", LastName: "Zimmer", IsActive: true})
	f.members.add(&domain.Member{ID: onShift, FirstName: "Max", LastName: "Mueller", IsActive: true})
	_, err := f.uc.Register(context.Background(), &onShift, RegisterInput{ShiftID: f.shift.ID, MemberID: &onShift})
	require.NoError(t, err)

	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Max Muller"})
	require.NoError(t, err)
	got, err := f.uc.SuggestMembers(context.Background(), reg.ID)
	require.NoError(t, err)
	require.Len(t, got, 1, "near miss suggested; unlike names, inactive and already-registered members left out")
	assert.Equal(t, maxi, got[0].ID)
	assert.Equal(t, "Maximilian Müller", got[0].Name)

	// The member whose e-mail was entered comes first, however unlike the name.
	withMail, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Max Muller", Email: "zz@example.org"})
	require.NoError(t, err)
	got, err = f.uc.SuggestMembers(context.Background(), withMail.ID)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, pending, got[0].ID)
	assert.True(t, got[0].Pending)
}

func TestKioskRegister_PendingEntryAssignsEvenAfterOrganiserConfirmed(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "max@example.org", FirstName: "Max", LastName: "Mustermann", IsActive: true})
	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Maxi", Email: "max@example.org"})
	require.NoError(t, err)

	// The organiser confirms attendance before the member clicks the link.
	_, err = f.uc.ConfirmShiftRegistration(context.Background(), uuid.New(), reg.ID)
	require.NoError(t, err)

	confirmed, err := f.uc.ConfirmByLink(context.Background(), *reg.ConfirmationToken)
	require.NoError(t, err)
	require.NotNil(t, confirmed.MemberID, "the link still assigns the entry (K-006)")
	assert.Equal(t, memberID, *confirmed.MemberID)
	assert.Equal(t, domain.RegistrationStateConfirmed, confirmed.State)
}

func TestKioskRegister_PendingReservationBindingOnConfirm(t *testing.T) {
	f := newRegFixture(1)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "max@example.org", FirstName: "Max", LastName: "Mustermann", IsActive: true})
	_, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Erika Beispiel"})
	require.NoError(t, err)
	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Maxi", Email: "max@example.org"})
	require.NoError(t, err)
	require.Equal(t, domain.RegistrationStateReserved, reg.State, "shift is full")

	confirmed, err := f.uc.ConfirmByLink(context.Background(), *reg.ConfirmationToken)
	require.NoError(t, err)
	require.NotNil(t, confirmed.MemberID)
	assert.Equal(t, domain.RegistrationStateRegistered, confirmed.State)
	assert.Nil(t, confirmed.ReservedUntil, "the member confirmed, so the reservation must not expire (K-006)")
}

func TestKioskRegister_GuestLinkThenAssignCreditsNoHours(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "hans@verein.de", FirstName: "Hans", LastName: "Gast", IsActive: true})
	reg, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Hans Gast", Email: "hans.privat@example.org"})
	require.NoError(t, err)
	require.Nil(t, reg.PendingMemberID)

	_, err = f.uc.ConfirmByLink(context.Background(), *reg.ConfirmationToken)
	require.NoError(t, err)
	assigned, err := f.uc.AssignRegistration(context.Background(), uuid.New(), reg.ID, memberID)
	require.NoError(t, err)
	assert.Equal(t, domain.RegistrationStateRegistered, assigned.State, "attendance is confirmed by an organiser only (K-007)")
}

func TestKioskRegister_MemberEmailTwiceAndMemberOnShift(t *testing.T) {
	f := newRegFixture(5)
	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: "max@example.org", FirstName: "Max", LastName: "Mustermann", IsActive: true})
	_, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Maxi", Email: "max@example.org"})
	require.NoError(t, err)

	// A second pending entry for the same member adds nothing, and answers
	// like any other e-mail entry (K-005).
	dup, err := f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Max M", Email: "max@example.org"})
	assert.NoError(t, err)
	assert.Nil(t, dup)
	assert.Len(t, f.email.sent, 1, "no second confirmation mail")

	// An unknown address entered twice answers the same way, so repeating an
	// address can't tell a member already on the shift from a guest.
	_, err = f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Hans Gast", Email: "hans@example.org"})
	require.NoError(t, err)
	dup, err = f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Hans Gast", Email: "hans@example.org"})
	assert.NoError(t, err)
	assert.Nil(t, dup)
	assert.Equal(t, 2, f.email.countKind("kiosk"), "one mail each to the member and the guest, none for the repeat")

	// A member already on the shift: same neutral answer as an unknown address (K-005).
	other := uuid.New()
	f.members.add(&domain.Member{ID: other, Email: "anna@example.org", FirstName: "Anna", LastName: "Bauer", IsActive: true})
	_, err = f.uc.Register(context.Background(), &other, RegisterInput{ShiftID: f.shift.ID, MemberID: &other})
	require.NoError(t, err)
	_, err = f.uc.KioskRegister(context.Background(), KioskRegisterInput{ShiftID: f.shift.ID, Name: "Anna", Email: "anna@example.org"})
	assert.NoError(t, err)
}
