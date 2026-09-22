package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditEntry is an immutable record of a security or privacy-relevant action.
type AuditEntry struct {
	ID        uuid.UUID       `json:"id"`
	ActorID   *uuid.UUID      `json:"actorId,omitempty"` // nil for system actions
	Action    string          `json:"action"`
	Entity    string          `json:"entity"`
	EntityID  string          `json:"entityId"`
	Before    json.RawMessage `json:"before,omitempty"`
	After     json.RawMessage `json:"after,omitempty"`
	ChangedAt time.Time       `json:"changedAt"`
}

// Common audit action constants.
const (
	AuditActionCreate     = "create"
	AuditActionUpdate     = "update"
	AuditActionDelete     = "delete"
	AuditActionDeactivate = "deactivate"
	AuditActionActivate   = "activate"
	AuditActionLogin      = "login"
	AuditActionLogout     = "logout"
	AuditActionImport     = "import"
	AuditActionExport     = "export"
	AuditActionConfirm    = "confirm"
	AuditActionRegister   = "register"
	AuditActionDeregister = "deregister"
	AuditActionLinkOIDC   = "link_oidc"
	AuditActionManualBook = "manual_booking"
	AuditActionCorrect    = "correct"
	AuditActionCompute    = "compute_billing"
	AuditActionGDPRExport = "gdpr_export"
	AuditActionGDPRDelete = "gdpr_delete"
	AuditActionResend     = "resend"
)

// Common audit entity constants.
const (
	AuditEntityMember        = "member"
	AuditEntityEvent         = "event"
	AuditEntityShift         = "shift"
	AuditEntityRegistration  = "registration"
	AuditEntityHourEntry     = "hour_entry"
	AuditEntitySettings      = "settings"
	AuditEntityBranding      = "branding"
	AuditEntityFeeTier       = "fee_tier"
	AuditEntityClubYear      = "club_year"
	AuditEntityAuth          = "auth"
	AuditEntityEmailTemplate = "email_template"
	AuditEntityEmailLog      = "email_log"
	AuditEntityLogo          = "logo"
)

// NewAuditEntry creates an audit entry with a new UUID and current timestamp.
func NewAuditEntry(actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) (*AuditEntry, error) {
	entry := &AuditEntry{
		ID:        uuid.New(),
		ActorID:   actorID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		ChangedAt: time.Now().UTC(),
	}

	if before != nil {
		b, err := json.Marshal(before)
		if err != nil {
			return nil, err
		}
		entry.Before = b
	}

	if after != nil {
		a, err := json.Marshal(after)
		if err != nil {
			return nil, err
		}
		entry.After = a
	}

	return entry, nil
}
