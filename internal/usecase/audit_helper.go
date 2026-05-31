package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// writeAuditEntry is a shared best-effort audit-log helper used by the newer
// usecases. Existing usecases keep their own (equivalent) writeAudit methods.
func writeAuditEntry(ctx context.Context, audit port.AuditRepository, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return audit.Insert(ctx, entry)
}
