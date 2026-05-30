package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

func newSettingsUC() (*SettingsUsecase, *fakeSettingsRepo, *fakeAuditRepo, *fakeCache) {
	settings := newFakeSettingsRepo()
	audit := newFakeAuditRepo()
	cache := newFakeCache()
	return NewSettingsUsecase(settings, audit, cache), settings, audit, cache
}

func TestGetSettings_Defaults(t *testing.T) {
	uc, _, _, _ := newSettingsUC()
	s, err := uc.GetSettings(context.Background())
	require.NoError(t, err)
	assert.Equal(t, domain.NameModeAbbrev, s.NameMode)
	assert.Equal(t, 48, s.ReservationHours)
}

func TestUpdateSettings_WritesAuditAndPersists(t *testing.T) {
	uc, settings, audit, cache := newSettingsUC()
	_ = cache.Set(context.Background(), "settings", "stale", 0)

	in := domain.AppSettings{
		NameMode:            domain.NameModeFull,
		KioskSearch:         true,
		ReservationHours:    24,
		DeregisterDeadlineH: 12,
		BillingMode:         domain.BillingModeAuto,
	}
	out, err := uc.UpdateSettings(context.Background(), uuid.New(), in)
	require.NoError(t, err)
	assert.Equal(t, domain.NameModeFull, out.NameMode)
	assert.Equal(t, 24, out.ReservationHours)
	assert.True(t, out.KioskSearch)

	// Persisted in repo.
	v, _ := settings.GetSetting(context.Background(), domain.SettingKeyReservationHours)
	assert.Equal(t, "24", v)

	// Cache invalidated.
	_, cerr := cache.Get(context.Background(), "settings")
	assert.ErrorIs(t, cerr, port.ErrCacheMiss)

	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntitySettings))
}

func TestUpdateBranding_WritesAudit(t *testing.T) {
	uc, _, audit, _ := newSettingsUC()
	out, err := uc.UpdateBranding(context.Background(), uuid.New(), domain.BrandingConfig{ClubName: "New Club"})
	require.NoError(t, err)
	assert.Equal(t, "New Club", out.ClubName)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityBranding))
}

func TestUpdateFeeTiers_AssignsPositions(t *testing.T) {
	uc, settings, audit, _ := newSettingsUC()
	yearID := uuid.New()
	tiers := []*domain.FeeTier{
		{AmountCents: 300},
		{AmountCents: 500},
	}
	out, err := uc.UpdateFeeTiers(context.Background(), uuid.New(), yearID, tiers)
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, 1, out[0].Position)
	assert.Equal(t, 2, out[1].Position)
	assert.NotEqual(t, uuid.Nil, out[0].ID)
	assert.Equal(t, yearID, out[0].ClubYearID)

	stored, _ := settings.GetFeeTiers(context.Background(), yearID)
	assert.Len(t, stored, 2)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityFeeTier))
}

func TestGetAuditLog(t *testing.T) {
	uc, _, audit, _ := newSettingsUC()
	e, _ := domain.NewAuditEntry(nil, domain.AuditActionUpdate, domain.AuditEntitySettings, "app_settings", nil, nil)
	_ = audit.Insert(context.Background(), e)
	list, err := uc.GetAuditLog(context.Background(), port.AuditFilter{Entity: domain.AuditEntitySettings})
	require.NoError(t, err)
	assert.Len(t, list, 1)
}
