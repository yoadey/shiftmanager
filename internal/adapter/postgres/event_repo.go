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

// EventRepo is a pgxpool-backed implementation of port.EventRepository.
type EventRepo struct {
	pool *pgxpool.Pool
}

var _ port.EventRepository = (*EventRepo)(nil)

// NewEventRepo creates a new EventRepo.
func NewEventRepo(pool *pgxpool.Pool) *EventRepo {
	return &EventRepo{pool: pool}
}

const eventColumns = `id, name, description, location, category, start_date, end_date, status, visibility, created_at, updated_at`

const sqlInsertEvent = `
INSERT INTO events (id, name, description, location, category, start_date, end_date, status, visibility, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

func (r *EventRepo) Create(ctx context.Context, e *domain.Event) error {
	_, err := r.pool.Exec(ctx, sqlInsertEvent,
		e.ID, e.Name, e.Description, e.Location, e.Category, e.StartDate, e.EndDate,
		e.Status, e.Visibility, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

const sqlGetEventByID = `SELECT ` + eventColumns + ` FROM events WHERE id = $1`

func (r *EventRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx, sqlGetEventByID, id)
	return scanEvent(row)
}

func (r *EventRepo) List(ctx context.Context, filter port.EventFilter) ([]*domain.Event, error) {
	var sb strings.Builder
	sb.WriteString(`SELECT ` + eventColumns + ` FROM events WHERE 1=1`)
	args := []any{}
	idx := 1

	if filter.Status != nil {
		sb.WriteString(fmt.Sprintf(" AND status = $%d", idx))
		args = append(args, *filter.Status)
		idx++
	}
	if filter.Visibility != nil {
		sb.WriteString(fmt.Sprintf(" AND visibility = $%d", idx))
		args = append(args, *filter.Visibility)
		idx++
	}
	if filter.FromDate != nil {
		sb.WriteString(fmt.Sprintf(" AND end_date >= $%d", idx))
		args = append(args, *filter.FromDate)
		idx++
	}
	if filter.ToDate != nil {
		sb.WriteString(fmt.Sprintf(" AND start_date <= $%d", idx))
		args = append(args, *filter.ToDate)
		idx++
	}

	sb.WriteString(" ORDER BY start_date DESC")

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
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

const sqlUpdateEvent = `
UPDATE events
SET name = $2, description = $3, location = $4, category = $5, start_date = $6,
    end_date = $7, status = $8, visibility = $9, updated_at = $10
WHERE id = $1`

func (r *EventRepo) Update(ctx context.Context, e *domain.Event) error {
	tag, err := r.pool.Exec(ctx, sqlUpdateEvent,
		e.ID, e.Name, e.Description, e.Location, e.Category, e.StartDate, e.EndDate,
		e.Status, e.Visibility, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

const sqlDeleteEvent = `DELETE FROM events WHERE id = $1`

func (r *EventRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, sqlDeleteEvent, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

const sqlUpdateEventStatus = `UPDATE events SET status = $2, updated_at = now() WHERE id = $1`

func (r *EventRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EventStatus) error {
	tag, err := r.pool.Exec(ctx, sqlUpdateEventStatus, id, status)
	if err != nil {
		return fmt.Errorf("update event status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

func scanEvent(row pgx.Row) (*domain.Event, error) {
	var e domain.Event
	err := row.Scan(
		&e.ID, &e.Name, &e.Description, &e.Location, &e.Category, &e.StartDate,
		&e.EndDate, &e.Status, &e.Visibility, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventNotFound
		}
		return nil, err
	}
	return &e, nil
}

// --- ShiftRepository ---

// ShiftRepo is a pgxpool-backed implementation of port.ShiftRepository.
type ShiftRepo struct {
	pool *pgxpool.Pool
}

var _ port.ShiftRepository = (*ShiftRepo)(nil)

// NewShiftRepo creates a new ShiftRepo.
func NewShiftRepo(pool *pgxpool.Pool) *ShiftRepo {
	return &ShiftRepo{pool: pool}
}

const shiftColumns = `id, event_id, name, start_at, end_at, min_helpers, max_helpers, required_qualification, shift_date`

const sqlInsertShift = `
INSERT INTO shifts (id, event_id, name, start_at, end_at, min_helpers, max_helpers, required_qualification, shift_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

func (r *ShiftRepo) Create(ctx context.Context, s *domain.Shift) error {
	_, err := r.pool.Exec(ctx, sqlInsertShift,
		s.ID, s.EventID, s.Name, s.StartAt, s.EndAt, s.MinHelpers, s.MaxHelpers,
		s.RequiredQualification, s.Date,
	)
	if err != nil {
		return fmt.Errorf("insert shift: %w", err)
	}
	return nil
}

const sqlGetShiftByID = `SELECT ` + shiftColumns + ` FROM shifts WHERE id = $1`

func (r *ShiftRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shift, error) {
	row := r.pool.QueryRow(ctx, sqlGetShiftByID, id)
	return scanShift(row)
}

const sqlFindShiftsByEvent = `SELECT ` + shiftColumns + ` FROM shifts WHERE event_id = $1 ORDER BY start_at`

func (r *ShiftRepo) FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain.Shift, error) {
	rows, err := r.pool.Query(ctx, sqlFindShiftsByEvent, eventID)
	if err != nil {
		return nil, fmt.Errorf("find shifts by event: %w", err)
	}
	defer rows.Close()

	var shifts []*domain.Shift
	for rows.Next() {
		s, err := scanShift(rows)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}
	return shifts, rows.Err()
}

