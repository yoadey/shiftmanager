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

var _ port.EventRepository = (*EventRepo)(nil)
var _ port.ShiftRepository = (*ShiftRepo)(nil)

// EventRepo is a GORM-backed implementation of port.EventRepository.
type EventRepo struct {
	db *gorm.DB
}

// NewEventRepo creates a new EventRepo.
func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Create(ctx context.Context, e *domain.Event) error {
	model := toEventModel(e)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *EventRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrEventNotFound
		}
		return nil, err
	}
	return toEventDomain(model), nil
}

func (r *EventRepo) List(ctx context.Context, filter port.EventFilter) ([]*domain.Event, error) {
	q := r.db.WithContext(ctx).Model(&EventModel{})

	if filter.Status != nil {
		q = q.Where("status = ?", string(*filter.Status))
	}
	if filter.Visibility != nil {
		q = q.Where("visibility = ?", string(*filter.Visibility))
	}
	if filter.FromDate != nil {
		q = q.Where("end_date >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		q = q.Where("start_date <= ?", *filter.ToDate)
	}

	q = q.Order("start_date DESC")

	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}

	var models []EventModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	events := make([]*domain.Event, 0, len(models))
	for _, m := range models {
		events = append(events, toEventDomain(m))
	}
	return events, nil
}

func (r *EventRepo) Update(ctx context.Context, e *domain.Event) error {
	model := toEventModel(e)
	result := r.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

func (r *EventRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&EventModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

func (r *EventRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EventStatus) error {
	result := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("id = ?", id.String()).
		Updates(map[string]interface{}{"status": string(status), "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

func toEventModel(e *domain.Event) EventModel {
	return EventModel{
		ID:          e.ID.String(),
		Name:        e.Name,
		Description: e.Description,
		Location:    e.Location,
		Category:    e.Category,
		StartDate:   e.StartDate,
		EndDate:     e.EndDate,
		Status:      string(e.Status),
		Visibility:  string(e.Visibility),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func toEventDomain(m EventModel) *domain.Event {
	return &domain.Event{
		ID:          uuid.MustParse(m.ID),
		Name:        m.Name,
		Description: m.Description,
		Location:    m.Location,
		Category:    m.Category,
		StartDate:   m.StartDate,
		EndDate:     m.EndDate,
		Status:      domain.EventStatus(m.Status),
		Visibility:  domain.EventVisibility(m.Visibility),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// --- ShiftRepo ---

// ShiftRepo is a GORM-backed implementation of port.ShiftRepository.
type ShiftRepo struct {
	db *gorm.DB
}

// NewShiftRepo creates a new ShiftRepo.
func NewShiftRepo(db *gorm.DB) *ShiftRepo {
	return &ShiftRepo{db: db}
}

func (r *ShiftRepo) Create(ctx context.Context, s *domain.Shift) error {
	model := toShiftModel(s)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *ShiftRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shift, error) {
	var model ShiftModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrShiftNotFound
		}
		return nil, err
	}
	return toShiftDomain(model), nil
}

func (r *ShiftRepo) FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain.Shift, error) {
	var models []ShiftModel
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID.String()).Order("start_at").Find(&models).Error; err != nil {
		return nil, err
	}
	return shiftSlice(models), nil
}

func (r *ShiftRepo) FindShiftsStartingBetween(ctx context.Context, from, to time.Time) ([]*domain.Shift, error) {
	var models []ShiftModel
	if err := r.db.WithContext(ctx).
		Where("start_at >= ? AND start_at < ?", from, to).
		Order("start_at").
		Find(&models).Error; err != nil {
		return nil, err
	}
	return shiftSlice(models), nil
}

func (r *ShiftRepo) FindUpcomingShifts(ctx context.Context, after time.Time) ([]*domain.Shift, error) {
	var models []ShiftModel
	if err := r.db.WithContext(ctx).Where("start_at >= ?", after).Order("start_at").Find(&models).Error; err != nil {
		return nil, err
	}
	return shiftSlice(models), nil
}

func (r *ShiftRepo) Update(ctx context.Context, s *domain.Shift) error {
	model := toShiftModel(s)
	result := r.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrShiftNotFound
	}
	return nil
}

func (r *ShiftRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&ShiftModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrShiftNotFound
	}
	return nil
}

func shiftSlice(models []ShiftModel) []*domain.Shift {
	shifts := make([]*domain.Shift, 0, len(models))
	for _, m := range models {
		shifts = append(shifts, toShiftDomain(m))
	}
	return shifts
}

func toShiftModel(s *domain.Shift) ShiftModel {
	return ShiftModel{
		ID:                    s.ID.String(),
		EventID:               s.EventID.String(),
		Name:                  s.Name,
		StartAt:               s.StartAt,
		EndAt:                 s.EndAt,
		MinHelpers:            s.MinHelpers,
		MaxHelpers:            s.MaxHelpers,
		RequiredQualification: s.RequiredQualification,
		ShiftDate:             s.Date,
	}
}

func toShiftDomain(m ShiftModel) *domain.Shift {
	return &domain.Shift{
		ID:                    uuid.MustParse(m.ID),
		EventID:               uuid.MustParse(m.EventID),
		Name:                  m.Name,
		StartAt:               m.StartAt,
		EndAt:                 m.EndAt,
		MinHelpers:            m.MinHelpers,
		MaxHelpers:            m.MaxHelpers,
		RequiredQualification: m.RequiredQualification,
		Date:                  m.ShiftDate,
	}
}
