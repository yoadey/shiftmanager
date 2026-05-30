package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// MemberUsecase handles all member-related business logic.
type MemberUsecase struct {
	members port.MemberRepository
	audit   port.AuditRepository
}

// NewMemberUsecase creates a new MemberUsecase.
func NewMemberUsecase(members port.MemberRepository, audit port.AuditRepository) *MemberUsecase {
	return &MemberUsecase{members: members, audit: audit}
}

// CreateMember validates and persists a new member.
func (uc *MemberUsecase) CreateMember(ctx context.Context, actorID uuid.UUID, input CreateMemberInput) (*domain.Member, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if input.FirstName == "" || input.LastName == "" {
		return nil, fmt.Errorf("first and last name are required")
	}

	existing, err := uc.members.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, domain.ErrMemberEmailConflict
	}

	m := &domain.Member{
		ID:                  uuid.New(),
		FirstName:           strings.TrimSpace(input.FirstName),
		LastName:            strings.TrimSpace(input.LastName),
		Email:               input.Email,
		JoinedAt:            time.Now().UTC(),
		IsActive:            true,
		IndividualGoalHours: input.IndividualGoalHours,
		Role:                coalesceRole(input.Role, domain.RoleMitglied),
	}

	if input.JoinedAt != nil {
		m.JoinedAt = *input.JoinedAt
	}

	if err := uc.members.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("create member: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionCreate, domain.AuditEntityMember, m.ID.String(), nil, m)

	return m, nil
}

// ListMembers returns members matching the filter.
func (uc *MemberUsecase) ListMembers(ctx context.Context, filter port.MemberFilter) ([]*domain.Member, error) {
	return uc.members.List(ctx, filter)
}

// GetMember returns a single member by ID.
func (uc *MemberUsecase) GetMember(ctx context.Context, id uuid.UUID) (*domain.Member, error) {
	return uc.members.GetByID(ctx, id)
}

// UpdateMember applies changes to an existing member.
func (uc *MemberUsecase) UpdateMember(ctx context.Context, actorID uuid.UUID, id uuid.UUID, input UpdateMemberInput) (*domain.Member, error) {
	m, err := uc.members.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	before := *m

	if input.FirstName != "" {
		m.FirstName = strings.TrimSpace(input.FirstName)
	}
	if input.LastName != "" {
		m.LastName = strings.TrimSpace(input.LastName)
	}
	if input.Email != "" {
		input.Email = strings.ToLower(strings.TrimSpace(input.Email))
		if input.Email != m.Email {
			if existing, err := uc.members.GetByEmail(ctx, input.Email); err == nil && existing != nil {
				return nil, domain.ErrMemberEmailConflict
			}
			m.Email = input.Email
		}
	}
	if input.IndividualGoalHours != nil {
		m.IndividualGoalHours = input.IndividualGoalHours
	}
	if input.Role != "" {
		m.Role = input.Role
	}

	if err := uc.members.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("update member: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionUpdate, domain.AuditEntityMember, m.ID.String(), before, m)

	return m, nil
}

// DeactivateMember soft-deletes a member by marking them inactive.
func (uc *MemberUsecase) DeactivateMember(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	m, err := uc.members.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !m.IsActive {
		return domain.ErrMemberInactive
	}

	now := time.Now().UTC()
	if err := uc.members.Deactivate(ctx, id, now); err != nil {
		return fmt.Errorf("deactivate member: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionDeactivate, domain.AuditEntityMember, id.String(), m, nil)

	return nil
}

// LinkOIDC associates an OIDC subject with a member account.
func (uc *MemberUsecase) LinkOIDC(ctx context.Context, actorID uuid.UUID, memberID uuid.UUID, provider, subject string) error {
	m, err := uc.members.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if !m.IsActive {
		return domain.ErrMemberInactive
	}

	link := &domain.OIDCLink{
		ID:       uuid.New(),
		MemberID: memberID,
		Provider: provider,
		Subject:  subject,
		LinkedAt: time.Now().UTC(),
	}

	if err := uc.members.LinkOIDC(ctx, link); err != nil {
		return fmt.Errorf("link oidc: %w", err)
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionLinkOIDC, domain.AuditEntityMember, memberID.String(), nil, link)

	return nil
}

// ImportCSVPreview parses a CSV reader and returns a diff preview without writing.
func (uc *MemberUsecase) ImportCSVPreview(ctx context.Context, r io.Reader) (*domain.CSVImportPreview, error) {
	rows, err := parseCSV(r)
	if err != nil {
		return nil, err
	}

	preview := &domain.CSVImportPreview{}
	for _, row := range rows {
		existing, err := uc.members.GetByEmail(ctx, strings.ToLower(row.Email))
		if err != nil || existing == nil {
			preview.ToCreate = append(preview.ToCreate, row)
		} else {
			changed := existing.FirstName != row.FirstName || existing.LastName != row.LastName
			if changed {
				preview.ToUpdate = append(preview.ToUpdate, row)
			} else {
				preview.ToSkip = append(preview.ToSkip, row)
			}
		}
	}

	return preview, nil
}

// ImportCSV parses a CSV reader and upserts members, returning a summary.
func (uc *MemberUsecase) ImportCSV(ctx context.Context, actorID uuid.UUID, r io.Reader) (*domain.CSVImportPreview, error) {
	rows, err := parseCSV(r)
	if err != nil {
		return nil, err
	}

	result := &domain.CSVImportPreview{}
	for _, row := range rows {
		email := strings.ToLower(row.Email)
		existing, err := uc.members.GetByEmail(ctx, email)
		if err != nil || existing == nil {
			m := &domain.Member{
				ID:                  uuid.New(),
				FirstName:           row.FirstName,
				LastName:            row.LastName,
				Email:               email,
				JoinedAt:            time.Now().UTC(),
				IsActive:            true,
				IndividualGoalHours: row.IndividualGoalHours,
				Role:                domain.RoleMitglied,
			}
			if row.JoinedAt != nil {
				m.JoinedAt = *row.JoinedAt
			}
			if err := uc.members.Create(ctx, m); err == nil {
				result.ToCreate = append(result.ToCreate, row)
			}
		} else {
			existing.FirstName = row.FirstName
			existing.LastName = row.LastName
			if row.IndividualGoalHours != nil {
				existing.IndividualGoalHours = row.IndividualGoalHours
			}
			if err := uc.members.Update(ctx, existing); err == nil {
				result.ToUpdate = append(result.ToUpdate, row)
			}
		}
	}

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionImport, domain.AuditEntityMember, "csv", nil, map[string]int{
		"created": len(result.ToCreate),
		"updated": len(result.ToUpdate),
	})

	return result, nil
}

// ExportCSV writes all active members to a CSV and returns the bytes.
func (uc *MemberUsecase) ExportCSV(ctx context.Context, actorID uuid.UUID) ([]byte, error) {
	active := true
	members, err := uc.members.List(ctx, port.MemberFilter{IsActive: &active})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "first_name", "last_name", "email", "joined_at", "individual_goal_hours", "role"})

	for _, m := range members {
		goal := ""
		if m.IndividualGoalHours != nil {
			goal = strconv.FormatFloat(*m.IndividualGoalHours, 'f', 2, 64)
		}
		_ = w.Write([]string{
			m.ID.String(),
			m.FirstName,
			m.LastName,
			m.Email,
			m.JoinedAt.Format("2006-01-02"),
			goal,
			m.Role,
		})
	}
	w.Flush()

	aid := actorID
	_ = uc.writeAudit(ctx, &aid, domain.AuditActionExport, domain.AuditEntityMember, "csv", nil, nil)

	return buf.Bytes(), nil
}

