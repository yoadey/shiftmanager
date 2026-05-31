package domain

import (
	"time"

	"github.com/google/uuid"
)

// Standard email template names. These match the rows seeded in migration 003
// and are the keys used to look up overridable subject/body from the database.
const (
	EmailTemplateKioskConfirmation = "kiosk-confirmation"
	EmailTemplateShiftConfirmation = "shift-confirmation"
	EmailTemplateShiftDeregister   = "shift-deregistration"
	EmailTemplateReminder1W        = "reminder-1w"
	EmailTemplateReminder1D        = "reminder-1d"
	EmailTemplateShiftCancelled    = "shift-cancelled"
	EmailTemplateYearBilling       = "year-billing"
	EmailTemplateMissingHours      = "missing-hours-warning"
	EmailTemplateUnderstaffed      = "shift-understaffed"
)

// EmailTemplate is an overridable transactional email template. Subject and body
// are Go text/template sources rendered with flattened placeholder data.
type EmailTemplate struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Subject string    `json:"subject"`
	Body    string    `json:"body"`
}

// EmailLogStatus is the outcome of a single email send attempt.
type EmailLogStatus string

const (
	EmailLogStatusSent   EmailLogStatus = "sent"
	EmailLogStatusFailed EmailLogStatus = "failed"
)

// EmailLogEntry records a single email send attempt for auditing and resend (N-004).
type EmailLogEntry struct {
	ID        uuid.UUID      `json:"id"`
	To        string         `json:"to"`
	Template  string         `json:"template"`
	Subject   string         `json:"subject"`
	Body      string         `json:"body"`
	Status    EmailLogStatus `json:"status"`
	Error     string         `json:"error"`
	CreatedAt time.Time      `json:"createdAt"`
}
