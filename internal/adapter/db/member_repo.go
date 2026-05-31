package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

var _ port.MemberRepository = (*MemberRepo)(nil)

// MemberRepo is a GORM-backed implementation of port.MemberRepository.
type MemberRepo struct {
	db *gorm.DB
}

// NewMemberRepo creates a new MemberRepo.
func NewMemberRepo(db *gorm.DB) *MemberRepo {
	return &MemberRepo{db: db}
}

func (r *MemberRepo) Create(ctx context.Context, m *domain.Member) error {
	model := toMemberModel(m)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	return nil
}

func (r *MemberRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error) {
	var model MemberModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	return toMemberDomain(model), nil
}

func (r *MemberRepo) GetByEmail(ctx context.Context, email string) (*domain.Member, error) {
	var model MemberModel
	err := r.db.WithContext(ctx).Where("email = ?", strings.ToLower(email)).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	return toMemberDomain(model), nil
}

func (r *MemberRepo) GetByOIDCSubject(ctx context.Context, provider, subject string) (*domain.Member, error) {
	var model MemberModel
	err := r.db.WithContext(ctx).
		Joins("JOIN oidc_links ON oidc_links.member_id = members.id").
		Where("oidc_links.provider = ? AND oidc_links.subject = ?", provider, subject).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	return toMemberDomain(model), nil
}

func (r *MemberRepo) List(ctx context.Context, filter port.MemberFilter) ([]*domain.Member, error) {
	q := r.db.WithContext(ctx).Model(&MemberModel{})

	if filter.IsActive != nil {
		q = q.Where("is_active = ?", *filter.IsActive)
	}
	if s := strings.TrimSpace(filter.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("lower(first_name) LIKE ? OR lower(last_name) LIKE ? OR lower(email) LIKE ?", like, like, like)
	}

	q = q.Order("last_name, first_name")

	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}

	var models []MemberModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	members := make([]*domain.Member, 0, len(models))
	for _, m := range models {
		members = append(members, toMemberDomain(m))
	}
	return members, nil
}

func (r *MemberRepo) Update(ctx context.Context, m *domain.Member) error {
	model := toMemberModel(m)
	result := r.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

func (r *MemberRepo) Deactivate(ctx context.Context, id uuid.UUID, leftAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&MemberModel{}).
		Where("id = ?", id.String()).
		Updates(map[string]interface{}{"is_active": false, "left_at": leftAt})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

func (r *MemberRepo) LinkOIDC(ctx context.Context, link *domain.OIDCLink) error {
	model := OIDCLinkModel{
		ID:       link.ID.String(),
		MemberID: link.MemberID.String(),
		Provider: link.Provider,
		Subject:  link.Subject,
		LinkedAt: link.LinkedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *MemberRepo) GetOIDCLinks(ctx context.Context, memberID uuid.UUID) ([]*domain.OIDCLink, error) {
	var models []OIDCLinkModel
	if err := r.db.WithContext(ctx).Where("member_id = ?", memberID.String()).Order("linked_at").Find(&models).Error; err != nil {
		return nil, err
	}
	links := make([]*domain.OIDCLink, 0, len(models))
	for _, m := range models {
		id := uuid.MustParse(m.ID)
		mid := uuid.MustParse(m.MemberID)
		links = append(links, &domain.OIDCLink{
			ID:       id,
			MemberID: mid,
			Provider: m.Provider,
			Subject:  m.Subject,
			LinkedAt: m.LinkedAt,
		})
	}
	return links, nil
}

func (r *MemberRepo) Count(ctx context.Context) (int, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&MemberModel{}).Count(&n).Error; err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *MemberRepo) CountActive(ctx context.Context) (int, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&MemberModel{}).Where("is_active = ?", true).Count(&n).Error; err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *MemberRepo) SetReminderOptOut(ctx context.Context, id uuid.UUID, optOut bool) error {
	result := r.db.WithContext(ctx).Model(&MemberModel{}).
		Where("id = ?", id.String()).
		Update("reminder_opt_out", optOut)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

func (r *MemberRepo) Anonymize(ctx context.Context, id uuid.UUID, leftAt time.Time) error {
	idStr := id.String()
	result := r.db.WithContext(ctx).Model(&MemberModel{}).
		Where("id = ?", idStr).
		Updates(map[string]interface{}{
			"first_name":   "Geloeschtes",
			"last_name":    "Mitglied",
			"email":        "deleted+" + idStr + "@invalid.local",
			"oidc_subject": nil,
			"is_active":    false,
			"left_at":      leftAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrMemberNotFound
	}
	// Delete OIDC links so the account can no longer be logged into.
	return r.db.WithContext(ctx).Where("member_id = ?", idStr).Delete(&OIDCLinkModel{}).Error
}

// --- Conversion helpers ---

func toMemberModel(m *domain.Member) MemberModel {
	return MemberModel{
		ID:                  m.ID.String(),
		FirstName:           m.FirstName,
		LastName:            m.LastName,
		Email:               m.Email,
		JoinedAt:            m.JoinedAt,
		LeftAt:              m.LeftAt,
		IsActive:            m.IsActive,
		IndividualGoalHours: m.IndividualGoalHours,
		OIDCSubject:         m.OIDCSubject,
		Role:                m.Role,
		ReminderOptOut:      m.ReminderOptOut,
	}
}

func toMemberDomain(m MemberModel) *domain.Member {
	return &domain.Member{
		ID:                  uuid.MustParse(m.ID),
		FirstName:           m.FirstName,
		LastName:            m.LastName,
		Email:               m.Email,
		JoinedAt:            m.JoinedAt,
		LeftAt:              m.LeftAt,
		IsActive:            m.IsActive,
		IndividualGoalHours: m.IndividualGoalHours,
		OIDCSubject:         m.OIDCSubject,
		Role:                m.Role,
		ReminderOptOut:      m.ReminderOptOut,
	}
}