const sqlFindShiftsStartingBetween = `SELECT ` + shiftColumns + ` FROM shifts WHERE start_at >= $1 AND start_at < $2 ORDER BY start_at`

func (r *ShiftRepo) FindShiftsStartingBetween(ctx context.Context, from, to time.Time) ([]*domain.Shift, error) {
	rows, err := r.pool.Query(ctx, sqlFindShiftsStartingBetween, from, to)
	if err != nil {
		return nil, fmt.Errorf("find shifts starting between: %w", err)
	}
	defer rows.Close()

	var shifts []*domain.Shift
	for rows.Next() {
		s, err := scanShift(rows)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}
	return shifts, rows.Err()
}

const sqlUpdateShift = `
UPDATE shifts
SET name = $2, start_at = $3, end_at = $4, min_helpers = $5, max_helpers = $6,
    required_qualification = $7, shift_date = $8
WHERE id = $1`

func (r *ShiftRepo) Update(ctx context.Context, s *domain.Shift) error {
	tag, err := r.pool.Exec(ctx, sqlUpdateShift,
		s.ID, s.Name, s.StartAt, s.EndAt, s.MinHelpers, s.MaxHelpers,
		s.RequiredQualification, s.Date,
	)
	if err != nil {
		return fmt.Errorf("update shift: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrShiftNotFound
	}
	return nil
}

const sqlDeleteShift = `DELETE FROM shifts WHERE id = $1`

func (r *ShiftRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, sqlDeleteShift, id)
	if err != nil {
		return fmt.Errorf("delete shift: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrShiftNotFound
	}
	return nil
}

func scanShift(row pgx.Row) (*domain.Shift, error) {
	var s domain.Shift
	err := row.Scan(
		&s.ID, &s.EventID, &s.Name, &s.StartAt, &s.EndAt, &s.MinHelpers,
		&s.MaxHelpers, &s.RequiredQualification, &s.Date,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrShiftNotFound
		}
		return nil, err
	}
	return &s, nil
}

// --- RegistrationRepository ---

// RegistrationRepo is a pgxpool-backed implementation of port.RegistrationRepository.
type RegistrationRepo struct {
	pool *pgxpool.Pool
}

var _ port.RegistrationRepository = (*RegistrationRepo)(nil)

// NewRegistrationRepo creates a new RegistrationRepo.
func NewRegistrationRepo(pool *pgxpool.Pool) *RegistrationRepo {
	return &RegistrationRepo{pool: pool}
}

const registrationColumns = `id, shift_id, member_id, guest_email, state, comment, reserved_until, booked_hours, confirmation_token, created_at`

