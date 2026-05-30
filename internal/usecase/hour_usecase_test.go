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

func newHourUC() (*HourUsecase, *fakeHourRepo, *fakeMemberRepo, *fakeShiftRepo, *fakeAuditRepo) {
	hours := newFakeHourRepo()
	members := newFakeMemberRepo()
	shifts := newFakeShiftRepo()
	audit := newFakeAuditRepo()
	return NewHourUsecase(hours, members, shifts, audit), hours, members, shifts, audit
}

func seedYear(hours *fakeHourRepo, defaultTarget float64) *domain.ClubYear {
	y := &domain.ClubYear{ID: uuid.New(), Label: "2026", DefaultTargetHours: defaultTarget, IsActive: true,
		StartDate: time.Now().UTC().Add(-24 * time.Hour), EndDate: time.Now().UTC().Add(24 * time.Hour)}
	hours.addYear(y)
	return y
}

func TestConfirmShiftHours(t *testing.T) {
	uc, hours, members, shifts, audit := newHourUC()
	year := seedYear(hours, 20)
	memberID := uuid.New()
	members.add(&domain.Member{ID: memberID, Email: "a@b.de", IsActive: true})
	shiftID := uuid.New()
	shifts.add(&domain.Shift{ID: shiftID, EventID: uuid.New()})

	entry, err := uc.ConfirmShiftHours(context.Background(), uuid.New(), memberID, shiftID, year.ID, 4.5, "shift")
	require.NoError(t, err)
	assert.Equal(t, domain.HourEntryStatusConfirmed, entry.Status)
	assert.Equal(t, domain.HourEntryTypeShift, entry.Type)
	assert.Equal(t, 4.5, entry.Hours)
	assert.True(t, audit.has(domain.AuditActionConfirm, domain.AuditEntityHourEntry))
}

func TestConfirmShiftHours_InvalidHours(t *testing.T) {
	uc, hours, members, shifts, _ := newHourUC()
	year := seedYear(hours, 20)
	memberID := uuid.New()
	members.add(&domain.Member{ID: memberID, IsActive: true})
	shiftID := uuid.New()
	shifts.add(&domain.Shift{ID: shiftID})
	_, err := uc.ConfirmShiftHours(context.Background(), uuid.New(), memberID, shiftID, year.ID, 0, "x")
	assert.Error(t, err)
}

func TestManualBooking_WritesAudit(t *testing.T) {
	uc, hours, members, _, audit := newHourUC()
	year := seedYear(hours, 20)
	memberID := uuid.New()
	members.add(&domain.Member{ID: memberID, IsActive: true})

	entry, err := uc.ManualBooking(context.Background(), uuid.New(), ManualBookingInput{
		MemberID: memberID, ClubYearID: year.ID, Hours: 3, Description: "Aufbau",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.HourEntryTypeManual, entry.Type)
	assert.Equal(t, domain.HourEntryStatusConfirmed, entry.Status)
	assert.True(t, audit.has(domain.AuditActionManualBook, domain.AuditEntityHourEntry))
}

func TestManualBooking_Validation(t *testing.T) {
	uc, hours, members, _, _ := newHourUC()
	year := seedYear(hours, 20)
	memberID := uuid.New()
	members.add(&domain.Member{ID: memberID, IsActive: true})

	_, err := uc.ManualBooking(context.Background(), uuid.New(), ManualBookingInput{MemberID: memberID, ClubYearID: year.ID, Hours: 0, Description: "x"})
	assert.Error(t, err)
	_, err = uc.ManualBooking(context.Background(), uuid.New(), ManualBookingInput{MemberID: memberID, ClubYearID: year.ID, Hours: 3, Description: ""})
	assert.Error(t, err)
}

func TestGetMemberAccount_ConfirmedReservedGoal(t *testing.T) {
	uc, hours, members, _, _ := newHourUC()
	year := seedYear(hours, 20) // default target 20
	memberID := uuid.New()
	// individual goal overrides global default.
	goal := 15.0
	members.add(&domain.Member{ID: memberID, IsActive: true, IndividualGoalHours: &goal})

	// confirmed = 6 + 4 = 10, pending = 3
	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: memberID, ClubYearID: year.ID, Hours: 6, Status: domain.HourEntryStatusConfirmed})
	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: memberID, ClubYearID: year.ID, Hours: 4, Status: domain.HourEntryStatusConfirmed})
	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: memberID, ClubYearID: year.ID, Hours: 3, Status: domain.HourEntryStatusPending})

	acc, err := uc.GetMemberAccount(context.Background(), memberID, year.ID)
	require.NoError(t, err)
	assert.Equal(t, 15.0, acc.TargetHours) // individual goal precedence
	assert.Equal(t, 10.0, acc.ConfirmedHours)
	assert.Equal(t, 3.0, acc.PendingHours)
	assert.Equal(t, 5.0, acc.MissingHours)
}

