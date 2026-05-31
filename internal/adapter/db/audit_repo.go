package db

import (
	"context"

	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

var _ port.AuditRepository = (*AuditRepo)(nil)

// AuditRepo is a GORM-backed implementation of port.AuditRepository.
// The audit log is append-only.
type AuditRepo struct {
	db *gorm.DB
}

// NewAuditRepo creates a new AuditRepo.
func NewAuditRepo(db *gorm.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Insert(ctx context.Context, e *domain.AuditEntry) error {
	model := AuditEntryModel{
		ID:        e.ID.String(),
		Action:    e.Action,
		Entity:    e.Entity,
		EntityID:  e.EntityID,
		ChangedAt: e.ChangedAt,
	}
	if e.ActorID != nil {
		s := e.ActorID.String()
		model.ActorID = &s
	}
	if len(e.Before) > 0 {
		model.Before = string(e.Before)
	}
	if len(e.After) > 0 {
		model.After = string(e.After)
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *AuditRepo) List(ctx context.Context, filter port.AuditFilter) ([]*domain.AuditEntry, error) {
	q := r.db.WithContext(ctx).Model(&AuditEntryModel{})

	if filter.ActorID != nil {
		q = q.Where("actor_id = ?", filter.ActorID.String())
	}
	if filter.Entity != "" {
		q = q.Where("entity = ?", filter.Entity)
	}
	if filter.EntityID != "" {
		q = q.Where("entity_id = ?", filter.EntityID)
	}
	if filter.From != nil {
		q = q.Where("changed_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("changed_at <= ?", *filter.To)
	}

	q = q.Order("changed_at DESC")

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	q = q.Limit(limit)

	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}

	var models []AuditEntryModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	entries := make([]*domain.AuditEntry, 0, len(models))
	for _, m := range models {
		e := auditModelToDomain(m)
		entries = append(entries, e)
	}
	return entries, nil
}

func auditModelToDomain(m AuditEntryModel) *domain.AuditEntry {
	e := &domain.AuditEntry{
		Action:    m.Action,
		Entity:    m.Entity,
		EntityID:  m.EntityID,
		ChangedAt: m.ChangedAt,
	}
	e.ID = mustParseUUID(m.ID)
	if m.ActorID != nil {
		id := mustParseUUID(*m.ActorID)
		e.ActorID = &id
	}
	if m.Before != "" {
		e.Before = []byte(m.Before)
	}
	if m.After != "" {
		e.After = []byte(m.After)
	}
	return e
}
