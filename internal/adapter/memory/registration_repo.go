package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.RegistrationRepository = (*RegistrationRepo)(nil)

// RegistrationRepo is an in-memory implementation of port.RegistrationRepository.
type RegistrationRepo struct {
	mu            sync.RWMutex
	registrations map[uuid.UUID]*domain.Registration
}

func NewRegistrationRepo() *RegistrationRepo {
	return &RegistrationRepo{registrations: make(map[uuid.UUID]*domain.Registration)}
}

func copyReg(r *domain.Registration) *domain.Registration {
	c := *r
	return &c
}

func (r *RegistrationRepo) Create(_ context.Context, reg *domain.Registration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registrations[reg.ID] = copyReg(reg)
	return nil
}

func (r *RegistrationRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reg, ok := r.registrations[id]
	if !ok {
		return nil, domain.ErrRegistrationNotFound
	}
	return copyReg(reg), nil
}

func (r *RegistrationRepo) GetByToken(_ context.Context, token uuid.UUID) (*domain.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, reg := range r.registrations {
		if reg.ConfirmationToken != nil && *reg.ConfirmationToken == token {
			return copyReg(reg), nil
		}
	}
	return nil, domain.ErrRegistrationNotFound
}

func (r *RegistrationRepo) FindByShiftID(_ context.Context, shiftID uuid.UUID) ([]*domain.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Registration
	for _, reg := range r.registrations {
		if reg.ShiftID == shiftID {
			result = append(result, copyReg(reg))
		}
	}
	return result, nil
}

func (r *RegistrationRepo) FindByMemberID(_ context.Context, memberID uuid.UUID) ([]*domain.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Registration
	for _, reg := range r.registrations {
		if reg.MemberID != nil && *reg.MemberID == memberID {
			result = append(result, copyReg(reg))
		}
	}
	return result, nil
}

func (r *RegistrationRepo) FindByMemberAndShift(_ context.Context, memberID uuid.UUID, shiftID uuid.UUID) (*domain.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, reg := range r.registrations {
		if reg.ShiftID == shiftID && reg.MemberID != nil && *reg.MemberID == memberID {
			return copyReg(reg), nil
		}
	}
	return nil, domain.ErrRegistrationNotFound
}

func (r *RegistrationRepo) FindByGuestEmailAndShift(_ context.Context, guestEmail string, shiftID uuid.UUID) (*domain.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, reg := range r.registrations {
		if reg.ShiftID == shiftID && reg.GuestEmail != nil && *reg.GuestEmail == guestEmail {
			return copyReg(reg), nil
		}
	}
	return nil, domain.ErrRegistrationNotFound
}

func (r *RegistrationRepo) CountActiveByShift(_ context.Context, shiftID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, reg := range r.registrations {
		if reg.ShiftID == shiftID &&
			reg.State != domain.RegistrationStateNoShow {
			n++
		}
	}
	return n, nil
}

func (r *RegistrationRepo) Update(_ context.Context, reg *domain.Registration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.registrations[reg.ID]; !ok {
		return domain.ErrRegistrationNotFound
	}
	r.registrations[reg.ID] = copyReg(reg)
	return nil
}

func (r *RegistrationRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.registrations[id]; !ok {
		return domain.ErrRegistrationNotFound
	}
	delete(r.registrations, id)
	return nil
}

func (r *RegistrationRepo) ListUnconfirmedExpiredReservations(_ context.Context, before time.Time) ([]*domain.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Registration
	for _, reg := range r.registrations {
		if reg.State == domain.RegistrationStateReserved &&
			reg.ReservedUntil != nil &&
			reg.ReservedUntil.Before(before) {
			result = append(result, copyReg(reg))
		}
	}
	return result, nil
}
