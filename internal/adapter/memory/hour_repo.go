package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.HourRepository = (*HourRepo)(nil)

// HourRepo is an in-memory implementation of port.HourRepository.
type HourRepo struct {
	mu          sync.RWMutex
	entries     map[uuid.UUID]*domain.HourEntry
	clubYears   map[uuid.UUID]*domain.ClubYear
	hourTargets map[string]*domain.HourTarget // key: "memberID:clubYearID"
}

func NewHourRepo() *HourRepo {
	return &HourRepo{
		entries:     make(map[uuid.UUID]*domain.HourEntry),
		clubYears:   make(map[uuid.UUID]*domain.ClubYear),
		hourTargets: make(map[string]*domain.HourTarget),
	}
}

func hourTargetKey(memberID, clubYearID uuid.UUID) string {
	return memberID.String() + ":" + clubYearID.String()
}

func copyHourEntry(e *domain.HourEntry) *domain.HourEntry {
	c := *e
	return &c
}

func copyClubYear(y *domain.ClubYear) *domain.ClubYear {
	c := *y
	return &c
}

func copyHourTarget(t *domain.HourTarget) *domain.HourTarget {
	c := *t
	return &c
}

func (r *HourRepo) CreateEntry(_ context.Context, e *domain.HourEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[e.ID] = copyHourEntry(e)
	return nil
}

func (r *HourRepo) GetEntryByID(_ context.Context, id uuid.UUID) (*domain.HourEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[id]
	if !ok {
		return nil, domain.ErrHourEntryNotFound
	}
	return copyHourEntry(e), nil
}

func (r *HourRepo) FindEntriesByMemberAndYear(_ context.Context, memberID, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.HourEntry
	for _, e := range r.entries {
		if e.MemberID == memberID && e.ClubYearID == clubYearID {
			result = append(result, copyHourEntry(e))
		}
	}
	return result, nil
}

func (r *HourRepo) FindEntriesByYear(_ context.Context, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.HourEntry
	for _, e := range r.entries {
		if e.ClubYearID == clubYearID {
			result = append(result, copyHourEntry(e))
		}
	}
	return result, nil
}

func (r *HourRepo) UpdateEntry(_ context.Context, e *domain.HourEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[e.ID]; !ok {
		return domain.ErrHourEntryNotFound
	}
	r.entries[e.ID] = copyHourEntry(e)
	return nil
}

func (r *HourRepo) DeleteEntry(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[id]; !ok {
		return domain.ErrHourEntryNotFound
	}
	delete(r.entries, id)
	return nil
}

func (r *HourRepo) CreateClubYear(_ context.Context, y *domain.ClubYear) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clubYears[y.ID] = copyClubYear(y)
	return nil
}

func (r *HourRepo) GetActiveClubYear(_ context.Context) (*domain.ClubYear, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, y := range r.clubYears {
		if y.IsActive {
			return copyClubYear(y), nil
		}
	}
	return nil, domain.ErrClubYearNotFound
}

func (r *HourRepo) GetClubYearByID(_ context.Context, id uuid.UUID) (*domain.ClubYear, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	y, ok := r.clubYears[id]
	if !ok {
		return nil, domain.ErrClubYearNotFound
	}
	return copyClubYear(y), nil
}

func (r *HourRepo) ListClubYears(_ context.Context) ([]*domain.ClubYear, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.ClubYear, 0, len(r.clubYears))
	for _, y := range r.clubYears {
		result = append(result, copyClubYear(y))
	}
	return result, nil
}

func (r *HourRepo) GetHourTarget(_ context.Context, memberID, clubYearID uuid.UUID) (*domain.HourTarget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.hourTargets[hourTargetKey(memberID, clubYearID)]
	if !ok {
		return nil, domain.ErrHourEntryNotFound
	}
	return copyHourTarget(t), nil
}

func (r *HourRepo) UpsertHourTarget(_ context.Context, t *domain.HourTarget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hourTargets[hourTargetKey(t.MemberID, t.ClubYearID)] = copyHourTarget(t)
	return nil
}
