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
	assert.Equal(t, domain.RegistrationStateConfirmed, reg.State)
	assert.Nil(t, reg.ConfirmationToken)
	assert.True(t, f.audit.has(domain.AuditActionConfirm, domain.AuditEntityRegistration))

	// Token consumed -> invalid now.
	_, err = f.uc.ConfirmByLink(context.Background(), token)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
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
