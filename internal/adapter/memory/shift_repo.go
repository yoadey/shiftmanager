package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.ShiftRepository = (*ShiftRepo)(nil)

// ShiftRepo is an in-memory implementation of port.ShiftRepository.
type ShiftRepo struct {
	mu     sync.RWMutex
	shifts map[uuid.UUID]*domain.Shift
}

func NewShiftRepo() *ShiftRepo {
	return &ShiftRepo{shifts: make(map[uuid.UUID]*domain.Shift)}
}

func copyShift(s *domain.Shift) *domain.Shift {
	c := *s
	return &c
}

func (r *ShiftRepo) Create(_ context.Context, s *domain.Shift) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shifts[s.ID] = copyShift(s)
	return nil
}

func (r *ShiftRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Shift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.shifts[id]
	if !ok {
		return nil, domain.ErrShiftNotFound
	}
	return copyShift(s), nil
}

func (r *ShiftRepo) FindByEventID(_ context.Context, eventID uuid.UUID) ([]*domain.Shift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Shift
	for _, s := range r.shifts {
		if s.EventID == eventID {
			result = append(result, copyShift(s))
		}
	}
	return result, nil
}

func (r *ShiftRepo) FindShiftsStartingBetween(_ context.Context, from, to time.Time) ([]*domain.Shift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Shift
	for _, s := range r.shifts {
		if !s.StartAt.Before(from) && !s.StartAt.After(to) {
			result = append(result, copyShift(s))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartAt.Before(result[j].StartAt) })
	return result, nil
}

func (r *ShiftRepo) FindUpcomingShifts(_ context.Context, after time.Time) ([]*domain.Shift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Shift
	for _, s := range r.shifts {
		if !s.StartAt.Before(after) {
			result = append(result, copyShift(s))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartAt.Before(result[j].StartAt) })
	return result, nil
}

func (r *ShiftRepo) Update(_ context.Context, s *domain.Shift) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.shifts[s.ID]; !ok {
		return domain.ErrShiftNotFound
	}
	r.shifts[s.ID] = copyShift(s)
	return nil
}

func (r *ShiftRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.shifts[id]; !ok {
		return domain.ErrShiftNotFound
	}
	delete(r.shifts, id)
	return nil
}
