package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

var _ port.EmailTemplateRepository = (*EmailTemplateRepo)(nil)
var _ port.EmailLogRepository = (*EmailLogRepo)(nil)

// EmailTemplateRepo is a GORM-backed implementation of port.EmailTemplateRepository.
type EmailTemplateRepo struct {
	db *gorm.DB
}

// NewEmailTemplateRepo creates a new EmailTemplateRepo.
func NewEmailTemplateRepo(db *gorm.DB) *EmailTemplateRepo {
	return &EmailTemplateRepo{db: db}
}

func (r *EmailTemplateRepo) ListTemplates(ctx context.Context) ([]*domain.EmailTemplate, error) {
	var models []EmailTemplateModel
	if err := r.db.WithContext(ctx).Order("name").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.EmailTemplate, 0, len(models))
	for _, m := range models {
		out = append(out, emailTemplateModelToDomain(m))
	}
	return out, nil
}

func (r *EmailTemplateRepo) GetTemplate(ctx context.Context, name string) (*domain.EmailTemplate, error) {
	var model EmailTemplateModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("email template %q not found", name)
		}
		return nil, err
	}
	return emailTemplateModelToDomain(model), nil
}

func (r *EmailTemplateRepo) UpsertTemplate(ctx context.Context, t *domain.EmailTemplate) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	var existing EmailTemplateModel
	err := r.db.WithContext(ctx).Where("name = ?", t.Name).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model := EmailTemplateModel{
			ID:      t.ID.String(),
			Name:    t.Name,
			Subject: t.Subject,
			Body:    t.Body,
		}
		return r.db.WithContext(ctx).Create(&model).Error
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"subject": t.Subject,
		"body":    t.Body,
	}).Error
}

func emailTemplateModelToDomain(m EmailTemplateModel) *domain.EmailTemplate {
	return &domain.EmailTemplate{
		ID:      uuid.MustParse(m.ID),
		Name:    m.Name,
		Subject: m.Subject,
		Body:    m.Body,
	}
}

// SeedDefaultTemplates inserts the default email templates if they do not already exist.
func SeedDefaultTemplates(db *gorm.DB) error {
	defaults := []domain.EmailTemplate{
		{
			Name:    domain.EmailTemplateKioskConfirmation,
			Subject: "Kioskbestätigung",
			Body:    "Hallo {{.FirstName}},\n\nDeine Kioskregistrierung für {{.ShiftName}} wurde bestätigt.\n",
		},
		{
			Name:    domain.EmailTemplateShiftConfirmation,
			Subject: "Schichtbestätigung",
			Body:    "Hallo {{.FirstName}},\n\nDeine Anmeldung für {{.ShiftName}} am {{.ShiftDate}} wurde bestätigt.\n",
		},
		{
			Name:    domain.EmailTemplateShiftDeregister,
			Subject: "Schichtabmeldung",
			Body:    "Hallo {{.FirstName}},\n\nDu wurdest von der Schicht {{.ShiftName}} abgemeldet.\n",
		},
		{
			Name:    domain.EmailTemplateReminder1W,
			Subject: "Erinnerung: Schicht in einer Woche",
			Body:    "Hallo {{.FirstName}},\n\nDenk daran: Du hast eine Schicht am {{.ShiftDate}}.\n",
		},
		{
			Name:    domain.EmailTemplateReminder1D,
			Subject: "Erinnerung: Schicht morgen",
			Body:    "Hallo {{.FirstName}},\n\nMorgen hast du eine Schicht: {{.ShiftName}}.\n",
		},
		{
			Name:    domain.EmailTemplateShiftCancelled,
			Subject: "Schicht abgesagt",
			Body:    "Hallo {{.FirstName}},\n\nDie Schicht {{.ShiftName}} wurde leider abgesagt.\n",
		},
		{
			Name:    domain.EmailTemplateYearBilling,
			Subject: "Jahresabrechnung",
			Body:    "Hallo {{.FirstName}},\n\nAnbei deine Jahresabrechnung für {{.YearLabel}}.\n",
		},
		{
			Name:    domain.EmailTemplateMissingHours,
			Subject: "Fehlende Stunden",
			Body:    "Hallo {{.FirstName}},\n\nDir fehlen noch {{.MissingHours}} Stunden.\n",
		},
		{
			Name:    domain.EmailTemplateUnderstaffed,
			Subject: "Schicht unterbesetzt",
			Body:    "Die Schicht {{.ShiftName}} ist unterbesetzt.\n",
		},
	}

	for _, t := range defaults {
		var existing EmailTemplateModel
		err := db.Where("name = ?", t.Name).First(&existing).Error
		if err == nil {
			continue // already exists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		model := EmailTemplateModel{
			ID:      uuid.New().String(),
			Name:    t.Name,
			Subject: t.Subject,
			Body:    t.Body,
		}
		if err := db.Create(&model).Error; err != nil {
			return err
		}
	}
	return nil
}

// --- EmailLogRepo ---

// EmailLogRepo is a GORM-backed implementation of port.EmailLogRepository.
type EmailLogRepo struct {
	db *gorm.DB
}

// NewEmailLogRepo creates a new EmailLogRepo.
func NewEmailLogRepo(db *gorm.DB) *EmailLogRepo {
	return &EmailLogRepo{db: db}
}

func (r *EmailLogRepo) Insert(ctx context.Context, e *domain.EmailLogEntry) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	model := EmailLogModel{
		ID:        e.ID.String(),
		To:        e.To,
		Template:  e.Template,
		Subject:   e.Subject,
		Body:      e.Body,
		Status:    string(e.Status),
		Error:     e.Error,
		CreatedAt: e.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *EmailLogRepo) List(ctx context.Context, limit, offset int) ([]*domain.EmailLogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	var models []EmailLogModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.EmailLogEntry, 0, len(models))
	for _, m := range models {
		out = append(out, emailLogModelToDomain(m))
	}
	return out, nil
}

func (r *EmailLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailLogEntry, error) {
	var model EmailLogModel
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrEmailLogNotFound
		}
		return nil, err
	}
	return emailLogModelToDomain(model), nil
}

func emailLogModelToDomain(m EmailLogModel) *domain.EmailLogEntry {
	return &domain.EmailLogEntry{
		ID:        uuid.MustParse(m.ID),
		To:        m.To,
		Template:  m.Template,
		Subject:   m.Subject,
		Body:      m.Body,
		Status:    domain.EmailLogStatus(m.Status),
		Error:     m.Error,
		CreatedAt: m.CreatedAt,
	}
}
