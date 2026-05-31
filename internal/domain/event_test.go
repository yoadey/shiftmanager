package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEvent_IsMultiDay(t *testing.T) {
	start := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	single := &Event{StartDate: start, EndDate: start.Add(4 * time.Hour)}
	multi := &Event{StartDate: start, EndDate: start.Add(26 * time.Hour)}
	assert.False(t, single.IsMultiDay())
	assert.True(t, multi.IsMultiDay())
}

func TestEvent_CanPublishCanDelete(t *testing.T) {
	assert.True(t, (&Event{Status: EventStatusDraft}).CanPublish())
	assert.False(t, (&Event{Status: EventStatusPublished}).CanPublish())

	assert.True(t, (&Event{Status: EventStatusDraft}).CanDelete())
	assert.True(t, (&Event{Status: EventStatusCancelled}).CanDelete())
	assert.False(t, (&Event{Status: EventStatusPublished}).CanDelete())
	assert.False(t, (&Event{Status: EventStatusCompleted}).CanDelete())
}

func TestShift_DurationHours(t *testing.T) {
	start := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	s := &Shift{StartAt: start, EndAt: start.Add(150 * time.Minute)}
	assert.InDelta(t, 2.5, s.DurationHours(), 1e-9)
}

func TestRegistration_IsExpired(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name string
		reg  Registration
		want bool
	}{
		{"reserved past deadline", Registration{State: RegistrationStateReserved, ReservedUntil: &past}, true},
		{"reserved future deadline", Registration{State: RegistrationStateReserved, ReservedUntil: &future}, false},
		{"reserved no deadline", Registration{State: RegistrationStateReserved}, false},
		{"registered ignored", Registration{State: RegistrationStateRegistered, ReservedUntil: &past}, false},
		{"confirmed ignored", Registration{State: RegistrationStateConfirmed, ReservedUntil: &past}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.reg.IsExpired(now))
		})
	}
}

func TestShiftOccupancy_Status(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		registered int
		min        int
		max        int
		want       OccupancyStatus
	}{
		{"offen", 0, 2, 5, OccupancyOffen},
		{"teilweise", 1, 3, 5, OccupancyTeilweise},
		{"besetzt at min", 3, 3, 5, OccupancyBesetzt},
		{"besetzt above min", 4, 3, 5, OccupancyBesetzt},
		{"ausgebucht at max", 5, 3, 5, OccupancyAusgebucht},
		{"ausgebucht over max", 6, 3, 5, OccupancyAusgebucht},
		{"unbounded min reached", 10, 3, 0, OccupancyBesetzt},
		{"unbounded empty", 0, 3, 0, OccupancyOffen},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := ShiftOccupancy{ShiftID: id, Registered: tt.registered, MinHelpers: tt.min, MaxHelpers: tt.max}
			assert.Equal(t, tt.want, o.Status())
		})
	}
}

func TestShiftOccupancy_FreeSlots(t *testing.T) {
	tests := []struct {
		name       string
		registered int
		max        int
		want       int
	}{
		{"some free", 2, 5, 3},
		{"full", 5, 5, 0},
		{"over capacity", 7, 5, 0},
		{"unbounded", 3, 0, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := ShiftOccupancy{Registered: tt.registered, MaxHelpers: tt.max}
			assert.Equal(t, tt.want, o.FreeSlots())
		})
	}
}

func TestShiftOccupancy_NeedsMore(t *testing.T) {
	tests := []struct {
		name       string
		registered int
		min        int
		want       int
	}{
		{"needs more", 1, 4, 3},
		{"satisfied exactly", 4, 4, 0},
		{"satisfied over", 6, 4, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := ShiftOccupancy{Registered: tt.registered, MinHelpers: tt.min}
			assert.Equal(t, tt.want, o.NeedsMore())
		})
	}
}
