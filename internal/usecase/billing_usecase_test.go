package usecase

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
)

func newBillingUC() (*BillingUsecase, *fakeHourRepo, *fakeMemberRepo, *fakeSettingsRepo, *fakeAuditRepo) {
	hours := newFakeHourRepo()
	members := newFakeMemberRepo()
	settings := newFakeSettingsRepo()
	audit := newFakeAuditRepo()
	return NewBillingUsecase(hours, members, settings, audit), hours, members, settings, audit
}

func seedBilling(t *testing.T) (*BillingUsecase, *fakeHourRepo, *fakeMemberRepo, *fakeSettingsRepo, *fakeAuditRepo, *domain.ClubYear, uuid.UUID) {
	t.Helper()
	uc, hours, members, settings, audit := newBillingUC()
	year := &domain.ClubYear{ID: uuid.New(), Label: "2026", DefaultTargetHours: 10, IsActive: true,
		StartDate: time.Now().UTC().Add(-time.Hour), EndDate: time.Now().UTC().Add(time.Hour)}
	hours.addYear(year)

	memberID := uuid.New()
	members.add(&domain.Member{ID: memberID, FirstName: "Max", LastName: "Müller", Email: "max@b.de", IsActive: true})

	// confirmed 6 of target 10 -> missing 4
	_ = hours.CreateEntry(context.Background(), &domain.HourEntry{ID: uuid.New(), MemberID: memberID, ClubYearID: year.ID, Hours: 6, Status: domain.HourEntryStatusConfirmed})

	// Two tiers: 300 then 500 (last applies to overflow).
	_ = settings.ReplaceFeeTiers(context.Background(), year.ID, []*domain.FeeTier{
		{ID: uuid.New(), ClubYearID: year.ID, Position: 1, AmountCents: 300},
		{ID: uuid.New(), ClubYearID: year.ID, Position: 2, AmountCents: 500},
	})
	return uc, hours, members, settings, audit, year, memberID
}

func TestComputeYearBilling(t *testing.T) {
	uc, _, _, _, audit, year, _ := seedBilling(t)
	report, err := uc.ComputeYearBilling(context.Background(), uuid.New(), year.ID)
	require.NoError(t, err)
	require.Len(t, report.Results, 1)
	r := report.Results[0]
	assert.Equal(t, 4.0, r.MissingHours)
	// tier1: 1h*300=300, tier2(last): 3h*500=1500 => 1800
	assert.Equal(t, 1800, r.TotalCents)
	assert.Equal(t, 1800, report.TotalCents)
	assert.True(t, audit.has(domain.AuditActionCompute, domain.AuditEntityClubYear))
}

func TestComputeYearBilling_NoFeeTiers(t *testing.T) {
	uc, hours, members, _, _ := newBillingUC()
	year := &domain.ClubYear{ID: uuid.New(), Label: "2026", DefaultTargetHours: 10}
	hours.addYear(year)
	members.add(&domain.Member{ID: uuid.New(), IsActive: true})
	_, err := uc.ComputeYearBilling(context.Background(), uuid.New(), year.ID)
	assert.ErrorIs(t, err, domain.ErrNoFeeTiers)
}

func TestExportBillingCSV(t *testing.T) {
	uc, _, _, _, audit, year, _ := seedBilling(t)
	data, err := uc.ExportBillingCSV(context.Background(), uuid.New(), year.ID)
	require.NoError(t, err)
	assert.Contains(t, string(data), "max@b.de")
	assert.Contains(t, string(data), "member_id")
	assert.True(t, audit.has(domain.AuditActionExport, domain.AuditEntityClubYear))
}

func TestExportBillingPDF(t *testing.T) {
	uc, _, _, _, _, year, _ := seedBilling(t)
	data, filename, err := uc.ExportBillingPDF(context.Background(), uuid.New(), year.ID)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(data, []byte("%PDF")), "should be a PDF")
	assert.Contains(t, filename, ".pdf")
	assert.Contains(t, filename, "2026")
}

func TestFormatCentsEUR(t *testing.T) {
	assert.Equal(t, "12.34 EUR", formatCentsEUR(1234))
	assert.Equal(t, "0.00 EUR", formatCentsEUR(0))
	assert.Equal(t, "5.00 EUR", formatCentsEUR(500))
}

func TestSanitizeFilename(t *testing.T) {
	assert.Equal(t, "2026-2027", sanitizeFilename("2026/2027"))
	assert.Equal(t, "report", sanitizeFilename(""))
	assert.Equal(t, "ClubYear-1", sanitizeFilename("ClubYear 1"))
}
