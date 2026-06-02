package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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

	if v, ok := raw[domain.SettingKeyClubName]; ok && v != "" {
		s.ClubName = v
	}
	if v, ok := raw[domain.SettingKeyClubYear]; ok && v != "" {
		s.ClubYear = v
	}
	if v, ok := raw[domain.SettingKeyYearGoal]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			s.YearGoal = i
		}
	}
	if v, ok := raw[domain.SettingKeyFeeSchedule]; ok && v != "" {
		s.FeeSchedule = parseFeeSchedule(v)
	}
	if v, ok := raw[domain.SettingKeyNameMode]; ok && v != "" {
		s.NameMode = domain.NameMode(v)
	}
	if v, ok := raw[domain.SettingKeyKioskSearch]; ok {
		s.KioskSearch = v == "true"
	}
	if v, ok := raw[domain.SettingKeyReservationHours]; ok {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			s.ReservationHours = i
		}
	}
	if v, ok := raw[domain.SettingKeyDeregisterDeadlineH]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			s.DeregisterDeadlineH = i
		}
	}
	if v, ok := raw[domain.SettingKeyBillingMode]; ok && v != "" {
		s.BillingMode = domain.BillingMode(v)
	}
	if v, ok := raw[domain.SettingKeyReminderHourOfDay]; ok {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			s.ReminderHourOfDay = i
		}
	}
	if v, ok := raw[domain.SettingKeyReminderLeadWeeks]; ok {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			s.ReminderLeadWeeks = i
		}
	}
	if v, ok := raw[domain.SettingKeyBillingWarningLeadWeeks]; ok {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
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
	kv := map[string]string{}
	if input.ClubName != "" {
		kv[domain.SettingKeyClubName] = input.ClubName
	}
	if input.ClubYear != "" {
		kv[domain.SettingKeyClubYear] = input.ClubYear
	}
	if input.YearGoal > 0 {
		kv[domain.SettingKeyYearGoal] = strconv.Itoa(input.YearGoal)
	}
	if len(input.FeeSchedule) > 0 {
		kv[domain.SettingKeyFeeSchedule] = formatFeeSchedule(input.FeeSchedule)
	}
	if input.NameMode != "" {
		kv[domain.SettingKeyNameMode] = string(input.NameMode)
	}
	kv[domain.SettingKeyKioskSearch] = strconv.FormatBool(input.KioskSearch)
	if input.ReservationHours > 0 {
		kv[domain.SettingKeyReservationHours] = strconv.Itoa(input.ReservationHours)
	}
	kv[domain.SettingKeyDeregisterDeadlineH] = strconv.Itoa(input.DeregisterDeadlineH)
	if input.BillingMode != "" {
		kv[domain.SettingKeyBillingMode] = string(input.BillingMode)
	}
	if input.ReminderHourOfDay > 0 {
		kv[domain.SettingKeyReminderHourOfDay] = strconv.Itoa(input.ReminderHourOfDay)
	}
	if input.ReminderLeadWeeks > 0 {
		kv[domain.SettingKeyReminderLeadWeeks] = strconv.Itoa(input.ReminderLeadWeeks)
	}
	if input.BillingWarningLeadWeeks > 0 {
		kv[domain.SettingKeyBillingWarningLeadWeeks] = strconv.Itoa(input.BillingWarningLeadWeeks)
	}
	kv[domain.SettingKeyKioskLocked] = strconv.FormatBool(input.KioskLocked)

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
	// Save current config to history before overwriting (B-008).
	if current, err := uc.settings.GetBranding(ctx); err == nil && current != nil {
		histEntry := &domain.BrandingHistoryEntry{
			ID:        uuid.New(),
			Branding:  *current,
			CreatedAt: timeNow(),
		}
		_ = uc.settings.InsertBrandingHistory(ctx, histEntry)
	}

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

func (uc *SettingsUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after any) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}

// formatFeeSchedule serialises a fee schedule slice as a comma-separated string.
func formatFeeSchedule(fees []float32) string {
	parts := make([]string, len(fees))
	for i, f := range fees {
		parts[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
	}
	return strings.Join(parts, ",")
}

// parseFeeSchedule deserialises a comma-separated fee schedule string.
func parseFeeSchedule(s string) []float32 {
	parts := strings.Split(s, ",")
	out := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if f, err := strconv.ParseFloat(p, 32); err == nil {
			out = append(out, float32(f))
		}
	}
	return out
}

// timeNow returns the current UTC time. Exposed as a variable for testing.
var timeNow = func() time.Time { return time.Now().UTC() }

// GetBrandingHistory returns recent branding configuration snapshots (B-008).
func (uc *SettingsUsecase) GetBrandingHistory(ctx context.Context) ([]*domain.BrandingHistoryEntry, error) {
	return uc.settings.ListBrandingHistory(ctx, 20)
}

// RollbackBranding restores a previous branding configuration snapshot (B-008).
func (uc *SettingsUsecase) RollbackBranding(ctx context.Context, actorID uuid.UUID, id uuid.UUID) (*domain.BrandingConfig, error) {
	entry, err := uc.settings.GetBrandingHistoryEntry(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get branding history entry: %w", err)
	}

	// Save current config to history before rolling back.
	if current, err := uc.settings.GetBranding(ctx); err == nil && current != nil {
		histEntry := &domain.BrandingHistoryEntry{
			ID:        uuid.New(),
			Branding:  *current,
			CreatedAt: timeNow(),
		}
		_ = uc.settings.InsertBrandingHistory(ctx, histEntry)
	}

	if err := uc.settings.UpdateBranding(ctx, &entry.Branding); err != nil {
		return nil, fmt.Errorf("rollback branding: %w", err)
	}
	if uc.cache != nil {
		_ = uc.cache.Delete(ctx, "branding")
	}
	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityBranding, "branding_config", nil, map[string]string{"rolledBackFrom": id.String()})
	return &entry.Branding, nil
}
