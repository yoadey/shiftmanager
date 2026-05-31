package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.MemberRepository = (*MemberRepo)(nil)

// MemberRepo is an in-memory implementation of port.MemberRepository.
type MemberRepo struct {
	mu       sync.RWMutex
	members  map[uuid.UUID]*domain.Member
	oidcLinks map[uuid.UUID][]*domain.OIDCLink // keyed by memberID
}

func NewMemberRepo() *MemberRepo {
	return &MemberRepo{
		members:   make(map[uuid.UUID]*domain.Member),
		oidcLinks: make(map[uuid.UUID][]*domain.OIDCLink),
	}
}

func copyMember(m *domain.Member) *domain.Member {
	c := *m
	return &c
}

func (r *MemberRepo) Create(_ context.Context, m *domain.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.members {
		if strings.EqualFold(existing.Email, m.Email) {
			return domain.ErrMemberEmailConflict
		}
	}
	r.members[m.ID] = copyMember(m)
	return nil
}

func (r *MemberRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.members[id]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	return copyMember(m), nil
}

func (r *MemberRepo) GetByEmail(_ context.Context, email string) (*domain.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.members {
		if strings.EqualFold(m.Email, email) {
			return copyMember(m), nil
		}
	}
	return nil, domain.ErrMemberNotFound
}

func (r *MemberRepo) GetByOIDCSubject(_ context.Context, provider, subject string) (*domain.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for memberID, links := range r.oidcLinks {
		for _, l := range links {
			if l.Provider == provider && l.Subject == subject {
				m, ok := r.members[memberID]
				if !ok {
					return nil, domain.ErrMemberNotFound
				}
				return copyMember(m), nil
			}
		}
	}
	return nil, domain.ErrMemberNotFound
}

func (r *MemberRepo) List(_ context.Context, filter port.MemberFilter) ([]*domain.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Member
	for _, m := range r.members {
		if filter.IsActive != nil && m.IsActive != *filter.IsActive {
			continue
		}
		if filter.Search != "" {
			q := strings.ToLower(filter.Search)
			haystack := strings.ToLower(m.FirstName + " " + m.LastName + " " + m.Email)
			if !strings.Contains(haystack, q) {
				continue
			}
		}
		result = append(result, copyMember(m))
	}

	// apply offset/limit
	if filter.Offset > 0 {
		if filter.Offset >= len(result) {
			return []*domain.Member{}, nil
		}
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *MemberRepo) Update(_ context.Context, m *domain.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.members[m.ID]; !ok {
		return domain.ErrMemberNotFound
	}
	r.members[m.ID] = copyMember(m)
	return nil
}

func (r *MemberRepo) Deactivate(_ context.Context, id uuid.UUID, leftAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.members[id]
	if !ok {
		return domain.ErrMemberNotFound
	}
	m.IsActive = false
	m.LeftAt = &leftAt
	return nil
}

func (r *MemberRepo) LinkOIDC(_ context.Context, link *domain.OIDCLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := *link
	r.oidcLinks[link.MemberID] = append(r.oidcLinks[link.MemberID], &c)
	return nil
}

func (r *MemberRepo) GetOIDCLinks(_ context.Context, memberID uuid.UUID) ([]*domain.OIDCLink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	links := r.oidcLinks[memberID]
	out := make([]*domain.OIDCLink, len(links))
	for i, l := range links {
		c := *l
		out[i] = &c
	}
	return out, nil
}

func (r *MemberRepo) Count(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.members), nil
}

func (r *MemberRepo) CountActive(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, m := range r.members {
		if m.IsActive {
			n++
		}
	}
	return n, nil
}

func (r *MemberRepo) SetReminderOptOut(_ context.Context, id uuid.UUID, optOut bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.members[id]
	if !ok {
		return domain.ErrMemberNotFound
	}
	m.ReminderOptOut = optOut
	return nil
}

func (r *MemberRepo) Anonymize(_ context.Context, id uuid.UUID, leftAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.members[id]
	if !ok {
		return domain.ErrMemberNotFound
	}
	m.FirstName = "DELETED"
	m.LastName = "DELETED"
	m.Email = "deleted@deleted.local"
	m.IsActive = false
	m.LeftAt = &leftAt
	m.OIDCSubject = nil
	return nil
}
