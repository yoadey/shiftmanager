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

type reminderFixture struct {
	shifts  *fakeShiftRepo
	regs    *fakeRegistrationRepo
	events  *fakeEventRepo
	members *fakeMemberRepo
	email   *fakeEmailService
	audit   *fakeAuditRepo
	uc      *ReminderUsecase
	now     time.Time
}

func newReminderFixture() *reminderFixture {
	f := &reminderFixture{
		shifts:  newFakeShiftRepo(),
		regs:    newFakeRegistrationRepo(),
		events:  newFakeEventRepo(),
		members: newFakeMemberRepo(),
		email:   &fakeEmailService{},
		audit:   newFakeAuditRepo(),
		now:     time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
	}
	f.uc = NewReminderUsecase(f.shifts, f.regs, f.events, f.members, f.email, f.audit)
	f.uc.now = func() time.Time { return f.now }
	return f
}

// addShiftWithReg creates an event, a shift starting at startAt, and registers a member.
func (f *reminderFixture) addShiftWithReg(startAt time.Time, state domain.RegistrationState, email string) uuid.UUID {
	eventID := uuid.New()
	f.events.add(&domain.Event{ID: eventID, Status: domain.EventStatusPublished, Name: "Fest"})
	shiftID := uuid.New()
	f.shifts.add(&domain.Shift{ID: shiftID, EventID: eventID, StartAt: startAt, EndAt: startAt.Add(2 * time.Hour)})

	memberID := uuid.New()
	f.members.add(&domain.Member{ID: memberID, Email: email, IsActive: true})
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: shiftID, MemberID: &memberID, State: state})
	return shiftID
}

func TestSendDueReminders_OneWeekAndOneDay(t *testing.T) {
	f := newReminderFixture()
	// One-week window: shift exactly 7 days out.
	f.addShiftWithReg(f.now.Add(7*24*time.Hour+10*time.Minute), domain.RegistrationStateRegistered, "week@b.de")
	// One-day window: shift ~24h out.
	f.addShiftWithReg(f.now.Add(24*time.Hour+10*time.Minute), domain.RegistrationStateConfirmed, "day@b.de")
	// Out of any window: shift 3 days out -> no reminder.
	f.addShiftWithReg(f.now.Add(3*24*time.Hour), domain.RegistrationStateRegistered, "noremind@b.de")

	n, err := f.uc.SendDueReminders(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, 2, f.email.countKind("reminder"))

	// Verify daysUntil values.
	days := map[int]bool{}
	for _, e := range f.email.sent {
		days[e.daysUntil] = true
	}
	assert.True(t, days[7])
	assert.True(t, days[1])
}

func TestSendDueReminders_GuestEmail(t *testing.T) {
	f := newReminderFixture()
	eventID := uuid.New()
	f.events.add(&domain.Event{ID: eventID, Status: domain.EventStatusPublished})
	shiftID := uuid.New()
	f.shifts.add(&domain.Shift{ID: shiftID, EventID: eventID, StartAt: f.now.Add(24 * time.Hour), EndAt: f.now.Add(26 * time.Hour)})
	guest := "guest@example.com"
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: shiftID, GuestEmail: &guest, State: domain.RegistrationStateRegistered})

	n, err := f.uc.SendDueReminders(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, "guest@example.com", f.email.sent[0].to)
}

func TestSendDueReminders_SkipsReservedAndNoShow(t *testing.T) {
	f := newReminderFixture()
	f.addShiftWithReg(f.now.Add(24*time.Hour+10*time.Minute), domain.RegistrationStateReserved, "reserved@b.de")
	f.addShiftWithReg(f.now.Add(24*time.Hour+20*time.Minute), domain.RegistrationStateNoShow, "noshow@b.de")

	n, err := f.uc.SendDueReminders(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestSendDueReminders_DeduplicatesPerEmail(t *testing.T) {
	f := newReminderFixture()
	eventID := uuid.New()
	f.events.add(&domain.Event{ID: eventID, Status: domain.EventStatusPublished})
	shiftID := uuid.New()
	f.shifts.add(&domain.Shift{ID: shiftID, EventID: eventID, StartAt: f.now.Add(24 * time.Hour), EndAt: f.now.Add(26 * time.Hour)})

	// Two registrations resolving to same guest email on the same shift.
	guest := "dup@example.com"
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: shiftID, GuestEmail: &guest, State: domain.RegistrationStateRegistered})
	f.regs.add(&domain.Registration{ID: uuid.New(), ShiftID: shiftID, GuestEmail: &guest, State: domain.RegistrationStateConfirmed})

	n, err := f.uc.SendDueReminders(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestSendDueReminders_WritesAudit(t *testing.T) {
	f := newReminderFixture()
	f.addShiftWithReg(f.now.Add(24*time.Hour+5*time.Minute), domain.RegistrationStateRegistered, "x@b.de")
	_, err := f.uc.SendDueReminders(context.Background())
	require.NoError(t, err)
	assert.True(t, f.audit.has(domain.AuditActionRegister, domain.AuditEntityRegistration))
}

func TestSendDueReminders_NoShifts(t *testing.T) {
	f := newReminderFixture()
	n, err := f.uc.SendDueReminders(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}
