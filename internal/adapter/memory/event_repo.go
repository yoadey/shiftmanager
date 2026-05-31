package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.EventRepository = (*EventRepo)(nil)

// EventRepo is an in-memory implementation of port.EventRepository.
type EventRepo struct {
	mu     sync.RWMutex
	events map[uuid.UUID]*domain.Event
}

func NewEventRepo() *EventRepo {
	return &EventRepo{events: make(map[uuid.UUID]*domain.Event)}
}

func copyEvent(e *domain.Event) *domain.Event {
	c := *e
	return &c
}

func (r *EventRepo) Create(_ context.Context, e *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[e.ID] = copyEvent(e)
	return nil
}

func (r *EventRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.events[id]
	if !ok {
		return nil, domain.ErrEventNotFound
	}
	return copyEvent(e), nil
}

func (r *EventRepo) List(_ context.Context, filter port.EventFilter) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Event
	for _, e := range r.events {
		if filter.Status != nil && e.Status != *filter.Status {
			continue
		}
		if filter.Visibility != nil && e.Visibility != *filter.Visibility {
			continue
		}
		if filter.FromDate != nil && e.EndDate.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && e.StartDate.After(*filter.ToDate) {
			continue
		}
		result = append(result, copyEvent(e))
	}

	if filter.Offset > 0 {
		if filter.Offset >= len(result) {
			return []*domain.Event{}, nil
		}
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *EventRepo) Update(_ context.Context, e *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.events[e.ID]; !ok {
		return domain.ErrEventNotFound
	}
	r.events[e.ID] = copyEvent(e)
	return nil
}

func (r *EventRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.events[id]; !ok {
		return domain.ErrEventNotFound
	}
	delete(r.events, id)
	return nil
}

func (r *EventRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.EventStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.events[id]
	if !ok {
		return domain.ErrEventNotFound
	}
	e.Status = status
	return nil
}
