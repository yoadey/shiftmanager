package domain

import (
	"time"

	"github.com/google/uuid"
)

// BrandingHistoryEntry records a snapshot of the branding configuration at a
// point in time (B-008).
type BrandingHistoryEntry struct {
	ID        uuid.UUID      `json:"id" gorm:"type:varchar(36);primaryKey"`
	Branding  BrandingConfig `json:"branding" gorm:"serializer:json"`
	CreatedAt time.Time      `json:"createdAt"`
}
