package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
)

// Fixed UUIDs for seed data so tests can reference them.
var (
	AdminID    = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	MemberID   = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	EventID    = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	Shift1ID   = uuid.MustParse("00000000-0000-0000-0000-000000000011")
	Shift2ID   = uuid.MustParse("00000000-0000-0000-0000-000000000012")
	ClubYearID = uuid.MustParse("00000000-0000-0000-0000-000000000020")
)

// Seed pre-populates the provided repositories with test data.
func Seed(
	ctx context.Context,
	members *MemberRepo,
	events *EventRepo,
	shifts *ShiftRepo,
	hours *HourRepo,
) error {
	// --- Club year ---
	clubYear := &domain.ClubYear{
		ID:                 ClubYearID,
		Label:              "2026",
		StartDate:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
		DefaultTargetHours: 10,
		IsActive:           true,
	}
	if err := hours.CreateClubYear(ctx, clubYear); err != nil {
		return err
	}

	// --- Members ---
	admin := &domain.Member{
		ID:        AdminID,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@test.local",
		JoinedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
		Role:      domain.RoleVorstand,
	}
	if err := members.Create(ctx, admin); err != nil {
		return err
	}

	member := &domain.Member{
		ID:        MemberID,
		FirstName: "Max",
		LastName:  "Mustermann",
		Email:     "max@test.local",
		JoinedAt:  time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
		Role:      domain.RoleMitglied,
	}
	if err := members.Create(ctx, member); err != nil {
		return err
	}

	// --- Event ---
	now := time.Now().UTC()
	event := &domain.Event{
		ID:          EventID,
		Name:        "Sommerfest",
		Description: "Das jährliche Sommerfest",
		Location:    "Vereinsheim",
		Category:    "Fest",
		StartDate:   time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC),
		Status:      domain.EventStatusPublished,
		Visibility:  domain.EventVisibilityPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := events.Create(ctx, event); err != nil {
		return err
	}

	// --- Shifts ---
	shiftDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	shift1 := &domain.Shift{
		ID:         Shift1ID,
		EventID:    EventID,
		Name:       "Aufbau",
		StartAt:    time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC),
		EndAt:      time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
		MinHelpers: 2,
		MaxHelpers: 5,
		Date:       shiftDate,
	}
	if err := shifts.Create(ctx, shift1); err != nil {
		return err
	}

	shift2 := &domain.Shift{
		ID:         Shift2ID,
		EventID:    EventID,
		Name:       "Service",
		StartAt:    time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
		EndAt:      time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC),
		MinHelpers: 3,
		MaxHelpers: 8,
		Date:       shiftDate,
	}
	if err := shifts.Create(ctx, shift2); err != nil {
		return err
	}

	return nil
}
