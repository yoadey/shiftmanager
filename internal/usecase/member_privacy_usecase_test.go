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

func newPrivacyUC() (*MemberPrivacyUsecase, *fakeMemberRepo, *fakeRegistrationRepo, *fakeHourRepo, *fakeAuditRepo) {
	members := newFakeMemberRepo()
	regs := newFakeRegistrationRepo()
	hours := newFakeHourRepo()
	audit := newFakeAuditRepo()
	return NewMemberPrivacyUsecase(members, regs, hours, audit), members, regs, hours, audit
}

func TestExportData(t *testing.T) {
	uc, members, regs, hours, audit := newPrivacyUC()
	mid := uuid.New()
	members.add(&domain.Member{ID: mid, FirstName: "Max", LastName: "Müller", Email: "max@x.de", IsActive: true})

	year := &domain.ClubYear{ID: uuid.New(), Label: "2026", IsActive: true}
	hours.addYear(year)
	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: mid, ClubYearID: year.ID, Hours: 5, Status: domain.HourEntryStatusConfirmed})
	_ = hours.UpsertHourTarget(context.Background(), &domain.HourTarget{ID: uuid.New(), MemberID: mid, ClubYearID: year.ID, TargetHours: 8})
	regs.add(&domain.Registration{ID: uuid.New(), ShiftID: uuid.New(), MemberID: &mid, State: domain.RegistrationStateConfirmed})

	export, err := uc.ExportData(context.Background(), uuid.New(), mid)
	require.NoError(t, err)
	assert.Equal(t, "max@x.de", export.Member.Email)
	assert.Len(t, export.Registrations, 1)
	assert.Len(t, export.HourEntries, 1)
	assert.Len(t, export.HourTargets, 1)
	assert.True(t, audit.has(domain.AuditActionGDPRExport, domain.AuditEntityMember))
}

func TestGDPRDelete_Anonymizes(t *testing.T) {
	uc, members, _, _, audit := newPrivacyUC()
	mid := uuid.New()
	sub := "oidc-sub"
	members.add(&domain.Member{ID: mid, FirstName: "Max", LastName: "Müller", Email: "max@x.de", IsActive: true, OIDCSubject: &sub})

	err := uc.GDPRDelete(context.Background(), uuid.New(), mid)
	require.NoError(t, err)

	m, _ := members.GetByID(context.Background(), mid)
	assert.Equal(t, "Mitglied", m.LastName)
	assert.NotEqual(t, "max@x.de", m.Email)
	assert.False(t, m.IsActive)
	assert.Nil(t, m.OIDCSubject)
	assert.NotNil(t, m.LeftAt)
	assert.True(t, audit.has(domain.AuditActionGDPRDelete, domain.AuditEntityMember))
}

func TestSetReminderOptOut(t *testing.T) {
	uc, members, _, _, audit := newPrivacyUC()
	mid := uuid.New()
	members.add(&domain.Member{ID: mid, IsActive: true})

	require.NoError(t, uc.SetReminderOptOut(context.Background(), mid, mid, true))
	m, _ := members.GetByID(context.Background(), mid)
	assert.True(t, m.ReminderOptOut)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityMember))
}

func TestGDPRDelete_NotFound(t *testing.T) {
	uc, _, _, _, _ := newPrivacyUC()
	err := uc.GDPRDelete(context.Background(), uuid.New(), uuid.New())
	assert.Error(t, err)
}

var _ = time.Now
