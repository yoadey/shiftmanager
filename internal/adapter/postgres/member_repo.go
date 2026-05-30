package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// MemberRepo is a pgxpool-backed implementation of port.MemberRepository.
type MemberRepo struct {
	pool *pgxpool.Pool
}

var _ port.MemberRepository = (*MemberRepo)(nil)

// NewMemberRepo creates a new MemberRepo.
func NewMemberRepo(pool *pgxpool.Pool) *MemberRepo {
	return &MemberRepo{pool: pool}
}

const memberColumns = `id, first_name, last_name, email, joined_at, left_at, is_active, individual_goal_hours, oidc_subject, role`

const sqlInsertMember = `
INSERT INTO members (id, first_name, last_name, email, joined_at, left_at, is_active, individual_goal_hours, oidc_subject, role)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (r *MemberRepo) Create(ctx context.Context, m *domain.Member) error {
	_, err := r.pool.Exec(ctx, sqlInsertMember,
		m.ID, m.FirstName, m.LastName, m.Email, m.JoinedAt, m.LeftAt,
		m.IsActive, m.IndividualGoalHours, m.OIDCSubject, m.Role,
	)
	if err != nil {
		return fmt.Errorf("insert member: %w", err)
	}
	return nil
}

const sqlGetMemberByID = `SELECT ` + memberColumns + ` FROM members WHERE id = $1`

func (r *MemberRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error) {
	row := r.pool.QueryRow(ctx, sqlGetMemberByID, id)
	return scanMember(row, domain.ErrMemberNotFound)
}

const sqlGetMemberByEmail = `SELECT ` + memberColumns + ` FROM members WHERE email = $1`

func (r *MemberRepo) GetByEmail(ctx context.Context, email string) (*domain.Member, error) {
	row := r.pool.QueryRow(ctx, sqlGetMemberByEmail, strings.ToLower(email))
	return scanMember(row, domain.ErrMemberNotFound)
}

const sqlGetMemberByOIDC = `
SELECT ` + memberColumns + `
FROM members m
JOIN oidc_links l ON l.member_id = m.id
WHERE l.provider = $1 AND l.subject = $2`

func (r *MemberRepo) GetByOIDCSubject(ctx context.Context, provider, subject string) (*domain.Member, error) {
	row := r.pool.QueryRow(ctx, sqlGetMemberByOIDC, provider, subject)
	return scanMember(row, domain.ErrMemberNotFound)
}

func (r *MemberRepo) List(ctx context.Context, filter port.MemberFilter) ([]*domain.Member, error) {
	var sb strings.Builder
	sb.WriteString(`SELECT ` + memberColumns + ` FROM members WHERE 1=1`)
	args := []any{}
	idx := 1

	if filter.IsActive != nil {
		sb.WriteString(fmt.Sprintf(" AND is_active = $%d", idx))
		args = append(args, *filter.IsActive)
		idx++
	}
	if s := strings.TrimSpace(filter.Search); s != "" {
		sb.WriteString(fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+s+"%")
		idx++
	}

	sb.WriteString(" ORDER BY last_name, first_name")

	if filter.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT $%d", idx))
		args = append(args, filter.Limit)
		idx++
	}
	if filter.Offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET $%d", idx))
		args = append(args, filter.Offset)
		idx++
	}

	rows, err := r.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		m, err := scanMember(rows, nil)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

const sqlUpdateMember = `
UPDATE members
SET first_name = $2, last_name = $3, email = $4, joined_at = $5, left_at = $6,
    is_active = $7, individual_goal_hours = $8, oidc_subject = $9, role = $10
WHERE id = $1`

func (r *MemberRepo) Update(ctx context.Context, m *domain.Member) error {
	tag, err := r.pool.Exec(ctx, sqlUpdateMember,
		m.ID, m.FirstName, m.LastName, m.Email, m.JoinedAt, m.LeftAt,
		m.IsActive, m.IndividualGoalHours, m.OIDCSubject, m.Role,
	)
	if err != nil {
		return fmt.Errorf("update member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

const sqlDeactivateMember = `UPDATE members SET is_active = false, left_at = $2 WHERE id = $1`

func (r *MemberRepo) Deactivate(ctx context.Context, id uuid.UUID, leftAt time.Time) error {
	tag, err := r.pool.Exec(ctx, sqlDeactivateMember, id, leftAt)
	if err != nil {
		return fmt.Errorf("deactivate member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

const sqlInsertOIDCLink = `
INSERT INTO oidc_links (id, member_id, provider, subject, linked_at)
VALUES ($1, $2, $3, $4, $5)`

func (r *MemberRepo) LinkOIDC(ctx context.Context, link *domain.OIDCLink) error {
	_, err := r.pool.Exec(ctx, sqlInsertOIDCLink,
		link.ID, link.MemberID, link.Provider, link.Subject, link.LinkedAt,
	)
	if err != nil {
		return fmt.Errorf("insert oidc link: %w", err)
	}
	return nil
}

const sqlGetOIDCLinks = `
SELECT id, member_id, provider, subject, linked_at
FROM oidc_links WHERE member_id = $1 ORDER BY linked_at`

func (r *MemberRepo) GetOIDCLinks(ctx context.Context, memberID uuid.UUID) ([]*domain.OIDCLink, error) {
	rows, err := r.pool.Query(ctx, sqlGetOIDCLinks, memberID)
	if err != nil {
		return nil, fmt.Errorf("get oidc links: %w", err)
	}
	defer rows.Close()

	var links []*domain.OIDCLink
	for rows.Next() {
		var l domain.OIDCLink
		if err := rows.Scan(&l.ID, &l.MemberID, &l.Provider, &l.Subject, &l.LinkedAt); err != nil {
			return nil, err
		}
		links = append(links, &l)
	}
	return links, rows.Err()
}

const sqlCountMembers = `SELECT count(*) FROM members`

func (r *MemberRepo) Count(ctx context.Context) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, sqlCountMembers).Scan(&n); err != nil {
		return 0, fmt.Errorf("count members: %w", err)
	}
	return n, nil
}

// scanMember scans a single member row. notFound, when non-nil, is returned in
// place of pgx.ErrNoRows.
func scanMember(row pgx.Row, notFound error) (*domain.Member, error) {
	var m domain.Member
	err := row.Scan(
		&m.ID, &m.FirstName, &m.LastName, &m.Email, &m.JoinedAt, &m.LeftAt,
		&m.IsActive, &m.IndividualGoalHours, &m.OIDCSubject, &m.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) && notFound != nil {
			return nil, notFound
		}
		return nil, err
	}
	return &m, nil
}
