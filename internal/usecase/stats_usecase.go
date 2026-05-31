package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// StatsUsecase aggregates system-wide statistics for the admin dashboard (D-004).
type StatsUsecase struct {
	hours         port.HourRepository
	members       port.MemberRepository
	shifts        port.ShiftRepository
	registrations port.RegistrationRepository
}

// NewStatsUsecase creates a new StatsUsecase.
func NewStatsUsecase(
	hours port.HourRepository,
	members port.MemberRepository,
	shifts port.ShiftRepository,
	registrations port.RegistrationRepository,
) *StatsUsecase {
	return &StatsUsecase{hours: hours, members: members, shifts: shifts, registrations: registrations}
}

// SystemStats is the system-wide statistics payload.
type SystemStats struct {
	ClubYearID          uuid.UUID `json:"clubYearId"`
	ClubYearLabel       string    `json:"clubYearLabel"`
	TotalConfirmedHours float64   `json:"totalConfirmedHours"`
	OpenShifts          int       `json:"openShifts"`     // upcoming shifts below min helpers
	UpcomingShifts      int       `json:"upcomingShifts"` // shifts starting in the future
	ActiveMembers       int       `json:"activeMembers"`
	MembersBelowTarget  int       `json:"membersBelowTarget"`
}

// GetStats computes system-wide statistics for the active club year.
func (uc *StatsUsecase) GetStats(ctx context.Context) (*SystemStats, error) {
	now := time.Now().UTC()

	stats := &SystemStats{}

	activeMembers, err := uc.members.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("count active members: %w", err)
	}
	stats.ActiveMembers = activeMembers

	// Upcoming and open (understaffed) shifts.
	upcoming, err := uc.shifts.FindUpcomingShifts(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("find upcoming shifts: %w", err)
	}
	stats.UpcomingShifts = len(upcoming)
	for _, s := range upcoming {
		count, err := uc.registrations.CountActiveByShift(ctx, s.ID)
		if err != nil {
			continue
		}
		if count < s.MinHelpers {
			stats.OpenShifts++
		}
	}

	// Hour-based stats for the active club year (best-effort: skip if none).
	year, err := uc.hours.GetActiveClubYear(ctx)
	if err != nil {
		return stats, nil
	}
	stats.ClubYearID = year.ID
	stats.ClubYearLabel = year.Label

	entries, err := uc.hours.FindEntriesByYear(ctx, year.ID)
	if err != nil {
		return nil, fmt.Errorf("load year entries: %w", err)
	}
	confirmedByMember := make(map[uuid.UUID]float64)
	for _, e := range entries {
		if e.Status == domain.HourEntryStatusConfirmed {
			confirmedByMember[e.MemberID] += e.Hours
			stats.TotalConfirmedHours += e.Hours
		}
	}

	active := true
	members, err := uc.members.List(ctx, port.MemberFilter{IsActive: &active})
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	for _, m := range members {
		target := year.DefaultTargetHours
		if m.IndividualGoalHours != nil {
			target = *m.IndividualGoalHours
		}
		if ht, err := uc.hours.GetHourTarget(ctx, m.ID, year.ID); err == nil && ht != nil {
			target = ht.TargetHours
		}
		if confirmedByMember[m.ID] < target {
			stats.MembersBelowTarget++
		}
	}

	return stats, nil
}
