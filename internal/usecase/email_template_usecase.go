package usecase

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// EmailTemplateUsecase handles admin CRUD for overridable email templates
// (Section 4) and email-log inspection/resend (N-004).
type EmailTemplateUsecase struct {
	templates port.EmailTemplateRepository
	emailLog  port.EmailLogRepository
	email     port.EmailService
	audit     port.AuditRepository
}

// NewEmailTemplateUsecase creates a new EmailTemplateUsecase. emailLog and email
// may be nil if the resend/log features are not wired.
func NewEmailTemplateUsecase(
	templates port.EmailTemplateRepository,
	emailLog port.EmailLogRepository,
	email port.EmailService,
	audit port.AuditRepository,
) *EmailTemplateUsecase {
	return &EmailTemplateUsecase{templates: templates, emailLog: emailLog, email: email, audit: audit}
}

// ListTemplates returns all email templates.
func (uc *EmailTemplateUsecase) ListTemplates(ctx context.Context) ([]*domain.EmailTemplate, error) {
	return uc.templates.ListTemplates(ctx)
}

// GetTemplate returns a single template by name.
func (uc *EmailTemplateUsecase) GetTemplate(ctx context.Context, name string) (*domain.EmailTemplate, error) {
	return uc.templates.GetTemplate(ctx, name)
}

// UpsertTemplate validates the template sources are parseable and persists them.
func (uc *EmailTemplateUsecase) UpsertTemplate(ctx context.Context, actorID uuid.UUID, name, subject, body string) (*domain.EmailTemplate, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	// Reject syntactically invalid templates so we never persist a body that
	// fails to render at send time.
	if _, err := template.New("subject").Parse(subject); err != nil {
		return nil, fmt.Errorf("invalid subject template: %w", err)
	}
	if _, err := template.New("body").Parse(body); err != nil {
		return nil, fmt.Errorf("invalid body template: %w", err)
	}

	t := &domain.EmailTemplate{Name: name, Subject: subject, Body: body}
	if err := uc.templates.UpsertTemplate(ctx, t); err != nil {
		return nil, fmt.Errorf("upsert template: %w", err)
	}

	aid := actorID
	_ = writeAuditEntry(ctx, uc.audit, &aid, domain.AuditActionUpdate, domain.AuditEntityEmailTemplate, name, nil, t)
	return t, nil
}

// ListEmailLog returns recent email-log entries (N-004).
func (uc *EmailTemplateUsecase) ListEmailLog(ctx context.Context, limit, offset int) ([]*domain.EmailLogEntry, error) {
	if uc.emailLog == nil {
		return nil, fmt.Errorf("email log not available")
	}
	return uc.emailLog.List(ctx, limit, offset)
}

// ResendEmail re-sends a previously logged email by replaying its template (N-004).
func (uc *EmailTemplateUsecase) ResendEmail(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	if uc.emailLog == nil || uc.email == nil {
		return fmt.Errorf("email resend not available")
	}
	entry, err := uc.emailLog.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// Re-render the stored template; the original placeholder values are not
	// retained, so we resend using the template's current content. The body that
	// was actually delivered is preserved in the log for reference.
	if err := uc.email.SendByTemplate(ctx, entry.To, entry.Template, map[string]any{}); err != nil {
		return fmt.Errorf("resend email: %w", err)
	}

	aid := actorID
	_ = writeAuditEntry(ctx, uc.audit, &aid, domain.AuditActionResend, domain.AuditEntityEmailLog, id.String(), nil, map[string]string{"to": entry.To, "template": entry.Template})
	return nil
}
