package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// ReminderUsecase sends reminder emails for upcoming shifts. It supports two
// reminder types: a one-week reminder and a one-day reminder.
type ReminderUsecase struct {
	shifts        port.ShiftRepository
	registrations port.RegistrationRepository
	events        port.EventRepository
	members       port.MemberRepository
	email         port.EmailService
	audit         port.AuditRepository
	settings      port.SettingsRepository
	now           func() time.Time
}

// NewReminderUsecase creates a new ReminderUsecase. settings may be nil, in which
// case the default reminder lead time is used.
func NewReminderUsecase(
	shifts port.ShiftRepository,
	registrations port.RegistrationRepository,
	events port.EventRepository,
	members port.MemberRepository,
	email port.EmailService,
	audit port.AuditRepository,
	settings port.SettingsRepository,
) *ReminderUsecase {
	return &ReminderUsecase{
		shifts:        shifts,
		registrations: registrations,
		events:        events,
		members:       members,
		email:         email,
		audit:         audit,
		settings:      settings,
		now:           func() time.Time { return time.Now().UTC() },
	}
}

// reminderWindow describes one reminder type: a [from, to) window of shift start
// times relative to now, and the number of days reported to the recipient.
type reminderWindow struct {
	from      time.Duration
	to        time.Duration
	daysUntil int
}

// SendDueReminders finds shifts starting in the next one-week and one-day windows
// and sends a reminder email to each active registrant (members and guests).
// It returns the number of emails sent.
//
// The scheduler runs this job hourly. To avoid re-sending every hour each window is
// only one hour wide, anchored at its reminder horizon (≈7d and ≈24h ahead).
func (uc *ReminderUsecase) SendDueReminders(ctx context.Context) (int, error) {
	now := uc.now()

	// N-002: the early reminder lead time is configurable (in weeks); default 1 week.
	leadWeeks := 1
	if uc.settings != nil {
		if v, err := getSettingInt(ctx, uc.settings, domain.SettingKeyReminderLeadWeeks, 1); err == nil && v > 0 {
			leadWeeks = v
		}
	}
	earlyLead := time.Duration(leadWeeks) * 7 * 24 * time.Hour

	windows := []reminderWindow{
		// Early reminder: shifts starting in [lead, lead+1h).
		{from: earlyLead, to: earlyLead + time.Hour, daysUntil: leadWeeks * 7},
		// One-day reminder: shifts starting in [24h, 25h).
		{from: 24 * time.Hour, to: 25 * time.Hour, daysUntil: 1},
	}

	sent := 0
	for _, win := range windows {
		from := now.Add(win.from)
		to := now.Add(win.to)

		shifts, err := uc.shifts.FindShiftsStartingBetween(ctx, from, to)
		if err != nil {
			return sent, fmt.Errorf("find shifts starting between: %w", err)
		}

		for _, shift := range shifts {
			n, err := uc.remindShift(ctx, shift, win.daysUntil)
			if err != nil {
				return sent, err
			}
			sent += n
		}
	}

	return sent, nil
}

// remindShift sends reminders for a single shift to all of its active registrants.
func (uc *ReminderUsecase) remindShift(ctx context.Context, shift *domain.Shift, daysUntil int) (int, error) {
	regs, err := uc.registrations.FindByShiftID(ctx, shift.ID)
	if err != nil {
		return 0, fmt.Errorf("load registrations: %w", err)
	}
	if len(regs) == 0 {
		return 0, nil
	}

	event, err := uc.events.GetByID(ctx, shift.EventID)
	if err != nil {
		return 0, fmt.Errorf("load event: %w", err)
	}

	sent := 0
	seen := make(map[string]bool)
	for _, reg := range regs {
		// Only remind active registrants (registered or confirmed).
		if reg.State != domain.RegistrationStateRegistered && reg.State != domain.RegistrationStateConfirmed {
			continue
		}

		// N-001: skip members who opted out of reminders. Mandatory mails
		// (confirmation/cancellation) are sent elsewhere and are unaffected.
		if reg.MemberID != nil {
			if m, err := uc.members.GetByID(ctx, *reg.MemberID); err == nil && m != nil && m.ReminderOptOut {
				continue
			}
		}

		to := uc.recipientEmail(ctx, reg)
		if to == "" || seen[to] {
			continue
		}
		seen[to] = true

		if uc.email == nil {
			continue
		}
		if err := uc.email.SendReminder(ctx, to, reg, shift, event, daysUntil); err != nil {
			// Best-effort: skip failed recipients but keep going.
			continue
		}
		sent++
		_ = uc.writeAudit(ctx, nil, domain.AuditActionRegister, domain.AuditEntityRegistration, reg.ID.String(), nil,
			map[string]interface{}{"reminder": true, "daysUntil": daysUntil, "to": to})
	}

	return sent, nil
}

// recipientEmail resolves the email address for a registration. Guests use their
// GuestEmail; members are resolved via the member repository.
func (uc *ReminderUsecase) recipientEmail(ctx context.Context, reg *domain.Registration) string {
	if reg.GuestEmail != nil && *reg.GuestEmail != "" {
		return *reg.GuestEmail
	}
	if reg.MemberID != nil {
		if m, err := uc.members.GetByID(ctx, *reg.MemberID); err == nil && m != nil {
			return m.Email
		}
	}
	return ""
}

func (uc *ReminderUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}
