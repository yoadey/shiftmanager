package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

var _ port.KioskLinkRepository = (*KioskLinkRepo)(nil)

// KioskLinkRepo is a GORM-backed implementation of port.KioskLinkRepository.
type KioskLinkRepo struct {
	db *gorm.DB
}

// NewKioskLinkRepo creates a new KioskLinkRepo.
func NewKioskLinkRepo(db *gorm.DB) *KioskLinkRepo {
	return &KioskLinkRepo{db: db}
}

func (r *KioskLinkRepo) Create(ctx context.Context, l *domain.KioskLink, tokenHash string) error {
	m := KioskLinkModel{
		ID:        l.ID.String(),
		Label:     l.Label,
		TokenHash: tokenHash,
		CreatedAt: l.CreatedAt,
		ExpiresAt: l.ExpiresAt,
		RevokedAt: l.RevokedAt,
	}
	if l.CreatedBy != nil {
		s := l.CreatedBy.String()
		m.CreatedBy = &s
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *KioskLinkRepo) List(ctx context.Context) ([]*domain.KioskLink, error) {
	var models []KioskLinkModel
	if err := r.db.WithContext(ctx).Order("created_at DESC, id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.KioskLink, 0, len(models))
	for _, m := range models {
		out = append(out, toKioskLinkDomain(m))
	}
	return out, nil
}

func (r *KioskLinkRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.KioskLink, error) {
	var m KioskLinkModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrKioskLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	return toKioskLinkDomain(m), nil
}

func (r *KioskLinkRepo) Revoke(ctx context.Context, id uuid.UUID, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&KioskLinkModel{}).
		Where("id = ? AND revoked_at IS NULL", id.String()).
		Update("revoked_at", at)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrKioskLinkNotFound
	}
	return nil
}

func toKioskLinkDomain(m KioskLinkModel) *domain.KioskLink {
	l := &domain.KioskLink{
		ID:        uuid.MustParse(m.ID),
		Label:     m.Label,
		CreatedAt: m.CreatedAt,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
	}
	if m.CreatedBy != nil {
		id := uuid.MustParse(*m.CreatedBy)
		l.CreatedBy = &id
	}
	return l
}