const sqlInsertRegistration = `
INSERT INTO registrations (id, shift_id, member_id, guest_email, state, comment, reserved_until, booked_hours, confirmation_token, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (r *RegistrationRepo) Create(ctx context.Context, reg *domain.Registration) error {
	_, err := r.pool.Exec(ctx, sqlInsertRegistration,
		reg.ID, reg.ShiftID, reg.MemberID, reg.GuestEmail, reg.State, reg.Comment,
		reg.ReservedUntil, reg.BookedHours, reg.ConfirmationToken, reg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert registration: %w", err)
	}
	return nil
}

const sqlGetRegByID = `SELECT ` + registrationColumns + ` FROM registrations WHERE id = $1`

func (r *RegistrationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Registration, error) {
	row := r.pool.QueryRow(ctx, sqlGetRegByID, id)
	return scanRegistration(row)
}

const sqlGetRegByToken = `SELECT ` + registrationColumns + ` FROM registrations WHERE confirmation_token = $1`

func (r *RegistrationRepo) GetByToken(ctx context.Context, token uuid.UUID) (*domain.Registration, error) {
	row := r.pool.QueryRow(ctx, sqlGetRegByToken, token)
	return scanRegistration(row)
}

const sqlFindRegByShift = `SELECT ` + registrationColumns + ` FROM registrations WHERE shift_id = $1 ORDER BY created_at`

func (r *RegistrationRepo) FindByShiftID(ctx context.Context, shiftID uuid.UUID) ([]*domain.Registration, error) {
	return r.queryRegistrations(ctx, sqlFindRegByShift, shiftID)
}

const sqlFindRegByMember = `SELECT ` + registrationColumns + ` FROM registrations WHERE member_id = $1 ORDER BY created_at DESC`

func (r *RegistrationRepo) FindByMemberID(ctx context.Context, memberID uuid.UUID) ([]*domain.Registration, error) {
	return r.queryRegistrations(ctx, sqlFindRegByMember, memberID)
}

const sqlFindRegByMemberAndShift = `SELECT ` + registrationColumns + ` FROM registrations WHERE member_id = $1 AND shift_id = $2`

func (r *RegistrationRepo) FindByMemberAndShift(ctx context.Context, memberID, shiftID uuid.UUID) (*domain.Registration, error) {
	row := r.pool.QueryRow(ctx, sqlFindRegByMemberAndShift, memberID, shiftID)
	return scanRegistration(row)
}

const sqlFindRegByGuestAndShift = `SELECT ` + registrationColumns + ` FROM registrations WHERE guest_email = $1 AND shift_id = $2`

func (r *RegistrationRepo) FindByGuestEmailAndShift(ctx context.Context, guestEmail string, shiftID uuid.UUID) (*domain.Registration, error) {
	row := r.pool.QueryRow(ctx, sqlFindRegByGuestAndShift, strings.ToLower(guestEmail), shiftID)
	return scanRegistration(row)
}

const sqlCountActiveByShift = `
SELECT count(*) FROM registrations
WHERE shift_id = $1 AND state IN ('registered', 'confirmed')`

func (r *RegistrationRepo) CountActiveByShift(ctx context.Context, shiftID uuid.UUID) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, sqlCountActiveByShift, shiftID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count active registrations: %w", err)
	}
	return n, nil
}

const sqlUpdateRegistration = `
UPDATE registrations
SET shift_id = $2, member_id = $3, guest_email = $4, state = $5, comment = $6,
    reserved_until = $7, booked_hours = $8, confirmation_token = $9
WHERE id = $1`

func (r *RegistrationRepo) Update(ctx context.Context, reg *domain.Registration) error {
	tag, err := r.pool.Exec(ctx, sqlUpdateRegistration,
		reg.ID, reg.ShiftID, reg.MemberID, reg.GuestEmail, reg.State, reg.Comment,
		reg.ReservedUntil, reg.BookedHours, reg.ConfirmationToken,
	)
	if err != nil {
		return fmt.Errorf("update registration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRegistrationNotFound
	}
	return nil
}

const sqlDeleteRegistration = `DELETE FROM registrations WHERE id = $1`

func (r *RegistrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, sqlDeleteRegistration, id)
	if err != nil {
		return fmt.Errorf("delete registration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRegistrationNotFound
	}
	return nil
}

const sqlListExpiredReservations = `
SELECT ` + registrationColumns + `
FROM registrations
WHERE state = 'reserved' AND reserved_until IS NOT NULL AND reserved_until < $1`

func (r *RegistrationRepo) ListUnconfirmedExpiredReservations(ctx context.Context, before time.Time) ([]*domain.Registration, error) {
	return r.queryRegistrations(ctx, sqlListExpiredReservations, before)
}

func (r *RegistrationRepo) queryRegistrations(ctx context.Context, sql string, args ...any) ([]*domain.Registration, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query registrations: %w", err)
	}
	defer rows.Close()

	var regs []*domain.Registration
	for rows.Next() {
		reg, err := scanRegistration(rows)
		if err != nil {
			return nil, err
		}
		regs = append(regs, reg)
	}
	return regs, rows.Err()
}

func scanRegistration(row pgx.Row) (*domain.Registration, error) {
	var reg domain.Registration
	err := row.Scan(
		&reg.ID, &reg.ShiftID, &reg.MemberID, &reg.GuestEmail, &reg.State,
		&reg.Comment, &reg.ReservedUntil, &reg.BookedHours, &reg.ConfirmationToken, &reg.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRegistrationNotFound
		}
		return nil, err
	}
	return &reg, nil
}
