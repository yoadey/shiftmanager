package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

var _ port.SettingsRepository = (*SettingsRepo)(nil)

// SettingsRepo is a GORM-backed implementation of port.SettingsRepository.
type SettingsRepo struct {
	db *gorm.DB
}

// NewSettingsRepo creates a new SettingsRepo.
func NewSettingsRepo(db *gorm.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

// --- Key/value settings ---

func (r *SettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	var model SettingModel
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("setting %q not found", key)
		}
		return "", err
	}
	return model.Value, nil
}

func (r *SettingsRepo) SetSetting(ctx context.Context, key, value string) error {
	var existing SettingModel
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(&SettingModel{Key: key, Value: value}).Error
	}
	return r.db.WithContext(ctx).Model(&existing).Update("value", value).Error
}

func (r *SettingsRepo) GetAllSettings(ctx context.Context) (map[string]string, error) {
	var models []SettingModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(models))
	for _, m := range models {
		out[m.Key] = m.Value
	}
	return out, nil
}

// --- Branding ---

func (r *SettingsRepo) GetBranding(ctx context.Context) (*domain.BrandingConfig, error) {
	var model BrandingModel
	err := r.db.WithContext(ctx).Where("id = ?", 1).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			def := domain.DefaultBranding()
			return &def, nil
		}
		return nil, err
	}
	return &domain.BrandingConfig{
		ClubName:     model.ClubName,
		PrimaryColor: model.PrimaryColor,
		AccentColor:  model.AccentColor,
		LogoURL:      model.LogoURL,
	}, nil
}

func (r *SettingsRepo) UpdateBranding(ctx context.Context, b *domain.BrandingConfig) error {
	var existing BrandingModel
	err := r.db.WithContext(ctx).Where("id = ?", 1).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	model := BrandingModel{
		ID:           1,
		ClubName:     b.ClubName,
		PrimaryColor: b.PrimaryColor,
		AccentColor:  b.AccentColor,
		LogoURL:      b.LogoURL,
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(&model).Error
	}
	return r.db.WithContext(ctx).Save(&model).Error
}

// --- Fee tiers ---

func (r *SettingsRepo) GetFeeTiers(ctx context.Context, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	var models []FeeTierModel
	if err := r.db.WithContext(ctx).
		Where("club_year_id = ?", clubYearID.String()).
		Order("position").
		Find(&models).Error; err != nil {
		return nil, err
	}
	tiers := make([]*domain.FeeTier, 0, len(models))
	for _, m := range models {
		tiers = append(tiers, &domain.FeeTier{
			ID:          uuid.MustParse(m.ID),
			ClubYearID:  uuid.MustParse(m.ClubYearID),
			Position:    m.Position,
			AmountCents: m.AmountCents,
		})
	}
	return tiers, nil
}

func (r *SettingsRepo) ReplaceFeeTiers(ctx context.Context, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("club_year_id = ?", clubYearID.String()).Delete(&FeeTierModel{}).Error; err != nil {
			return err
		}
		for _, t := range tiers {
			if t.ID == uuid.Nil {
				t.ID = uuid.New()
			}
			t.ClubYearID = clubYearID
			model := FeeTierModel{
				ID:          t.ID.String(),
				ClubYearID:  t.ClubYearID.String(),
				Position:    t.Position,
				AmountCents: t.AmountCents,
			}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// --- Per-member fee tier overrides (G-004) ---

func (r *SettingsRepo) GetMemberFeeTiers(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	var models []MemberFeeTierModel
	if err := r.db.WithContext(ctx).
		Where("member_id = ? AND club_year_id = ?", memberID.String(), clubYearID.String()).
		Order("position").
		Find(&models).Error; err != nil {
		return nil, err
	}
	tiers := make([]*domain.FeeTier, 0, len(models))
	for _, m := range models {
		tiers = append(tiers, &domain.FeeTier{
			ID:          uuid.MustParse(m.ID),
			ClubYearID:  uuid.MustParse(m.ClubYearID),
			Position:    m.Position,
			AmountCents: m.AmountCents,
		})
	}
	return tiers, nil
}

func (r *SettingsRepo) ReplaceMemberFeeTiers(ctx context.Context, memberID, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("member_id = ? AND club_year_id = ?", memberID.String(), clubYearID.String()).Delete(&MemberFeeTierModel{}).Error; err != nil {
			return err
		}
		for _, t := range tiers {
			if t.ID == uuid.Nil {
				t.ID = uuid.New()
			}
			t.ClubYearID = clubYearID
			model := MemberFeeTierModel{
				ID:          t.ID.String(),
				MemberID:    memberID.String(),
				ClubYearID:  t.ClubYearID.String(),
				Position:    t.Position,
				AmountCents: t.AmountCents,
			}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