// writeAudit is a best-effort audit log write; errors are intentionally ignored.
func (uc *MemberUsecase) writeAudit(ctx context.Context, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return uc.audit.Insert(ctx, entry)
}

// parseCSV parses a CSV stream into import rows.
// Expected columns: first_name, last_name, email, [joined_at], [individual_goal_hours]
func parseCSV(r io.Reader) ([]domain.CSVImportRow, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[strings.ToLower(strings.TrimSpace(col))] = i
	}

	mustHave := []string{"first_name", "last_name", "email"}
	for _, col := range mustHave {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("CSV missing required column: %s", col)
		}
	}

	var rows []domain.CSVImportRow
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV record: %w", err)
		}

		row := domain.CSVImportRow{
			FirstName: strings.TrimSpace(record[colIndex["first_name"]]),
			LastName:  strings.TrimSpace(record[colIndex["last_name"]]),
			Email:     strings.TrimSpace(record[colIndex["email"]]),
		}

		if idx, ok := colIndex["joined_at"]; ok && idx < len(record) && record[idx] != "" {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(record[idx]))
			if err == nil {
				row.JoinedAt = &t
			}
		}

		if idx, ok := colIndex["individual_goal_hours"]; ok && idx < len(record) && record[idx] != "" {
			f, err := strconv.ParseFloat(strings.TrimSpace(record[idx]), 64)
			if err == nil {
				row.IndividualGoalHours = &f
			}
		}

		if row.Email != "" {
			rows = append(rows, row)
		}
	}

	return rows, nil
}

// coalesceRole returns the first non-empty role, defaulting to the second argument.
func coalesceRole(role, def string) string {
	if role != "" {
		return role
	}
	return def
}

// CreateMemberInput holds the fields needed to create a member.
type CreateMemberInput struct {
	FirstName           string
	LastName            string
	Email               string
	JoinedAt            *time.Time
	IndividualGoalHours *float64
	Role                string
}

// UpdateMemberInput holds the fields that may be changed on a member.
type UpdateMemberInput struct {
	FirstName           string
	LastName            string
	Email               string
	IndividualGoalHours *float64
	Role                string
}
