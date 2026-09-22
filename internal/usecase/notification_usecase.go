package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// NotificationUsecase implements the year-end billing/missing-hours mails
// (item 7) and the understaffed-shift notification (SC-007). It is driven by the
// scheduler and guarded to run at most once per day.
type NotificationUsecase struct {
	hours          port.HourRepository
	members        port.MemberRepository
	shifts         port.ShiftRepository
	events         port.EventRepository
	registrations  port.RegistrationRepository
	settings       port.SettingsRepository
	email          port.EmailService
	audit          port.AuditRepository
	organizerEmail string
	now            func() time.Time
}

// NewNotificationUsecase creates a new NotificationUsecase. organizerEmail is the
// address that receives understaffed-shift notices (typically the club address).
func NewNotificationUsecase(
	hours port.HourRepository,
	members port.MemberRepository,
	shifts port.ShiftRepository,
	events port.EventRepository,
	registrations port.RegistrationRepository,
	settings port.SettingsRepository,
	email port.EmailService,
	audit port.AuditRepository,
	organizerEmail string,
) *NotificationUsecase {
	return &NotificationUsecase{
		hours:          hours,
		members:        members,
		shifts:         shifts,
		events:         events,
		registrations:  registrations,
		settings:       settings,
		email:          email,
		audit:          audit,
		organizerEmail: organizerEmail,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

// RunDailyNotifications runs the year-end billing/warning mails and the
// understaffed-shift check. It returns the number of emails sent. It is designed
// to be invoked once per day by the scheduler; the time-window guards keep it
// from re-sending the same mails on repeated runs within the same day.
func (uc *NotificationUsecase) RunDailyNotifications(ctx context.Context) (int, error) {
	sent := 0

	n, err := uc.runYearEndMails(ctx)
	if err != nil {
		return sent, err
	}
	sent += n

	n, err = uc.notifyUnderstaffed(ctx)
	if err != nil {
		return sent, err
	}
	sent += n

	return sent, nil
}

// runYearEndMails sends the missing-hours warning X weeks before year end and the
// year-billing mail at/after year end, to active members with missing hours.
func (uc *NotificationUsecase) runYearEndMails(ctx context.Context) (int, error) {
	if uc.email == nil {
		return 0, nil
	}
	year, err := uc.hours.GetActiveClubYear(ctx)
	if err != nil {
		return 0, nil // no active year; nothing to do
	}

	now := uc.now()

	warnLeadWeeks := 4
	if v, err := getSettingInt(ctx, uc.settings, domain.SettingKeyBillingWarningLeadWeeks, 4); err == nil && v > 0 {
		warnLeadWeeks = v
	}
	warnStart := year.EndDate.Add(-time.Duration(warnLeadWeeks) * 7 * 24 * time.Hour)

	// Determine which phase we are in.
	billingPhase := !now.Before(year.EndDate)
	warningPhase := !billingPhase && !now.Before(warnStart)
	if !billingPhase && !warningPhase {
		return 0, nil
	}

	// BillingMode: only auto-send the year-billing mail when billing is automatic.
	billingMode := domain.BillingModeManual
	if v, err := uc.settings.GetSetting(ctx, domain.SettingKeyBillingMode); err == nil {
		billingMode = domain.BillingMode(v)
	}
	if billingPhase && billingMode != domain.BillingModeAuto {
		return 0, nil
	}

	tiers, _ := uc.settings.GetFeeTiers(ctx, year.ID)
	tierValues := make([]domain.FeeTier, len(tiers))
	for i, t := range tiers {
		tierValues[i] = *t
	}

	active := true
	members, err := uc.members.List(ctx, port.MemberFilter{IsActive: &active})
	if err != nil {
		return 0, fmt.Errorf("list members: %w", err)
	}

	entries, err := uc.hours.FindEntriesByYear(ctx, year.ID)
	if err != nil {
		return 0, fmt.Errorf("load year entries: %w", err)
	}
	confirmedByMember := make(map[uuid.UUID]float64)
	for _, e := range entries {
		if e.Status == domain.HourEntryStatusConfirmed {
			confirmedByMember[e.MemberID] += e.Hours
		}
	}

	sent := 0
	for _, m := range members {
		if m.Email == "" {
			continue
		}
		target := year.DefaultTargetHours
		if m.IndividualGoalHours != nil {
			target = *m.IndividualGoalHours
		}
		if ht, err := uc.hours.GetHourTarget(ctx, m.ID, year.ID); err == nil && ht != nil {
			target = ht.TargetHours
		}
		missing := target - confirmedByMember[m.ID]
		if missing <= 0 {
			continue
		}

		if billingPhase {
			memberTiers := tierValues
			if override, err := uc.settings.GetMemberFeeTiers(ctx, m.ID, year.ID); err == nil && len(override) > 0 {
				memberTiers = make([]domain.FeeTier, len(override))
				for i, t := range override {
					memberTiers[i] = *t
				}
			}
			account := domain.MemberHourAccount{MemberID: m.ID, ClubYearID: year.ID, TargetHours: target, ConfirmedHours: confirmedByMember[m.ID], MissingHours: missing}
			result := domain.ComputeBilling(*m, account, memberTiers)
			if err := uc.email.SendYearBilling(ctx, m.Email, m, missing, result.TotalCents, year); err == nil {
				sent++
			}
		} else {
			if err := uc.email.SendMissingHoursWarning(ctx, m.Email, m, missing, year); err == nil {
				sent++
			}
		}
	}

	phase := "warning"
	if billingPhase {
		phase = "billing"
	}
	_ = writeAuditEntry(ctx, uc.audit, nil, domain.AuditActionUpdate, domain.AuditEntityClubYear, year.ID.String(), nil,
		map[string]any{"yearEndMails": phase, "sent": sent})

	return sent, nil
}

// notifyUnderstaffed scans upcoming shifts and notifies the organizer about any
// shift starting within the next 7 days that is below its minimum helper count.
func (uc *NotificationUsecase) notifyUnderstaffed(ctx context.Context) (int, error) {
	if uc.email == nil || uc.organizerEmail == "" {
		return 0, nil
	}
	now := uc.now()
	shifts, err := uc.shifts.FindShiftsStartingBetween(ctx, now, now.Add(7*24*time.Hour))
	if err != nil {
		return 0, fmt.Errorf("find shifts: %w", err)
	}

	sent := 0
	for _, s := range shifts {
		if s.MinHelpers <= 0 {
			continue
		}
		count, err := uc.registrations.CountActiveByShift(ctx, s.ID)
		if err != nil || count >= s.MinHelpers {
			continue
		}
		event, err := uc.events.GetByID(ctx, s.EventID)
		if err != nil {
			continue
		}
		if err := uc.email.SendUnderstaffedNotice(ctx, uc.organizerEmail, s, event); err == nil {
			sent++
			_ = writeAuditEntry(ctx, uc.audit, nil, domain.AuditActionUpdate, domain.AuditEntityShift, s.ID.String(), nil,
				map[string]any{"understaffedNotice": true, "registered": count, "min": s.MinHelpers})
		}
	}
	return sent, nil
}

// NotifyShiftUnderstaffed sends an understaffed notice for a single shift if it
// has dropped below its minimum (called inline e.g. after a deregistration).
func (uc *NotificationUsecase) NotifyShiftUnderstaffed(ctx context.Context, shiftID uuid.UUID) error {
	if uc.email == nil || uc.organizerEmail == "" {
		return nil
	}
	shift, err := uc.shifts.GetByID(ctx, shiftID)
	if err != nil {
		return err
	}
	if shift.MinHelpers <= 0 {
		return nil
	}
	count, err := uc.registrations.CountActiveByShift(ctx, shiftID)
	if err != nil || count >= shift.MinHelpers {
		return nil
	}
	event, err := uc.events.GetByID(ctx, shift.EventID)
	if err != nil {
		return err
	}
	if err := uc.email.SendUnderstaffedNotice(ctx, uc.organizerEmail, shift, event); err != nil {
		return err
	}
	_ = writeAuditEntry(ctx, uc.audit, nil, domain.AuditActionUpdate, domain.AuditEntityShift, shiftID.String(), nil,
		map[string]any{"understaffedNotice": true, "registered": count, "min": shift.MinHelpers})
	return nil
}
