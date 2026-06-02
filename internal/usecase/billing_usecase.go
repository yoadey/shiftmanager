package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
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

		// G-004: a member's per-member fee tier override takes precedence over the
		// club-year-wide tier list when present.
		memberTiers := tierValues
		if override, err := uc.settings.GetMemberFeeTiers(ctx, m.ID, clubYearID); err == nil && len(override) > 0 {
			memberTiers = make([]domain.FeeTier, len(override))
			for i, t := range override {
				memberTiers[i] = *t
			}
		}

		result := domain.ComputeBilling(*m, account, memberTiers)
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

// ExportBillingPDF renders the year billing report into a PDF document. It returns
// the PDF bytes and a suggested download filename (ending in .pdf).
func (uc *BillingUsecase) ExportBillingPDF(ctx context.Context, actorID uuid.UUID, clubYearID uuid.UUID) ([]byte, string, error) {
	report, err := uc.ComputeYearBilling(ctx, actorID, clubYearID)
	if err != nil {
		return nil, "", err
	}

	data, err := renderBillingPDF(report)
	if err != nil {
		return nil, "", fmt.Errorf("render billing pdf: %w", err)
	}

	filename := fmt.Sprintf("billing-%s.pdf", sanitizeFilename(report.ClubYear.Label))

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionExport, domain.AuditEntityClubYear, clubYearID.String(), nil, map[string]string{"format": "pdf"})

	return data, filename, nil
}

// renderBillingPDF builds the PDF document for a year billing report.
func renderBillingPDF(report *domain.YearBillingReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Beitragsabrechnung "+report.ClubYear.Label, false)
	pdf.AddPage()

	// Title.
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, fmt.Sprintf("Beitragsabrechnung %s", report.ClubYear.Label), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Erstellt: %s", report.ComputedAt), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// Table header.
	header := []string{"Mitglied", "Soll-Std.", "Best. Std.", "Fehl-Std.", "Betrag (EUR)"}
	widths := []float64{70, 28, 28, 28, 30}
	aligns := []string{"L", "R", "R", "R", "R"}

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(230, 230, 230)
	for i, h := range header {
		pdf.CellFormat(widths[i], 8, h, "1", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows.
	pdf.SetFont("Helvetica", "", 10)
	for _, r := range report.Results {
		cells := []string{
			r.Member.FullName(),
			fmt.Sprintf("%.2f", r.TargetHours),
			fmt.Sprintf("%.2f", r.ConfirmedHours),
			fmt.Sprintf("%.2f", r.MissingHours),
			formatCentsEUR(r.TotalCents),
		}
		for i, c := range cells {
			pdf.CellFormat(widths[i], 7, c, "1", 0, aligns[i], false, 0, "")
		}
		pdf.Ln(-1)
	}

	// Total row.
	pdf.SetFont("Helvetica", "B", 10)
	labelWidth := widths[0] + widths[1] + widths[2] + widths[3]
	pdf.CellFormat(labelWidth, 8, "Gesamt", "1", 0, "R", false, 0, "")
	pdf.CellFormat(widths[4], 8, formatCentsEUR(report.TotalCents), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// formatCentsEUR formats an integer cent amount as a EUR string, e.g. 1234 -> "12.34 EUR".
func formatCentsEUR(cents int) string {
	return fmt.Sprintf("%.2f EUR", float64(cents)/100.0)
}

// sanitizeFilename replaces characters that are unsafe in filenames with hyphens.
func sanitizeFilename(s string) string {
	if s == "" {
		return "report"
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}

func (uc *BillingUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}
