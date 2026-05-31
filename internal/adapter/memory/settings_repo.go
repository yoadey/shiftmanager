package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.SettingsRepository = (*SettingsRepo)(nil)

// SettingsRepo is an in-memory implementation of port.SettingsRepository.
type SettingsRepo struct {
	mu             sync.RWMutex
	settings       map[string]string
	branding       *domain.BrandingConfig
	feeTiers       map[uuid.UUID][]*domain.FeeTier // keyed by clubYearID
	memberFeeTiers map[string][]*domain.FeeTier    // keyed by "memberID:clubYearID"
}

func NewSettingsRepo() *SettingsRepo {
	b := domain.DefaultBranding()
	return &SettingsRepo{
		settings:       make(map[string]string),
		branding:       &b,
		feeTiers:       make(map[uuid.UUID][]*domain.FeeTier),
		memberFeeTiers: make(map[string][]*domain.FeeTier),
	}
}

func memberFeeKey(memberID, clubYearID uuid.UUID) string {
	return fmt.Sprintf("%s:%s", memberID, clubYearID)
}

func (r *SettingsRepo) GetSetting(_ context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.settings[key], nil
}

func (r *SettingsRepo) SetSetting(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings[key] = value
	return nil
}

func (r *SettingsRepo) GetAllSettings(_ context.Context) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.settings))
	for k, v := range r.settings {
		out[k] = v
	}
	return out, nil
}

func (r *SettingsRepo) GetBranding(_ context.Context) (*domain.BrandingConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c := *r.branding
	return &c, nil
}

func (r *SettingsRepo) UpdateBranding(_ context.Context, b *domain.BrandingConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := *b
	r.branding = &c
	return nil
}

func copyFeeTiers(tiers []*domain.FeeTier) []*domain.FeeTier {
	out := make([]*domain.FeeTier, len(tiers))
	for i, t := range tiers {
		c := *t
		out[i] = &c
	}
	return out
}

func (r *SettingsRepo) GetFeeTiers(_ context.Context, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copyFeeTiers(r.feeTiers[clubYearID]), nil
}

func (r *SettingsRepo) ReplaceFeeTiers(_ context.Context, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.feeTiers[clubYearID] = copyFeeTiers(tiers)
	return nil
}

func (r *SettingsRepo) GetMemberFeeTiers(_ context.Context, memberID, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copyFeeTiers(r.memberFeeTiers[memberFeeKey(memberID, clubYearID)]), nil
}

func (r *SettingsRepo) ReplaceMemberFeeTiers(_ context.Context, memberID, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.memberFeeTiers[memberFeeKey(memberID, clubYearID)] = copyFeeTiers(tiers)
	return nil
}
