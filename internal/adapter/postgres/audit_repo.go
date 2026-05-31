package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// AuditRepo is a pgxpool-backed implementation of port.AuditRepository. The
// audit log is append-only; there are no update or delete operations.
type AuditRepo struct {
	pool *pgxpool.Pool
}

var _ port.AuditRepository = (*AuditRepo)(nil)

// NewAuditRepo creates a new AuditRepo.
func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

const sqlInsertAudit = `
INSERT INTO audit_log (id, actor_id, action, entity, entity_id, before_state, after_state, changed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

func (r *AuditRepo) Insert(ctx context.Context, e *domain.AuditEntry) error {
	var before, after any
	if len(e.Before) > 0 {
		before = []byte(e.Before)
	}
	if len(e.After) > 0 {
		after = []byte(e.After)
	}
	_, err := r.pool.Exec(ctx, sqlInsertAudit,
		e.ID, e.ActorID, e.Action, e.Entity, e.EntityID, before, after, e.ChangedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit entry: %w", err)
	}
	return nil
}

func (r *AuditRepo) List(ctx context.Context, filter port.AuditFilter) ([]*domain.AuditEntry, error) {
	var sb strings.Builder
	sb.WriteString(`SELECT id, actor_id, action, entity, entity_id, before_state, after_state, changed_at FROM audit_log WHERE 1=1`)
	args := []any{}
	idx := 1

	if filter.ActorID != nil {
		sb.WriteString(fmt.Sprintf(" AND actor_id = $%d", idx))
		args = append(args, *filter.ActorID)
		idx++
	}
	if filter.Entity != "" {
		sb.WriteString(fmt.Sprintf(" AND entity = $%d", idx))
		args = append(args, filter.Entity)
		idx++
	}
	if filter.EntityID != "" {
		sb.WriteString(fmt.Sprintf(" AND entity_id = $%d", idx))
		args = append(args, filter.EntityID)
		idx++
	}
	if filter.From != nil {
		sb.WriteString(fmt.Sprintf(" AND changed_at >= $%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		sb.WriteString(fmt.Sprintf(" AND changed_at <= $%d", idx))
		args = append(args, *filter.To)
		idx++
	}

	sb.WriteString(" ORDER BY changed_at DESC")

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	sb.WriteString(fmt.Sprintf(" LIMIT $%d", idx))
	args = append(args, limit)
	idx++

	if filter.Offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET $%d", idx))
		args = append(args, filter.Offset)
		idx++
	}

	rows, err := r.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list audit entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		var before, after []byte
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.Entity, &e.EntityID, &before, &after, &e.ChangedAt); err != nil {
			return nil, err
		}
		e.Before = before
		e.After = after
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}
