package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// SettingsUsecase handles application configuration and audit log queries.
type SettingsUsecase struct {
	settings port.SettingsRepository
	audit    port.AuditRepository
	cache    port.CacheService
}

// NewSettingsUsecase creates a new SettingsUsecase.
func NewSettingsUsecase(settings port.SettingsRepository, audit port.AuditRepository, cache port.CacheService) *SettingsUsecase {
	return &SettingsUsecase{settings: settings, audit: audit, cache: cache}
}

// GetSettings loads and returns all application settings.
func (uc *SettingsUsecase) GetSettings(ctx context.Context) (*domain.AppSettings, error) {
	raw, err := uc.settings.GetAllSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	s := domain.DefaultAppSettings()

	if v, ok := raw[domain.SettingKeyNameMode]; ok {
		s.NameMode = domain.NameMode(v)
	}
	if v, ok := raw[domain.SettingKeyKioskSearch]; ok {
		s.KioskSearch = v == "true"
	}
	if v, ok := raw[domain.SettingKeyReservationHours]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			s.ReservationHours = i
		}
	}
	if v, ok := raw[domain.SettingKeyDeregisterDeadlineH]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			s.DeregisterDeadlineH = i
		}
	}
	if v, ok := raw[domain.SettingKeyBillingMode]; ok {
		s.BillingMode = domain.BillingMode(v)
	}
	if v, ok := raw[domain.SettingKeyReminderHourOfDay]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			s.ReminderHourOfDay = i
		}
	}
	if v, ok := raw[domain.SettingKeyReminderLeadWeeks]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			s.ReminderLeadWeeks = i
		}
	}
	if v, ok := raw[domain.SettingKeyBillingWarningLeadWeeks]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			s.BillingWarningLeadWeeks = i
		}
	}
	if v, ok := raw[domain.SettingKeyKioskLocked]; ok {
		s.KioskLocked = v == "true"
	}

	return &s, nil
}

// UpdateSettings persists changed application settings.
func (uc *SettingsUsecase) UpdateSettings(ctx context.Context, actorID uuid.UUID, input domain.AppSettings) (*domain.AppSettings, error) {
	kv := map[string]string{
		domain.SettingKeyNameMode:                string(input.NameMode),
		domain.SettingKeyKioskSearch:             strconv.FormatBool(input.KioskSearch),
		domain.SettingKeyReservationHours:        strconv.Itoa(input.ReservationHours),
		domain.SettingKeyDeregisterDeadlineH:     strconv.Itoa(input.DeregisterDeadlineH),
		domain.SettingKeyBillingMode:             string(input.BillingMode),
		domain.SettingKeyReminderHourOfDay:       strconv.Itoa(input.ReminderHourOfDay),
		domain.SettingKeyReminderLeadWeeks:       strconv.Itoa(input.ReminderLeadWeeks),
		domain.SettingKeyBillingWarningLeadWeeks: strconv.Itoa(input.BillingWarningLeadWeeks),
		domain.SettingKeyKioskLocked:             strconv.FormatBool(input.KioskLocked),
	}

	for k, v := range kv {
		if err := uc.settings.SetSetting(ctx, k, v); err != nil {
			return nil, fmt.Errorf("set setting %s: %w", k, err)
		}
	}

	// Invalidate cache.
	if uc.cache != nil {
		_ = uc.cache.Delete(ctx, "settings")
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntitySettings, "app_settings", nil, input)

	return uc.GetSettings(ctx)
}

// GetBranding returns the current branding configuration.
func (uc *SettingsUsecase) GetBranding(ctx context.Context) (*domain.BrandingConfig, error) {
	b, err := uc.settings.GetBranding(ctx)
	if err != nil {
		return nil, fmt.Errorf("load branding: %w", err)
	}
	return b, nil
}

// BrandingUpdateResult is the response to a branding update; it includes any
// WCAG contrast warnings (B-003) without rejecting the update.
type BrandingUpdateResult struct {
	Branding *domain.BrandingConfig `json:"branding"`
	Warnings []string               `json:"warnings,omitempty"`
}

// UpdateBranding persists a new branding configuration and returns any WCAG
// contrast warnings. Low-contrast colours produce a warning but are not rejected.
func (uc *SettingsUsecase) UpdateBranding(ctx context.Context, actorID uuid.UUID, input domain.BrandingConfig) (*BrandingUpdateResult, error) {
	if err := uc.settings.UpdateBranding(ctx, &input); err != nil {
		return nil, fmt.Errorf("update branding: %w", err)
	}

	// Invalidate cache.
	if uc.cache != nil {
		_ = uc.cache.Delete(ctx, "branding")
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityBranding, "branding_config", nil, input)

	return &BrandingUpdateResult{Branding: &input, Warnings: brandingContrastWarnings(input)}, nil
}

