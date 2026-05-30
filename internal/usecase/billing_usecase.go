package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// BillingUsecase handles year-end billing computation and export.
type BillingUsecase struct {
	hours    port.HourRepository
	members  port.MemberRepository
	settings port.SettingsRepository
	audit    port.AuditRepository
}

// NewBillingUsecase creates a new BillingUsecase.
func NewBillingUsecase(
	hours port.HourRepository,
	members port.MemberRepository,
	settings port.SettingsRepository,
	audit port.AuditRepository,
) *BillingUsecase {
	return &BillingUsecase{
		hours:    hours,
		members:  members,
		settings: settings,
		audit:    audit,
	}
}

// ComputeYearBilling calculates billing for all active members for the given club year.
func (uc *BillingUsecase) ComputeYearBilling(ctx context.Context, actorID uuid.UUID, clubYearID uuid.UUID) (*domain.YearBillingReport, error) {
	year, err := uc.hours.GetClubYearByID(ctx, clubYearID)
	if err != nil {
		return nil, err
	}

	tiers, err := uc.settings.GetFeeTiers(ctx, clubYearID)
	if err != nil {
		return nil, fmt.Errorf("load fee tiers: %w", err)
	}
	if len(tiers) == 0 {
		return nil, domain.ErrNoFeeTiers
	}

	// Convert pointer slice to value slice.
	tierValues := make([]domain.FeeTier, len(tiers))
	for i, t := range tiers {
		tierValues[i] = *t
	}

	active := true
	members, err := uc.members.List(ctx, port.MemberFilter{IsActive: &active})
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	allEntries, err := uc.hours.FindEntriesByYear(ctx, clubYearID)
	if err != nil {
		return nil, fmt.Errorf("load hour entries: %w", err)
	}

	confirmedByMember := make(map[uuid.UUID]float64)
	for _, e := range allEntries {
		if e.Status == domain.HourEntryStatusConfirmed {
			confirmedByMember[e.MemberID] += e.Hours
		}
	}

	report := &domain.YearBillingReport{
		ClubYear:   *year,
		ComputedAt: time.Now().UTC().Format(time.RFC3339),
	}

	for _, m := range members {
		target := year.DefaultTargetHours
		if m.IndividualGoalHours != nil {
			target = *m.IndividualGoalHours
		}
		if ht, err := uc.hours.GetHourTarget(ctx, m.ID, clubYearID); err == nil && ht != nil {
			target = ht.TargetHours
		}

		confirmed := confirmedByMember[m.ID]
		missing := target - confirmed
		if missing < 0 {
			missing = 0
		}

		account := domain.MemberHourAccount{
			MemberID:       m.ID,
			ClubYearID:     clubYearID,
			TargetHours:    target,
			ConfirmedHours: confirmed,
			MissingHours:   missing,
		}

		result := domain.ComputeBilling(*m, account, tierValues)
		report.Results = append(report.Results, result)
		report.TotalCents += result.TotalCents
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCompute, domain.AuditEntityClubYear, clubYearID.String(), nil, map[string]interface{}{
		"memberCount": len(report.Results),
		"totalCents":  report.TotalCents,
	})

	return report, nil
}

// ExportBillingCSV returns a CSV representation of the billing report.
func (uc *BillingUsecase) ExportBillingCSV(ctx context.Context, actorID uuid.UUID, clubYearID uuid.UUID) ([]byte, error) {
	report, err := uc.ComputeYearBilling(ctx, actorID, clubYearID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{
		"member_id", "last_name", "first_name", "email",
		"target_hours", "confirmed_hours", "missing_hours", "total_eur",
	})

	for _, r := range report.Results {
		_ = w.Write([]string{
			r.MemberID.String(),
			r.Member.LastName,
			r.Member.FirstName,
			r.Member.Email,
			fmt.Sprintf("%.2f", r.TargetHours),
			fmt.Sprintf("%.2f", r.ConfirmedHours),
			fmt.Sprintf("%.2f", r.MissingHours),
			fmt.Sprintf("%.2f", r.TotalEuro()),
		})
	}

	w.Flush()

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionExport, domain.AuditEntityClubYear, clubYearID.String(), nil, map[string]string{"format": "csv"})

	return buf.Bytes(), nil
}

// ExportBillingPDF returns the billing data as a CSV file.
// NOTE: PDF generation requires an external library (e.g. gofpdf or chromedp).
// This stub returns CSV-formatted data with a note indicating PDF is not yet implemented.
// Replace the body of this function with actual PDF rendering when a PDF library is added.
func (uc *BillingUsecase) ExportBillingPDF(ctx context.Context, actorID uuid.UUID, clubYearID uuid.UUID) ([]byte, string, error) {
	// STUB: PDF generation not implemented. Returns CSV instead.
	// To implement: add a PDF library such as github.com/jung-kurt/gofpdf
	// and render the YearBillingReport into a properly formatted PDF.
	data, err := uc.ExportBillingCSV(ctx, actorID, clubYearID)
	if err != nil {
		return nil, "", err
	}
	return data, "text/csv; charset=utf-8", nil
}

func (uc *BillingUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}
