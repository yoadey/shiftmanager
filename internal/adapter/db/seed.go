package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Fixed test UUIDs — stable across runs.
var (
	AdminID    = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	MemberID   = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	EventID    = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	Shift1ID   = uuid.MustParse("00000000-0000-0000-0000-000000000011")
	Shift2ID   = uuid.MustParse("00000000-0000-0000-0000-000000000012")
	ClubYearID = uuid.MustParse("00000000-0000-0000-0000-000000000020")
)

// SeedTestData inserts fixture data for integration tests.
// It seeds: active club year 2026, an admin (vorstand) member, a regular member,
// a published "Sommerfest" event with 2 shifts.
func SeedTestData(db *gorm.DB) error {
	now := time.Now().UTC()
	yearStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	eventStart := time.Date(2026, 7, 4, 14, 0, 0, 0, time.UTC)
	eventEnd := time.Date(2026, 7, 4, 23, 59, 0, 0, time.UTC)
	shift1Start := time.Date(2026, 7, 4, 14, 0, 0, 0, time.UTC)
	shift1End := time.Date(2026, 7, 4, 18, 0, 0, 0, time.UTC)
	shift2Start := time.Date(2026, 7, 4, 18, 0, 0, 0, time.UTC)
	shift2End := time.Date(2026, 7, 4, 22, 0, 0, 0, time.UTC)

	clubYear := ClubYearModel{
		ID:                 ClubYearID.String(),
		Label:              "2026",
		StartDate:          yearStart,
		EndDate:            yearEnd,
		DefaultTargetHours: 8,
		IsActive:           true,
	}
	if err := db.FirstOrCreate(&clubYear, "id = ?", clubYear.ID).Error; err != nil {
		return err
	}

	admin := MemberModel{
		ID:        AdminID.String(),
		FirstName: "Admin",
		LastName:  "Vorstand",
		Email:     "admin@test.local",
		JoinedAt:  now,
		IsActive:  true,
		Role:      "vorstand",
	}
	if err := db.FirstOrCreate(&admin, "id = ?", admin.ID).Error; err != nil {
		return err
	}

	member := MemberModel{
		ID:        MemberID.String(),
		FirstName: "Max",
		LastName:  "Mustermann",
		Email:     "max@test.local",
		JoinedAt:  now,
		IsActive:  true,
		Role:      "mitglied",
	}
	if err := db.FirstOrCreate(&member, "id = ?", member.ID).Error; err != nil {
		return err
	}

	event := EventModel{
		ID:          EventID.String(),
		Name:        "Sommerfest",
		Description: "Das alljährliche Sommerfest",
		Location:    "Vereinsheim",
		Category:    "fest",
		StartDate:   eventStart,
		EndDate:     eventEnd,
		Status:      "published",
		Visibility:  "public",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.FirstOrCreate(&event, "id = ?", event.ID).Error; err != nil {
		return err
	}

	shiftDate := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	shift1 := ShiftModel{
		ID:                    Shift1ID.String(),
		EventID:               EventID.String(),
		Name:                  "Aufbau",
		StartAt:               shift1Start,
		EndAt:                 shift1End,
		MinHelpers:            2,
		MaxHelpers:            5,
		RequiredQualification: "",
		ShiftDate:             shiftDate,
	}
	if err := db.FirstOrCreate(&shift1, "id = ?", shift1.ID).Error; err != nil {
		return err
	}

	shift2 := ShiftModel{
		ID:                    Shift2ID.String(),
		EventID:               EventID.String(),
		Name:                  "Service",
		StartAt:               shift2Start,
		EndAt:                 shift2End,
		MinHelpers:            2,
		MaxHelpers:            4,
		RequiredQualification: "",
		ShiftDate:             shiftDate,
	}
	if err := db.FirstOrCreate(&shift2, "id = ?", shift2.ID).Error; err != nil {
		return err
	}

	return nil
}
