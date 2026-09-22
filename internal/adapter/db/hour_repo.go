package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

var _ port.HourRepository = (*HourRepo)(nil)

// HourRepo is a GORM-backed implementation of port.HourRepository.
type HourRepo struct {
	db *gorm.DB
}

// NewHourRepo creates a new HourRepo.
func NewHourRepo(db *gorm.DB) *HourRepo {
	return &HourRepo{db: db}
}

// --- Hour entries ---

func (r *HourRepo) CreateEntry(ctx context.Context, e *domain.HourEntry) error {
	model := toHourEntryModel(e)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *HourRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.HourEntry, error) {
	var model HourEntryModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrHourEntryNotFound
		}
		return nil, err
	}
	return toHourEntryDomain(model), nil
}

func (r *HourRepo) FindEntriesByMemberAndYear(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	var models []HourEntryModel
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND club_year_id = ?", memberID.String(), clubYearID.String()).
		Order("created_at").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return hourEntrySlice(models), nil
}

func (r *HourRepo) FindEntriesByYear(ctx context.Context, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	var models []HourEntryModel
	err := r.db.WithContext(ctx).
		Where("club_year_id = ?", clubYearID.String()).
		Order("member_id, created_at").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return hourEntrySlice(models), nil
}

func (r *HourRepo) UpdateEntry(ctx context.Context, e *domain.HourEntry) error {
	model := toHourEntryModel(e)
	result := r.db.WithContext(ctx).
		Model(&HourEntryModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]interface{}{
			"hours":       model.Hours,
			"type":        model.Type,
			"status":      model.Status,
			"description": model.Description,
			"shift_id":    model.ShiftID,
			"booked_by":   model.BookedBy,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrHourEntryNotFound
	}
	return nil
}

func (r *HourRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&HourEntryModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrHourEntryNotFound
	}
	return nil
}

func hourEntrySlice(models []HourEntryModel) []*domain.HourEntry {
	entries := make([]*domain.HourEntry, 0, len(models))
	for _, m := range models {
		entries = append(entries, toHourEntryDomain(m))
	}
	return entries
}

func toHourEntryModel(e *domain.HourEntry) HourEntryModel {
	m := HourEntryModel{
		ID:          e.ID.String(),
		MemberID:    e.MemberID.String(),
		ClubYearID:  e.ClubYearID.String(),
		Hours:       e.Hours,
		Type:        string(e.Type),
		Status:      string(e.Status),
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
	}
	if e.ShiftID != nil {
		s := e.ShiftID.String()
		m.ShiftID = &s
	}
	if e.BookedBy != nil {
		s := e.BookedBy.String()
		m.BookedBy = &s
	}
	return m
}

func toHourEntryDomain(m HourEntryModel) *domain.HourEntry {
	e := &domain.HourEntry{
		ID:          uuid.MustParse(m.ID),
		MemberID:    uuid.MustParse(m.MemberID),
		ClubYearID:  uuid.MustParse(m.ClubYearID),
		Hours:       m.Hours,
		Type:        domain.HourEntryType(m.Type),
		Status:      domain.HourEntryStatus(m.Status),
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
	if m.ShiftID != nil {
		id := uuid.MustParse(*m.ShiftID)
		e.ShiftID = &id
	}
	if m.BookedBy != nil {
		id := uuid.MustParse(*m.BookedBy)
		e.BookedBy = &id
	}
	return e
}

// --- Club years ---

func (r *HourRepo) CreateClubYear(ctx context.Context, y *domain.ClubYear) error {
	model := toClubYearModel(y)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *HourRepo) GetActiveClubYear(ctx context.Context) (*domain.ClubYear, error) {
	var model ClubYearModel
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("start_date DESC").
		Limit(1).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrClubYearNotFound
		}
		return nil, err
	}
	return toClubYearDomain(model), nil
}

func (r *HourRepo) GetClubYearByID(ctx context.Context, id uuid.UUID) (*domain.ClubYear, error) {
	var model ClubYearModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrClubYearNotFound
		}
		return nil, err
	}
	return toClubYearDomain(model), nil
}

func (r *HourRepo) ListClubYears(ctx context.Context) ([]*domain.ClubYear, error) {
	var models []ClubYearModel
	if err := r.db.WithContext(ctx).Order("start_date DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	years := make([]*domain.ClubYear, 0, len(models))
	for _, m := range models {
		years = append(years, toClubYearDomain(m))
	}
	return years, nil
}

func toClubYearModel(y *domain.ClubYear) ClubYearModel {
	return ClubYearModel{
		ID:                 y.ID.String(),
		Label:              y.Label,
		StartDate:          y.StartDate,
		EndDate:            y.EndDate,
		DefaultTargetHours: y.DefaultTargetHours,
		IsActive:           y.IsActive,
	}
}

func toClubYearDomain(m ClubYearModel) *domain.ClubYear {
	return &domain.ClubYear{
		ID:                 uuid.MustParse(m.ID),
		Label:              m.Label,
		StartDate:          m.StartDate,
		EndDate:            m.EndDate,
		DefaultTargetHours: m.DefaultTargetHours,
		IsActive:           m.IsActive,
	}
}

// --- Hour targets ---

func (r *HourRepo) GetHourTarget(ctx context.Context, memberID, clubYearID uuid.UUID) (*domain.HourTarget, error) {
	var model HourTargetModel
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND club_year_id = ?", memberID.String(), clubYearID.String()).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrHourEntryNotFound
		}
		return nil, err
	}
	return &domain.HourTarget{
		ID:          uuid.MustParse(model.ID),
		MemberID:    uuid.MustParse(model.MemberID),
		ClubYearID:  uuid.MustParse(model.ClubYearID),
		TargetHours: model.TargetHours,
	}, nil
}

func (r *HourRepo) UpsertHourTarget(ctx context.Context, t *domain.HourTarget) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	model := HourTargetModel{
		ID:          t.ID.String(),
		MemberID:    t.MemberID.String(),
		ClubYearID:  t.ClubYearID.String(),
		TargetHours: t.TargetHours,
	}
	// Use save with a conflict clause: try to find by member+year and update, else create.
	var existing HourTargetModel
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND club_year_id = ?", model.MemberID, model.ClubYearID).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(&model).Error
	}
	return r.db.WithContext(ctx).Model(&existing).Update("target_hours", model.TargetHours).Error
}
