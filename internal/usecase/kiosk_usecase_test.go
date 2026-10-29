package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

type fakeKioskLinkRepo struct {
	links  map[uuid.UUID]*domain.KioskLink
	hashes map[string]uuid.UUID
}

var _ port.KioskLinkRepository = (*fakeKioskLinkRepo)(nil)

func newFakeKioskLinkRepo() *fakeKioskLinkRepo {
	return &fakeKioskLinkRepo{links: map[uuid.UUID]*domain.KioskLink{}, hashes: map[string]uuid.UUID{}}
}

func (f *fakeKioskLinkRepo) Create(ctx context.Context, l *domain.KioskLink, tokenHash string) error {
	cp := *l
	f.links[l.ID] = &cp
	f.hashes[tokenHash] = l.ID
	return nil
}

func (f *fakeKioskLinkRepo) List(ctx context.Context) ([]*domain.KioskLink, error) {
	var out []*domain.KioskLink
	for _, l := range f.links {
		cp := *l
		out = append(out, &cp)
	}
	return out, nil
}

func (f *fakeKioskLinkRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.KioskLink, error) {
	id, ok := f.hashes[tokenHash]
	if !ok {
		return nil, domain.ErrKioskLinkNotFound
	}
	cp := *f.links[id]
	return &cp, nil
}

func (f *fakeKioskLinkRepo) Revoke(ctx context.Context, id uuid.UUID, at time.Time) error {
	l, ok := f.links[id]
	if !ok || l.RevokedAt != nil {
		return domain.ErrKioskLinkNotFound
	}
	l.RevokedAt = &at
	return nil
}

func TestKioskLinks(t *testing.T) {
	repo := newFakeKioskLinkRepo()
	audit := newFakeAuditRepo()
	uc := NewKioskUsecase(repo, audit)
	ctx := context.Background()
	actor := uuid.New()

	link, token, err := uc.CreateLink(ctx, actor, " Vereinsheim ", 90)
	require.NoError(t, err)
	assert.Equal(t, "Vereinsheim", link.Label)
	assert.WithinDuration(t, time.Now().UTC().AddDate(0, 0, 90), link.ExpiresAt, time.Minute)
	assert.Len(t, token, 43, "32 random bytes, base64url")
	assert.True(t, uc.ValidToken(ctx, token))
	assert.False(t, uc.ValidToken(ctx, token+"x"))
	assert.False(t, uc.ValidToken(ctx, ""))
	for hash := range repo.hashes {
		assert.NotContains(t, hash, token, "only a hash is stored")
	}
	assert.True(t, audit.has(domain.AuditActionCreate, domain.AuditEntityKioskLink))

	// Expired links no longer unlock the kiosk.
	repo.links[link.ID].ExpiresAt = time.Now().UTC().Add(-time.Minute)
	assert.False(t, uc.ValidToken(ctx, token))
	repo.links[link.ID].ExpiresAt = time.Now().UTC().Add(time.Hour)

	require.NoError(t, uc.RevokeLink(ctx, actor, link.ID))
	assert.False(t, uc.ValidToken(ctx, token))
	assert.ErrorIs(t, uc.RevokeLink(ctx, actor, link.ID), domain.ErrKioskLinkNotFound)
	assert.True(t, audit.has(domain.AuditActionRevoke, domain.AuditEntityKioskLink))

	for _, days := range []int{0, -1, MaxKioskLinkDays + 1} {
		_, _, err := uc.CreateLink(ctx, actor, "", days)
		assert.Error(t, err, days)
	}
}
