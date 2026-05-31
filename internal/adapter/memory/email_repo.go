package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

var _ port.EmailTemplateRepository = (*EmailRepo)(nil)
var _ port.EmailLogRepository = (*EmailRepo)(nil)

// EmailRepo is a combined in-memory implementation of port.EmailTemplateRepository
// and port.EmailLogRepository.
type EmailRepo struct {
	mu        sync.RWMutex
	templates map[string]*domain.EmailTemplate
	log       []*domain.EmailLogEntry
}

func NewEmailRepo() *EmailRepo {
	r := &EmailRepo{
		templates: make(map[string]*domain.EmailTemplate),
	}
	r.seedTemplates()
	return r
}

func (r *EmailRepo) seedTemplates() {
	defaults := []struct {
		name    string
		subject string
		body    string
	}{
		{domain.EmailTemplateKioskConfirmation, "Kiosk-Bestätigung", "Deine Kiosk-Anmeldung wurde bestätigt."},
		{domain.EmailTemplateShiftConfirmation, "Schichtbestätigung", "Deine Schichtanmeldung wurde bestätigt."},
		{domain.EmailTemplateShiftDeregister, "Schichtabmeldung", "Du wurdest von der Schicht abgemeldet."},
		{domain.EmailTemplateReminder1W, "Erinnerung: Schicht in 1 Woche", "Deine Schicht findet in einer Woche statt."},
		{domain.EmailTemplateReminder1D, "Erinnerung: Schicht morgen", "Deine Schicht findet morgen statt."},
		{domain.EmailTemplateShiftCancelled, "Schicht abgesagt", "Die Schicht wurde abgesagt."},
		{domain.EmailTemplateYearBilling, "Jahresabrechnung", "Anbei deine Jahresabrechnung."},
		{domain.EmailTemplateMissingHours, "Fehlende Stunden", "Dir fehlen noch Stunden bis zum Jahresende."},
		{domain.EmailTemplateUnderstaffed, "Schicht unterbesetzt", "Eine Schicht ist noch unterbesetzt."},
	}
	for _, d := range defaults {
		r.templates[d.name] = &domain.EmailTemplate{
			ID:      uuid.New(),
			Name:    d.name,
			Subject: d.subject,
			Body:    d.body,
		}
	}
}

func copyTemplate(t *domain.EmailTemplate) *domain.EmailTemplate {
	c := *t
	return &c
}

func (r *EmailRepo) ListTemplates(_ context.Context) ([]*domain.EmailTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.EmailTemplate, 0, len(r.templates))
	for _, t := range r.templates {
		result = append(result, copyTemplate(t))
	}
	return result, nil
}

func (r *EmailRepo) GetTemplate(_ context.Context, name string) (*domain.EmailTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.templates[name]
	if !ok {
		return nil, domain.ErrEmailTemplateNotFound
	}
	return copyTemplate(t), nil
}

func (r *EmailRepo) UpsertTemplate(_ context.Context, t *domain.EmailTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.templates[t.Name] = copyTemplate(t)
	return nil
}

func copyLogEntry(e *domain.EmailLogEntry) *domain.EmailLogEntry {
	c := *e
	return &c
}

func (r *EmailRepo) Insert(_ context.Context, e *domain.EmailLogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log = append(r.log, copyLogEntry(e))
	return nil
}

func (r *EmailRepo) List(_ context.Context, limit, offset int) ([]*domain.EmailLogEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.EmailLogEntry, len(r.log))
	for i, e := range r.log {
		result[i] = copyLogEntry(e)
	}
	if offset > 0 {
		if offset >= len(result) {
			return []*domain.EmailLogEntry{}, nil
		}
		result = result[offset:]
	}
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

func (r *EmailRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.EmailLogEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.log {
		if e.ID == id {
			return copyLogEntry(e), nil
		}
	}
	return nil, domain.ErrEmailLogNotFound
}