// brandingContrastWarnings returns human-readable WCAG AA contrast warnings for
// the primary and accent colours (B-003). The colours are checked against a
// white background (the dominant UI surface): a colour used as text/foreground
// on white needs at least a 4.5:1 ratio. Warnings are advisory only; the update
// is never rejected.
func brandingContrastWarnings(b domain.BrandingConfig) []string {
	var warnings []string
	check := func(label, hex string) {
		if hex == "" {
			return
		}
		if !domain.MeetsAAContrast(hex, "#FFFFFF") {
			ratio := domain.ContrastRatio(hex, "#FFFFFF")
			warnings = append(warnings, fmt.Sprintf(
				"%s color %s has insufficient contrast (%.2f:1, < 4.5:1) against a white background", label, hex, ratio))
		}
	}
	check("primary", b.PrimaryColor)
	check("accent", b.AccentColor)
	return warnings
}

// SetLogoURL updates only the branding logo URL (used by the logo upload, B-004).
func (uc *SettingsUsecase) SetLogoURL(ctx context.Context, actorID uuid.UUID, logoURL string) (*domain.BrandingConfig, error) {
	b, err := uc.settings.GetBranding(ctx)
	if err != nil {
		return nil, fmt.Errorf("load branding: %w", err)
	}
	b.LogoURL = logoURL
	if err := uc.settings.UpdateBranding(ctx, b); err != nil {
		return nil, fmt.Errorf("update branding: %w", err)
	}
	if uc.cache != nil {
		_ = uc.cache.Delete(ctx, "branding")
	}
	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityLogo, "branding_config", nil, map[string]string{"logoUrl": logoURL})
	return b, nil
}

// IsKioskLocked reports whether the public kiosk is currently locked (K-012).
func (uc *SettingsUsecase) IsKioskLocked(ctx context.Context) bool {
	v, err := uc.settings.GetSetting(ctx, domain.SettingKeyKioskLocked)
	if err != nil {
		return false
	}
	return v == "true"
}

// GetMemberFeeTiers returns the per-member fee tier override list (G-004).
func (uc *SettingsUsecase) GetMemberFeeTiers(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	return uc.settings.GetMemberFeeTiers(ctx, memberID, clubYearID)
}

// UpdateMemberFeeTiers replaces the per-member fee tier overrides (G-004).
func (uc *SettingsUsecase) UpdateMemberFeeTiers(ctx context.Context, actorID uuid.UUID, memberID, clubYearID uuid.UUID, tiers []*domain.FeeTier) ([]*domain.FeeTier, error) {
	for i, t := range tiers {
		if t.ID == uuid.Nil {
			t.ID = uuid.New()
		}
		t.ClubYearID = clubYearID
		t.Position = i + 1
	}
	if err := uc.settings.ReplaceMemberFeeTiers(ctx, memberID, clubYearID, tiers); err != nil {
		return nil, fmt.Errorf("replace member fee tiers: %w", err)
	}
	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityFeeTier, memberID.String(), nil, tiers)
	return tiers, nil
}

// GetFeeTiers returns the fee tiers for a specific club year.
func (uc *SettingsUsecase) GetFeeTiers(ctx context.Context, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	return uc.settings.GetFeeTiers(ctx, clubYearID)
}

// UpdateFeeTiers replaces all fee tiers for a specific club year.
func (uc *SettingsUsecase) UpdateFeeTiers(ctx context.Context, actorID uuid.UUID, clubYearID uuid.UUID, tiers []*domain.FeeTier) ([]*domain.FeeTier, error) {
	// Assign IDs and club year to new tiers.
	for i, t := range tiers {
		if t.ID == uuid.Nil {
			t.ID = uuid.New()
		}
		t.ClubYearID = clubYearID
		t.Position = i + 1
	}

	if err := uc.settings.ReplaceFeeTiers(ctx, clubYearID, tiers); err != nil {
		return nil, fmt.Errorf("replace fee tiers: %w", err)
	}

	// Invalidate cache.
	if uc.cache != nil {
		_ = uc.cache.Delete(ctx, fmt.Sprintf("fee_tiers:%s", clubYearID))
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityFeeTier, clubYearID.String(), nil, tiers)

	return tiers, nil
}

// GetAuditLog returns audit log entries with optional filtering.
func (uc *SettingsUsecase) GetAuditLog(ctx context.Context, filter port.AuditFilter) ([]*domain.AuditEntry, error) {
	return uc.audit.List(ctx, filter)
}

func (uc *SettingsUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}
