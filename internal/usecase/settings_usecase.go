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

	return &s, nil
}

// UpdateSettings persists changed application settings.
func (uc *SettingsUsecase) UpdateSettings(ctx context.Context, actorID uuid.UUID, input domain.AppSettings) (*domain.AppSettings, error) {
	kv := map[string]string{
		domain.SettingKeyNameMode:            string(input.NameMode),
		domain.SettingKeyKioskSearch:         strconv.FormatBool(input.KioskSearch),
		domain.SettingKeyReservationHours:    strconv.Itoa(input.ReservationHours),
		domain.SettingKeyDeregisterDeadlineH: strconv.Itoa(input.DeregisterDeadlineH),
		domain.SettingKeyBillingMode:         string(input.BillingMode),
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

// UpdateBranding persists a new branding configuration.
func (uc *SettingsUsecase) UpdateBranding(ctx context.Context, actorID uuid.UUID, input domain.BrandingConfig) (*domain.BrandingConfig, error) {
	if err := uc.settings.UpdateBranding(ctx, &input); err != nil {
		return nil, fmt.Errorf("update branding: %w", err)
	}

	// Invalidate cache.
	if uc.cache != nil {
		_ = uc.cache.Delete(ctx, "branding")
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityBranding, "branding_config", nil, input)

	return &input, nil
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
