package memory

import (
	"context"
	"sync"

	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.AuditRepository = (*AuditRepo)(nil)

// AuditRepo is an in-memory implementation of port.AuditRepository.
type AuditRepo struct {
	mu      sync.RWMutex
	entries []*domain.AuditEntry
}

func NewAuditRepo() *AuditRepo {
	return &AuditRepo{}
}

func (r *AuditRepo) Insert(_ context.Context, e *domain.AuditEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := *e
	r.entries = append(r.entries, &c)
	return nil
}

func (r *AuditRepo) List(_ context.Context, filter port.AuditFilter) ([]*domain.AuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.AuditEntry
	for _, e := range r.entries {
		if filter.ActorID != nil {
			if e.ActorID == nil || *e.ActorID != *filter.ActorID {
				continue
			}
		}
		if filter.Entity != "" && e.Entity != filter.Entity {
			continue
		}
		if filter.EntityID != "" && e.EntityID != filter.EntityID {
			continue
		}
		if filter.From != nil && e.ChangedAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && e.ChangedAt.After(*filter.To) {
			continue
		}
		c := *e
		result = append(result, &c)
	}

	if filter.Offset > 0 {
		if filter.Offset >= len(result) {
			return []*domain.AuditEntry{}, nil
		}
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}
