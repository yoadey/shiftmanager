package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// HourEntryType distinguishes shift-based from manually booked hours.
type HourEntryType string

const (
	HourEntryTypeShift  HourEntryType = "shift"
	HourEntryTypeManual HourEntryType = "manual"
)

// HourEntryStatus reflects whether hours have been confirmed.
type HourEntryStatus string

const (
	HourEntryStatusPending   HourEntryStatus = "pending"
	HourEntryStatusConfirmed HourEntryStatus = "confirmed"
)

// HourEntry records worked or booked hours for a member.
type HourEntry struct {
	ID          uuid.UUID       `json:"id"`
	MemberID    uuid.UUID       `json:"memberId"`
	ShiftID     *uuid.UUID      `json:"shiftId,omitempty"`
	ClubYearID  uuid.UUID       `json:"clubYearId"`
	Hours       float64         `json:"hours"`
	Type        HourEntryType   `json:"type"`
	Status      HourEntryStatus `json:"status"`
	BookedBy    *uuid.UUID      `json:"bookedBy,omitempty"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"createdAt"`
}

// ClubYear represents a membership/financial year for the club.
type ClubYear struct {
	ID                 uuid.UUID `json:"id"`
	Label              string    `json:"label"`
	StartDate          time.Time `json:"startDate"`
	EndDate            time.Time `json:"endDate"`
	DefaultTargetHours float64   `json:"defaultTargetHours"`
	IsActive           bool      `json:"isActive"`
}

// IsOpen returns true when the current time falls within the club year.
func (y *ClubYear) IsOpen(now time.Time) bool {
	return !now.Before(y.StartDate) && !now.After(y.EndDate)
}

// HourTarget stores a per-member override of the default year goal.
type HourTarget struct {
	ID         uuid.UUID `json:"id"`
	MemberID   uuid.UUID `json:"memberId"`
	ClubYearID uuid.UUID `json:"clubYearId"`
	TargetHours float64  `json:"targetHours"`
}

// MemberHourAccount is a summary of a member's hours for a year.
type MemberHourAccount struct {
	MemberID       uuid.UUID `json:"memberId"`
	ClubYearID     uuid.UUID `json:"clubYearId"`
	TargetHours    float64   `json:"targetHours"`
	ConfirmedHours float64   `json:"confirmedHours"`
	PendingHours   float64   `json:"pendingHours"`
	MissingHours   float64   `json:"missingHours"`
}

// YearSummaryRow is one row of the year-end summary.
type YearSummaryRow struct {
	Member         Member  `json:"member"`
	TargetHours    float64 `json:"targetHours"`
	ConfirmedHours float64 `json:"confirmedHours"`
	MissingHours   float64 `json:"missingHours"`
}

// Errors for hour operations.
var (
	ErrHourEntryNotFound = fmt.Errorf("hour entry not found")
	ErrClubYearNotFound  = fmt.Errorf("club year not found")
	ErrHourAlreadyConfirmed = fmt.Errorf("hour entry is already confirmed")
)