func TestGetMemberAccount_HourTargetOverridesGoal(t *testing.T) {
	uc, hours, members, _, _ := newHourUC()
	year := seedYear(hours, 20)
	memberID := uuid.New()
	goal := 15.0
	members.add(&domain.Member{ID: memberID, IsActive: true, IndividualGoalHours: &goal})
	// Explicit hour target overrides both default and individual goal.
	_ = hours.UpsertHourTarget(context.Background(), &domain.HourTarget{ID: uuid.New(), MemberID: memberID, ClubYearID: year.ID, TargetHours: 8})

	acc, err := uc.GetMemberAccount(context.Background(), memberID, year.ID)
	require.NoError(t, err)
	assert.Equal(t, 8.0, acc.TargetHours)
}

func TestGetMemberAccount_DefaultTarget(t *testing.T) {
	uc, hours, members, _, _ := newHourUC()
	year := seedYear(hours, 25)
	memberID := uuid.New()
	members.add(&domain.Member{ID: memberID, IsActive: true})
	acc, err := uc.GetMemberAccount(context.Background(), memberID, year.ID)
	require.NoError(t, err)
	assert.Equal(t, 25.0, acc.TargetHours)
	assert.Equal(t, 25.0, acc.MissingHours)
}

func TestGetYearSummary(t *testing.T) {
	uc, hours, members, _, _ := newHourUC()
	year := seedYear(hours, 20)
	m1 := uuid.New()
	m2 := uuid.New()
	members.add(&domain.Member{ID: m1, IsActive: true})
	members.add(&domain.Member{ID: m2, IsActive: false}) // inactive, excluded

	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: m1, ClubYearID: year.ID, Hours: 12, Status: domain.HourEntryStatusConfirmed})

	rows, err := uc.GetYearSummary(context.Background(), year.ID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 12.0, rows[0].ConfirmedHours)
	assert.Equal(t, 8.0, rows[0].MissingHours)
}

func TestCorrectAndDeleteEntry(t *testing.T) {
	uc, hours, members, _, audit := newHourUC()
	year := seedYear(hours, 20)
	memberID := uuid.New()
	members.add(&domain.Member{ID: memberID, IsActive: true})
	entryID := uuid.New()
	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: entryID, MemberID: memberID, ClubYearID: year.ID, Hours: 5, Status: domain.HourEntryStatusConfirmed})

	corrected, err := uc.CorrectEntry(context.Background(), uuid.New(), entryID, 7, "fix")
	require.NoError(t, err)
	assert.Equal(t, 7.0, corrected.Hours)
	assert.True(t, audit.has(domain.AuditActionCorrect, domain.AuditEntityHourEntry))

	err = uc.DeleteEntry(context.Background(), uuid.New(), entryID)
	require.NoError(t, err)
	assert.True(t, audit.has(domain.AuditActionDelete, domain.AuditEntityHourEntry))
}
