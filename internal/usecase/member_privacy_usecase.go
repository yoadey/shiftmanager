package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// MemberPrivacyUsecase implements the GDPR data export (DS-003), right-to-be-
// forgotten deletion (DS-004) and per-member reminder preferences (N-001).
type MemberPrivacyUsecase struct {
	members       port.MemberRepository
	registrations port.RegistrationRepository
	hours         port.HourRepository
	audit         port.AuditRepository
}

// NewMemberPrivacyUsecase creates a new MemberPrivacyUsecase.
func NewMemberPrivacyUsecase(
	members port.MemberRepository,
	registrations port.RegistrationRepository,
	hours port.HourRepository,
	audit port.AuditRepository,
) *MemberPrivacyUsecase {
	return &MemberPrivacyUsecase{members: members, registrations: registrations, hours: hours, audit: audit}
}

// MemberDataExport is the GDPR data-portability document for a single member.
type MemberDataExport struct {
	Member        *domain.Member         `json:"member"`
	Registrations []*domain.Registration `json:"registrations"`
	HourEntries   []*domain.HourEntry    `json:"hourEntries"`
	HourTargets   []*domain.HourTarget   `json:"hourTargets"`
	ExportedAt    time.Time              `json:"exportedAt"`
}

// ExportData assembles a complete export of a member's personal data (DS-003).
func (uc *MemberPrivacyUsecase) ExportData(ctx context.Context, actorID uuid.UUID, memberID uuid.UUID) (*MemberDataExport, error) {
	member, err := uc.members.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	regs, err := uc.registrations.FindByMemberID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("load registrations: %w", err)
	}

	// Collect hour entries and targets across all club years.
	var entries []*domain.HourEntry
	var targets []*domain.HourTarget
	if years, err := uc.hours.ListClubYears(ctx); err == nil {
		for _, y := range years {
			if es, err := uc.hours.FindEntriesByMemberAndYear(ctx, memberID, y.ID); err == nil {
				entries = append(entries, es...)
			}
			if t, err := uc.hours.GetHourTarget(ctx, memberID, y.ID); err == nil && t != nil {
				targets = append(targets, t)
			}
		}
	}

	export := &MemberDataExport{
		Member:        member,
		Registrations: regs,
		HourEntries:   entries,
		HourTargets:   targets,
		ExportedAt:    time.Now().UTC(),
	}

	aid := actorID
	_ = writeAuditEntry(ctx, uc.audit, &aid, domain.AuditActionGDPRExport, domain.AuditEntityMember, memberID.String(), nil,
		map[string]int{"registrations": len(regs), "hourEntries": len(entries)})

	return export, nil
}

// GDPRDelete anonymizes a member rather than hard-deleting them (DS-004).
//
// Exception / retention rationale: hour_entries and audit_log rows are NOT
// deleted because they are financial/accounting records the club is legally
// required to retain (Aufbewahrungspflicht). Instead the member row's PII
// (names, email) is overwritten with redacted placeholders, the OIDC link is
// removed and the account is deactivated. The hour entries remain attached to
// the now-anonymized member id, so billing/audit history stays consistent
// without exposing personal data.
func (uc *MemberPrivacyUsecase) GDPRDelete(ctx context.Context, actorID uuid.UUID, memberID uuid.UUID) error {
	member, err := uc.members.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	before := *member

	now := time.Now().UTC()
	if err := uc.members.Anonymize(ctx, memberID, now); err != nil {
		return fmt.Errorf("anonymize member: %w", err)
	}

	aid := actorID
	_ = writeAuditEntry(ctx, uc.audit, &aid, domain.AuditActionGDPRDelete, domain.AuditEntityMember, memberID.String(), before,
		map[string]string{"result": "anonymized", "note": "financial/audit records retained"})

	return nil
}

// SetReminderOptOut updates a member's own reminder opt-out preference (N-001).
func (uc *MemberPrivacyUsecase) SetReminderOptOut(ctx context.Context, actorID uuid.UUID, memberID uuid.UUID, optOut bool) error {
	if _, err := uc.members.GetByID(ctx, memberID); err != nil {
		return err
	}
	if err := uc.members.SetReminderOptOut(ctx, memberID, optOut); err != nil {
		return fmt.Errorf("set reminder opt-out: %w", err)
	}

	aid := actorID
	_ = writeAuditEntry(ctx, uc.audit, &aid, domain.AuditActionUpdate, domain.AuditEntityMember, memberID.String(), nil,
		map[string]bool{"reminderOptOut": optOut})
	return nil
}
