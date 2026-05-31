package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
)

func TestUpdateSettings_NewKeysRoundTrip(t *testing.T) {
	uc, _, _, _ := newSettingsUC()
	in := domain.AppSettings{
		NameMode: domain.NameModeAbbrev, ReservationHours: 48, DeregisterDeadlineH: 24,
		BillingMode:       domain.BillingModeManual,
		ReminderHourOfDay: 9, ReminderLeadWeeks: 2, BillingWarningLeadWeeks: 6, KioskLocked: true,
	}
	out, err := uc.UpdateSettings(context.Background(), uuid.New(), in)
	require.NoError(t, err)
	assert.Equal(t, 9, out.ReminderHourOfDay)
	assert.Equal(t, 2, out.ReminderLeadWeeks)
	assert.Equal(t, 6, out.BillingWarningLeadWeeks)
	assert.True(t, out.KioskLocked)
}

func TestIsKioskLocked(t *testing.T) {
	uc, settings, _, _ := newSettingsUC()
	assert.False(t, uc.IsKioskLocked(context.Background()))
	_ = settings.SetSetting(context.Background(), domain.SettingKeyKioskLocked, "true")
	assert.True(t, uc.IsKioskLocked(context.Background()))
}

func TestUpdateBranding_ContrastWarnings(t *testing.T) {
	uc, _, _, _ := newSettingsUC()
	// Gold accent on a white background is low contrast -> warning for accent;
	// black primary on white is fine -> no warning for primary.
	res, err := uc.UpdateBranding(context.Background(), uuid.New(), domain.BrandingConfig{
		ClubName: "C", PrimaryColor: "#000000", AccentColor: "#F4B63F",
	})
	require.NoError(t, err)
	require.Len(t, res.Warnings, 1)
	assert.Contains(t, res.Warnings[0], "accent")
}

func TestUpdateBranding_NoWarningForGoodContrast(t *testing.T) {
	uc, _, _, _ := newSettingsUC()
	// Both dark colours have good contrast on white -> no warnings.
	res, err := uc.UpdateBranding(context.Background(), uuid.New(), domain.BrandingConfig{
		ClubName: "C", PrimaryColor: "#000000", AccentColor: "#0A3D62",
	})
	require.NoError(t, err)
	assert.Empty(t, res.Warnings)
}

func TestSetLogoURL(t *testing.T) {
	uc, _, audit, _ := newSettingsUC()
	b, err := uc.SetLogoURL(context.Background(), uuid.New(), "http://x/uploads/logo.png")
	require.NoError(t, err)
	assert.Equal(t, "http://x/uploads/logo.png", b.LogoURL)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityLogo))
}

func TestUpdateMemberFeeTiers(t *testing.T) {
	uc, settings, audit, _ := newSettingsUC()
	mid := uuid.New()
	yid := uuid.New()
	out, err := uc.UpdateMemberFeeTiers(context.Background(), uuid.New(), mid, yid, []*domain.FeeTier{{AmountCents: 800}})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, 1, out[0].Position)
	stored, _ := settings.GetMemberFeeTiers(context.Background(), mid, yid)
	assert.Len(t, stored, 1)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityFeeTier))
}
