package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// HourRepo is a pgxpool-backed implementation of port.HourRepository covering
// hour entries, club years and per-member hour targets.
type HourRepo struct {
	pool *pgxpool.Pool
}

var _ port.HourRepository = (*HourRepo)(nil)

// NewHourRepo creates a new HourRepo.
func NewHourRepo(pool *pgxpool.Pool) *HourRepo {
	return &HourRepo{pool: pool}
}

// --- Hour entries ---

const hourEntryColumns = `id, member_id, shift_id, club_year_id, hours, type, status, booked_by, description, created_at`

const sqlInsertHourEntry = `
INSERT INTO hour_entries (id, member_id, shift_id, club_year_id, hours, type, status, booked_by, description, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (r *HourRepo) CreateEntry(ctx context.Context, e *domain.HourEntry) error {
	_, err := r.pool.Exec(ctx, sqlInsertHourEntry,
		e.ID, e.MemberID, e.ShiftID, e.ClubYearID, e.Hours, e.Type, e.Status,
		e.BookedBy, e.Description, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert hour entry: %w", err)
	}
	return nil
}

const sqlGetHourEntryByID = `SELECT ` + hourEntryColumns + ` FROM hour_entries WHERE id = $1`

func (r *HourRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.HourEntry, error) {
	row := r.pool.QueryRow(ctx, sqlGetHourEntryByID, id)
	return scanHourEntry(row)
}

const sqlFindEntriesByMemberAndYear = `
SELECT ` + hourEntryColumns + `
FROM hour_entries WHERE member_id = $1 AND club_year_id = $2 ORDER BY created_at`

func (r *HourRepo) FindEntriesByMemberAndYear(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	return r.queryEntries(ctx, sqlFindEntriesByMemberAndYear, memberID, clubYearID)
}

const sqlFindEntriesByYear = `
SELECT ` + hourEntryColumns + `
FROM hour_entries WHERE club_year_id = $1 ORDER BY member_id, created_at`

func (r *HourRepo) FindEntriesByYear(ctx context.Context, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	return r.queryEntries(ctx, sqlFindEntriesByYear, clubYearID)
}

const sqlUpdateHourEntry = `
UPDATE hour_entries
SET member_id = $2, shift_id = $3, club_year_id = $4, hours = $5, type = $6,
    status = $7, booked_by = $8, description = $9
WHERE id = $1`

func (r *HourRepo) UpdateEntry(ctx context.Context, e *domain.HourEntry) error {
	tag, err := r.pool.Exec(ctx, sqlUpdateHourEntry,
		e.ID, e.MemberID, e.ShiftID, e.ClubYearID, e.Hours, e.Type, e.Status,
		e.BookedBy, e.Description,
	)
	if err != nil {
		return fmt.Errorf("update hour entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrHourEntryNotFound
	}
	return nil
}

const sqlDeleteHourEntry = `DELETE FROM hour_entries WHERE id = $1`

func (r *HourRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, sqlDeleteHourEntry, id)
	if err != nil {
		return fmt.Errorf("delete hour entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrHourEntryNotFound
	}
	return nil
}

func (r *HourRepo) queryEntries(ctx context.Context, sql string, args ...any) ([]*domain.HourEntry, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query hour entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.HourEntry
	for rows.Next() {
		e, err := scanHourEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func scanHourEntry(row pgx.Row) (*domain.HourEntry, error) {
	var e domain.HourEntry
	err := row.Scan(
		&e.ID, &e.MemberID, &e.ShiftID, &e.ClubYearID, &e.Hours, &e.Type,
		&e.Status, &e.BookedBy, &e.Description, &e.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHourEntryNotFound
		}
		return nil, err
	}
	return &e, nil
}

// --- Club years ---

const clubYearColumns = `id, label, start_date, end_date, default_target_hours, is_active`

const sqlInsertClubYear = `
INSERT INTO club_years (id, label, start_date, end_date, default_target_hours, is_active)
VALUES ($1, $2, $3, $4, $5, $6)`

func (r *HourRepo) CreateClubYear(ctx context.Context, y *domain.ClubYear) error {
	_, err := r.pool.Exec(ctx, sqlInsertClubYear,
		y.ID, y.Label, y.StartDate, y.EndDate, y.DefaultTargetHours, y.IsActive,
	)
	if err != nil {
		return fmt.Errorf("insert club year: %w", err)
	}
	return nil
}

const sqlGetActiveClubYear = `SELECT ` + clubYearColumns + ` FROM club_years WHERE is_active = true ORDER BY start_date DESC LIMIT 1`

func (r *HourRepo) GetActiveClubYear(ctx context.Context) (*domain.ClubYear, error) {
	row := r.pool.QueryRow(ctx, sqlGetActiveClubYear)
	return scanClubYear(row)
}

const sqlGetClubYearByID = `SELECT ` + clubYearColumns + ` FROM club_years WHERE id = $1`

func (r *HourRepo) GetClubYearByID(ctx context.Context, id uuid.UUID) (*domain.ClubYear, error) {
	row := r.pool.QueryRow(ctx, sqlGetClubYearByID, id)
	return scanClubYear(row)
}

const sqlListClubYears = `SELECT ` + clubYearColumns + ` FROM club_years ORDER BY start_date DESC`

func (r *HourRepo) ListClubYears(ctx context.Context) ([]*domain.ClubYear, error) {
	rows, err := r.pool.Query(ctx, sqlListClubYears)
	if err != nil {
		return nil, fmt.Errorf("list club years: %w", err)
	}
	defer rows.Close()

	var years []*domain.ClubYear
	for rows.Next() {
		y, err := scanClubYear(rows)
		if err != nil {
			return nil, err
		}
		years = append(years, y)
	}
	return years, rows.Err()
}

func scanClubYear(row pgx.Row) (*domain.ClubYear, error) {
	var y domain.ClubYear
	err := row.Scan(&y.ID, &y.Label, &y.StartDate, &y.EndDate, &y.DefaultTargetHours, &y.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrClubYearNotFound
		}
		return nil, err
	}
	return &y, nil
}

// --- Hour targets ---

const sqlGetHourTarget = `
SELECT id, member_id, club_year_id, target_hours
FROM hour_targets WHERE member_id = $1 AND club_year_id = $2`

func (r *HourRepo) GetHourTarget(ctx context.Context, memberID, clubYearID uuid.UUID) (*domain.HourTarget, error) {
	var t domain.HourTarget
	err := r.pool.QueryRow(ctx, sqlGetHourTarget, memberID, clubYearID).
		Scan(&t.ID, &t.MemberID, &t.ClubYearID, &t.TargetHours)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHourEntryNotFound
		}
		return nil, err
	}
	return &t, nil
}

const sqlUpsertHourTarget = `
INSERT INTO hour_targets (id, member_id, club_year_id, target_hours)
VALUES ($1, $2, $3, $4)
ON CONFLICT (member_id, club_year_id)
DO UPDATE SET target_hours = EXCLUDED.target_hours`

func (r *HourRepo) UpsertHourTarget(ctx context.Context, t *domain.HourTarget) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, sqlUpsertHourTarget, t.ID, t.MemberID, t.ClubYearID, t.TargetHours)
	if err != nil {
		return fmt.Errorf("upsert hour target: %w", err)
	}
	return nil
}
