package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// yearFeeTiers returns the fee tiers of a club year (G-001), falling back to
// the global feeSchedule setting when the year defines none.
func yearFeeTiers(ctx context.Context, settings port.SettingsRepository, clubYearID uuid.UUID) ([]domain.FeeTier, error) {
	tiers, err := settings.GetFeeTiers(ctx, clubYearID)
	if err != nil {
		return nil, fmt.Errorf("load fee tiers: %w", err)
	}

	out := make([]domain.FeeTier, len(tiers))
	for i, t := range tiers {
		out[i] = *t
	}

	if len(out) == 0 {
		if raw, err := settings.GetSetting(ctx, domain.SettingKeyFeeSchedule); err == nil && raw != "" {
			for i, euros := range parseFeeSchedule(raw) {
				out = append(out, domain.FeeTier{
					Position:    i + 1,
					AmountCents: int(euros * 100),
				})
			}
		}
	}
	return out, nil
}

// memberFeeTiers returns the tiers that apply to one member: a per-member
// override (G-004) takes precedence over the club-year-wide yearTiers.
func memberFeeTiers(ctx context.Context, settings port.SettingsRepository, memberID, clubYearID uuid.UUID, yearTiers []domain.FeeTier) []domain.FeeTier {
	override, err := settings.GetMemberFeeTiers(ctx, memberID, clubYearID)
	if err != nil || len(override) == 0 {
		return yearTiers
	}
	out := make([]domain.FeeTier, len(override))
	for i, t := range override {
		out[i] = *t
	}
	return out
}
