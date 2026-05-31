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

var _ port.RegistrationRepository = (*RegistrationRepo)(nil)

// RegistrationRepo is a GORM-backed implementation of port.RegistrationRepository.
type RegistrationRepo struct {
	db *gorm.DB
}

// NewRegistrationRepo creates a new RegistrationRepo.
func NewRegistrationRepo(db *gorm.DB) *RegistrationRepo {
	return &RegistrationRepo{db: db}
}

func (r *RegistrationRepo) Create(ctx context.Context, reg *domain.Registration) error {
	model := toRegistrationModel(reg)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *RegistrationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Registration, error) {
	var model RegistrationModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRegistrationNotFound
		}
		return nil, err
	}
	return toRegistrationDomain(model), nil
}

func (r *RegistrationRepo) GetByToken(ctx context.Context, token uuid.UUID) (*domain.Registration, error) {
	tokenStr := token.String()
	var model RegistrationModel
	err := r.db.WithContext(ctx).Where("confirmation_token = ?", tokenStr).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRegistrationNotFound
		}
		return nil, err
	}
	return toRegistrationDomain(model), nil
}

func (r *RegistrationRepo) FindByShiftID(ctx context.Context, shiftID uuid.UUID) ([]*domain.Registration, error) {
	var models []RegistrationModel
	if err := r.db.WithContext(ctx).Where("shift_id = ?", shiftID.String()).Order("created_at").Find(&models).Error; err != nil {
		return nil, err
	}
	return registrationSlice(models), nil
}

func (r *RegistrationRepo) FindByMemberID(ctx context.Context, memberID uuid.UUID) ([]*domain.Registration, error) {
	var models []RegistrationModel
	if err := r.db.WithContext(ctx).Where("member_id = ?", memberID.String()).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return registrationSlice(models), nil
}

func (r *RegistrationRepo) FindByMemberAndShift(ctx context.Context, memberID uuid.UUID, shiftID uuid.UUID) (*domain.Registration, error) {
	var model RegistrationModel
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND shift_id = ?", memberID.String(), shiftID.String()).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRegistrationNotFound
		}
		return nil, err
	}
	return toRegistrationDomain(model), nil
}

func (r *RegistrationRepo) FindByGuestEmailAndShift(ctx context.Context, guestEmail string, shiftID uuid.UUID) (*domain.Registration, error) {
	var model RegistrationModel
	err := r.db.WithContext(ctx).
		Where("guest_email = ? AND shift_id = ?", strings.ToLower(guestEmail), shiftID.String()).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRegistrationNotFound
		}
		return nil, err
	}
	return toRegistrationDomain(model), nil
}

func (r *RegistrationRepo) CountActiveByShift(ctx context.Context, shiftID uuid.UUID) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&RegistrationModel{}).
		Where("shift_id = ? AND state IN ?", shiftID.String(), []string{"registered", "confirmed"}).
		Count(&n).Error
	return int(n), err
}

func (r *RegistrationRepo) Update(ctx context.Context, reg *domain.Registration) error {
	model := toRegistrationModel(reg)
	result := r.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRegistrationNotFound
	}
	return nil
}

func (r *RegistrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&RegistrationModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRegistrationNotFound
	}
	return nil
}

func (r *RegistrationRepo) ListUnconfirmedExpiredReservations(ctx context.Context, before time.Time) ([]*domain.Registration, error) {
	var models []RegistrationModel
	err := r.db.WithContext(ctx).
		Where("state = ? AND reserved_until IS NOT NULL AND reserved_until < ?", "reserved", before).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return registrationSlice(models), nil
}

func registrationSlice(models []RegistrationModel) []*domain.Registration {
	regs := make([]*domain.Registration, 0, len(models))
	for _, m := range models {
		regs = append(regs, toRegistrationDomain(m))
	}
	return regs
}

func toRegistrationModel(r *domain.Registration) RegistrationModel {
	m := RegistrationModel{
		ID:        r.ID.String(),
		ShiftID:   r.ShiftID.String(),
		State:     string(r.State),
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt,
	}
	if r.MemberID != nil {
		s := r.MemberID.String()
		m.MemberID = &s
	}
	if r.GuestEmail != nil {
		ge := *r.GuestEmail
		m.GuestEmail = &ge
	}
	m.ReservedUntil = r.ReservedUntil
	m.BookedHours = r.BookedHours
	if r.ConfirmationToken != nil {
		s := r.ConfirmationToken.String()
		m.ConfirmationToken = &s
	}
	return m
}

func toRegistrationDomain(m RegistrationModel) *domain.Registration {
	r := &domain.Registration{
		ID:            uuid.MustParse(m.ID),
		ShiftID:       uuid.MustParse(m.ShiftID),
		State:         domain.RegistrationState(m.State),
		Comment:       m.Comment,
		ReservedUntil: m.ReservedUntil,
		BookedHours:   m.BookedHours,
		CreatedAt:     m.CreatedAt,
	}
	if m.MemberID != nil {
		id := uuid.MustParse(*m.MemberID)
		r.MemberID = &id
	}
	if m.GuestEmail != nil {
		ge := *m.GuestEmail
		r.GuestEmail = &ge
	}
	if m.ConfirmationToken != nil {
		id := uuid.MustParse(*m.ConfirmationToken)
		r.ConfirmationToken = &id
	}
	return r
}
