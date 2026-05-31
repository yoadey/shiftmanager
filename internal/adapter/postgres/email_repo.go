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

// EmailTemplateRepo is a pgxpool-backed implementation of
// port.EmailTemplateRepository.
type EmailTemplateRepo struct {
	pool *pgxpool.Pool
}

var _ port.EmailTemplateRepository = (*EmailTemplateRepo)(nil)

// NewEmailTemplateRepo creates a new EmailTemplateRepo.
func NewEmailTemplateRepo(pool *pgxpool.Pool) *EmailTemplateRepo {
	return &EmailTemplateRepo{pool: pool}
}

const sqlListTemplates = `SELECT id, name, subject, body FROM email_templates ORDER BY name`

func (r *EmailTemplateRepo) ListTemplates(ctx context.Context) ([]*domain.EmailTemplate, error) {
	rows, err := r.pool.Query(ctx, sqlListTemplates)
	if err != nil {
		return nil, fmt.Errorf("list email templates: %w", err)
	}
	defer rows.Close()

	var out []*domain.EmailTemplate
	for rows.Next() {
		var t domain.EmailTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Subject, &t.Body); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

const sqlGetTemplate = `SELECT id, name, subject, body FROM email_templates WHERE name = $1`

func (r *EmailTemplateRepo) GetTemplate(ctx context.Context, name string) (*domain.EmailTemplate, error) {
	var t domain.EmailTemplate
	err := r.pool.QueryRow(ctx, sqlGetTemplate, name).Scan(&t.ID, &t.Name, &t.Subject, &t.Body)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("email template %q not found", name)
		}
		return nil, err
	}
	return &t, nil
}

const sqlUpsertTemplate = `
INSERT INTO email_templates (id, name, subject, body)
VALUES ($1, $2, $3, $4)
ON CONFLICT (name) DO UPDATE SET subject = EXCLUDED.subject, body = EXCLUDED.body`

func (r *EmailTemplateRepo) UpsertTemplate(ctx context.Context, t *domain.EmailTemplate) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, sqlUpsertTemplate, t.ID, t.Name, t.Subject, t.Body)
	if err != nil {
		return fmt.Errorf("upsert email template: %w", err)
	}
	return nil
}

// EmailLogRepo is a pgxpool-backed implementation of port.EmailLogRepository.
type EmailLogRepo struct {
	pool *pgxpool.Pool
}

var _ port.EmailLogRepository = (*EmailLogRepo)(nil)

// NewEmailLogRepo creates a new EmailLogRepo.
func NewEmailLogRepo(pool *pgxpool.Pool) *EmailLogRepo {
	return &EmailLogRepo{pool: pool}
}

const sqlInsertEmailLog = `
INSERT INTO email_log (id, to_address, template, subject, body, status, error, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

func (r *EmailLogRepo) Insert(ctx context.Context, e *domain.EmailLogEntry) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, sqlInsertEmailLog,
		e.ID, e.To, e.Template, e.Subject, e.Body, string(e.Status), e.Error, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert email log: %w", err)
	}
	return nil
}

const emailLogColumns = `id, to_address, template, subject, body, status, error, created_at`

const sqlListEmailLog = `SELECT ` + emailLogColumns + ` FROM email_log ORDER BY created_at DESC LIMIT $1 OFFSET $2`

func (r *EmailLogRepo) List(ctx context.Context, limit, offset int) ([]*domain.EmailLogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, sqlListEmailLog, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list email log: %w", err)
	}
	defer rows.Close()

	var out []*domain.EmailLogEntry
	for rows.Next() {
		e, err := scanEmailLog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

const sqlGetEmailLogByID = `SELECT ` + emailLogColumns + ` FROM email_log WHERE id = $1`

func (r *EmailLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailLogEntry, error) {
	row := r.pool.QueryRow(ctx, sqlGetEmailLogByID, id)
	return scanEmailLog(row)
}

func scanEmailLog(row pgx.Row) (*domain.EmailLogEntry, error) {
	var e domain.EmailLogEntry
	var status string
	err := row.Scan(&e.ID, &e.To, &e.Template, &e.Subject, &e.Body, &status, &e.Error, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("email log entry not found")
		}
		return nil, err
	}
	e.Status = domain.EmailLogStatus(status)
	return &e, nil
}
