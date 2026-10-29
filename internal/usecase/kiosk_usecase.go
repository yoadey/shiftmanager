package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// Bounds for a kiosk link's validity in days (K-016).
const (
	MinKioskLinkDays = 1
	MaxKioskLinkDays = 3650
)

// KioskUsecase manages the links that unlock the kiosk on a device (K-016).
type KioskUsecase struct {
	links port.KioskLinkRepository
	audit port.AuditRepository
}

// NewKioskUsecase creates a new KioskUsecase.
func NewKioskUsecase(links port.KioskLinkRepository, audit port.AuditRepository) *KioskUsecase {
	return &KioskUsecase{links: links, audit: audit}
}

// CreateLink creates a kiosk link valid for validDays and returns it together
// with its token. The token is only returned here; just its hash is stored.
func (uc *KioskUsecase) CreateLink(ctx context.Context, actorID uuid.UUID, label string, validDays int) (*domain.KioskLink, string, error) {
	if validDays < MinKioskLinkDays || validDays > MaxKioskLinkDays {
		return nil, "", fmt.Errorf("validity must be between %d and %d days", MinKioskLinkDays, MaxKioskLinkDays)
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	now := time.Now().UTC()
	aid := actorID
	link := &domain.KioskLink{
		ID:        uuid.New(),
		Label:     strings.TrimSpace(label),
		CreatedBy: &aid,
		CreatedAt: now,
		ExpiresAt: now.AddDate(0, 0, validDays),
	}
	if err := uc.links.Create(ctx, link, hashKioskToken(token)); err != nil {
		return nil, "", fmt.Errorf("create kiosk link: %w", err)
	}
	_ = writeAuditEntry(ctx, uc.audit, &aid, domain.AuditActionCreate, domain.AuditEntityKioskLink, link.ID.String(), nil, link)
	return link, token, nil
}

// ListLinks returns all kiosk links, newest first.
func (uc *KioskUsecase) ListLinks(ctx context.Context) ([]*domain.KioskLink, error) {
	return uc.links.List(ctx)
}

// RevokeLink stops a kiosk link from unlocking the kiosk.
func (uc *KioskUsecase) RevokeLink(ctx context.Context, actorID, id uuid.UUID) error {
	if err := uc.links.Revoke(ctx, id, time.Now().UTC()); err != nil {
		return err
	}
	aid := actorID
	_ = writeAuditEntry(ctx, uc.audit, &aid, domain.AuditActionRevoke, domain.AuditEntityKioskLink, id.String(), nil, nil)
	return nil
}

// ValidToken reports whether token belongs to an active (not expired, not
// revoked) kiosk link.
func (uc *KioskUsecase) ValidToken(ctx context.Context, token string) bool {
	if token == "" {
		return false
	}
	link, err := uc.links.GetByTokenHash(ctx, hashKioskToken(token))
	return err == nil && link.ActiveAt(time.Now().UTC())
}

func hashKioskToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
