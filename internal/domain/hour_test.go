package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClubYear_IsOpen(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	y := &ClubYear{StartDate: start, EndDate: end}

	assert.True(t, y.IsOpen(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))
	assert.True(t, y.IsOpen(start))
	assert.True(t, y.IsOpen(end))
	assert.False(t, y.IsOpen(start.Add(-time.Hour)))
	assert.False(t, y.IsOpen(end.Add(time.Hour)))
}

func TestNewAuditEntry(t *testing.T) {
	entry, err := NewAuditEntry(nil, AuditActionCreate, AuditEntityMember, "id-1",
		nil, map[string]string{"k": "v"})
	assert.NoError(t, err)
	assert.Equal(t, AuditActionCreate, entry.Action)
	assert.Equal(t, AuditEntityMember, entry.Entity)
	assert.Equal(t, "id-1", entry.EntityID)
	assert.Nil(t, entry.Before)
	assert.NotEmpty(t, entry.After)
	assert.False(t, entry.ChangedAt.IsZero())
}

func TestDefaultAppSettings(t *testing.T) {
	s := DefaultAppSettings()
	assert.Equal(t, NameModeAbbrev, s.NameMode)
	assert.Equal(t, 48, s.ReservationHours)
	assert.Equal(t, 24, s.DeregisterDeadlineH)
	assert.Equal(t, BillingModeManual, s.BillingMode)
}
