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

func newNotificationUC(now time.Time, organizer string) (*NotificationUsecase, *fakeHourRepo, *fakeMemberRepo, *fakeShiftRepo, *fakeEventRepo, *fakeRegistrationRepo, *fakeSettingsRepo, *fakeEmailService) {
	hours := newFakeHourRepo()
	members := newFakeMemberRepo()
	shifts := newFakeShiftRepo()
	events := newFakeEventRepo()
	regs := newFakeRegistrationRepo()
	settings := newFakeSettingsRepo()
	email := &fakeEmailService{}
	uc := NewNotificationUsecase(hours, members, shifts, events, regs, settings, email, newFakeAuditRepo(), organizer)
	uc.now = func() time.Time { return now }
	return uc, hours, members, shifts, events, regs, settings, email
}

func TestRunYearEndMails_WarningPhase(t *testing.T) {
	now := time.Date(2026, 12, 15, 12, 0, 0, 0, time.UTC) // within 4 weeks of year end
	uc, hours, members, _, _, _, settings, email := newNotificationUC(now, "")
	year := &domain.ClubYear{ID: uuid.New(), Label: "2026", DefaultTargetHours: 10, IsActive: true,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC)}
	hours.addYear(year)
	_ = settings.SetSetting(context.Background(), domain.SettingKeyBillingWarningLeadWeeks, "4")

	mid := uuid.New()
	members.add(&domain.Member{ID: mid, IsActive: true, Email: "m@x.de"}) // 0 confirmed of 10 -> missing

	n, err := uc.RunDailyNotifications(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, email.countKind("missing_hours"))
}

func TestRunYearEndMails_BillingPhaseRequiresAuto(t *testing.T) {
	now := time.Date(2027, 1, 2, 12, 0, 0, 0, time.UTC) // after year end
	uc, hours, members, _, _, _, settings, email := newNotificationUC(now, "")
	year := &domain.ClubYear{ID: uuid.New(), Label: "2026", DefaultTargetHours: 10, IsActive: true,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC)}
	hours.addYear(year)
	members.add(&domain.Member{ID: uuid.New(), IsActive: true, Email: "m@x.de"})
	_ = settings.ReplaceFeeTiers(context.Background(), year.ID, []*domain.FeeTier{{ID: uuid.New(), ClubYearID: year.ID, Position: 1, AmountCents: 500}})

	// Manual mode: no billing mail.
	_ = settings.SetSetting(context.Background(), domain.SettingKeyBillingMode, string(domain.BillingModeManual))
	n, err := uc.RunDailyNotifications(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	// Auto mode: billing mail sent.
	_ = settings.SetSetting(context.Background(), domain.SettingKeyBillingMode, string(domain.BillingModeAuto))
	n, err = uc.RunDailyNotifications(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, email.countKind("year_billing"))
}

func TestNotifyUnderstaffed(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, _, _, shifts, events, _, _, email := newNotificationUC(now, "organizer@club.de")
	eid := uuid.New()
	events.add(&domain.Event{ID: eid, Name: "Fest"})
	shifts.add(&domain.Shift{ID: uuid.New(), EventID: eid, StartAt: now.Add(48 * time.Hour), EndAt: now.Add(50 * time.Hour), MinHelpers: 2})

	n, err := uc.RunDailyNotifications(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, n, 1)
	assert.Equal(t, 1, email.countKind("understaffed"))
}

func TestNotifyShiftUnderstaffed_Inline(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, _, _, shifts, events, _, _, email := newNotificationUC(now, "organizer@club.de")
	eid := uuid.New()
	events.add(&domain.Event{ID: eid, Name: "Fest"})
	sid := uuid.New()
	shifts.add(&domain.Shift{ID: sid, EventID: eid, StartAt: now.Add(48 * time.Hour), EndAt: now.Add(50 * time.Hour), MinHelpers: 1})

	require.NoError(t, uc.NotifyShiftUnderstaffed(context.Background(), sid))
	assert.Equal(t, 1, email.countKind("understaffed"))
}
