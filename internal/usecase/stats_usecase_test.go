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

func TestGetStats(t *testing.T) {
	hours := newFakeHourRepo()
	members := newFakeMemberRepo()
	shifts := newFakeShiftRepo()
	regs := newFakeRegistrationRepo()
	uc := NewStatsUsecase(hours, members, shifts, regs)

	now := time.Now().UTC()
	year := &domain.ClubYear{ID: uuid.New(), Label: "2026", DefaultTargetHours: 10, IsActive: true,
		StartDate: now.Add(-time.Hour), EndDate: now.Add(720 * time.Hour)}
	hours.addYear(year)

	// One member below target, one above.
	below := uuid.New()
	above := uuid.New()
	members.add(&domain.Member{ID: below, IsActive: true, Email: "b@x.de"})
	members.add(&domain.Member{ID: above, IsActive: true, Email: "a@x.de"})
	members.add(&domain.Member{ID: uuid.New(), IsActive: false, Email: "inactive@x.de"})

	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: below, ClubYearID: year.ID, Hours: 3, Status: domain.HourEntryStatusConfirmed})
	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: above, ClubYearID: year.ID, Hours: 12, Status: domain.HourEntryStatusConfirmed})

	// One understaffed upcoming shift (min 2, no registrations) and one fully staffed.
	understaffed := uuid.New()
	shifts.add(&domain.Shift{ID: understaffed, StartAt: now.Add(48 * time.Hour), EndAt: now.Add(50 * time.Hour), MinHelpers: 2})
	staffed := uuid.New()
	shifts.add(&domain.Shift{ID: staffed, StartAt: now.Add(24 * time.Hour), EndAt: now.Add(26 * time.Hour), MinHelpers: 1})
	regs.add(&domain.Registration{ID: uuid.New(), ShiftID: staffed, MemberID: &below, State: domain.RegistrationStateConfirmed})

	stats, err := uc.GetStats(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 2, stats.ActiveMembers)
	assert.Equal(t, 15.0, stats.TotalConfirmedHours)
	assert.Equal(t, 1, stats.MembersBelowTarget)
	assert.Equal(t, 2, stats.UpcomingShifts)
	assert.Equal(t, 1, stats.OpenShifts)
	assert.Equal(t, "2026", stats.ClubYearLabel)
}

func TestGetStats_NoActiveYear(t *testing.T) {
	uc := NewStatsUsecase(newFakeHourRepo(), newFakeMemberRepo(), newFakeShiftRepo(), newFakeRegistrationRepo())
	stats, err := uc.GetStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, stats.ActiveMembers)
}
