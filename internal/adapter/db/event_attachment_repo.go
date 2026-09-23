package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

var _ port.EventAttachmentRepository = (*EventAttachmentRepo)(nil)

// EventAttachmentRepo is a GORM-backed implementation of port.EventAttachmentRepository.
type EventAttachmentRepo struct {
	db *gorm.DB
}

// NewEventAttachmentRepo creates a new EventAttachmentRepo.
func NewEventAttachmentRepo(db *gorm.DB) *EventAttachmentRepo {
	return &EventAttachmentRepo{db: db}
}

func (r *EventAttachmentRepo) Create(ctx context.Context, a *domain.EventAttachment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	model := EventAttachmentModel{
		ID:          a.ID.String(),
		EventID:     a.EventID.String(),
		FileName:    a.FileName,
		URL:         a.URL,
		ContentType: a.ContentType,
		SizeBytes:   a.SizeBytes,
		UploadedAt:  a.UploadedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *EventAttachmentRepo) ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*domain.EventAttachment, error) {
	var models []EventAttachmentModel
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID.String()).Order("uploaded_at").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.EventAttachment, 0, len(models))
	for _, m := range models {
		out = append(out, eventAttachmentModelToDomain(m))
	}
	return out, nil
}

func (r *EventAttachmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EventAttachment, error) {
	var model EventAttachmentModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrEventAttachmentNotFound
		}
		return nil, err
	}
	return eventAttachmentModelToDomain(model), nil
}

func (r *EventAttachmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&EventAttachmentModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEventAttachmentNotFound
	}
	return nil
}

func (r *EventAttachmentRepo) DeleteByEvent(ctx context.Context, eventID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("event_id = ?", eventID.String()).Delete(&EventAttachmentModel{}).Error
}

func eventAttachmentModelToDomain(m EventAttachmentModel) *domain.EventAttachment {
	return &domain.EventAttachment{
		ID:          uuid.MustParse(m.ID),
		EventID:     uuid.MustParse(m.EventID),
		FileName:    m.FileName,
		URL:         m.URL,
		ContentType: m.ContentType,
		SizeBytes:   m.SizeBytes,
		UploadedAt:  m.UploadedAt,
	}
}
