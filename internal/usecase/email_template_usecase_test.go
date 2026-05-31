package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
)

func newTemplateUC() (*EmailTemplateUsecase, *fakeEmailTemplateRepo, *fakeEmailLogRepo, *fakeEmailService, *fakeAuditRepo) {
	templates := newFakeEmailTemplateRepo()
	logRepo := newFakeEmailLogRepo()
	email := &fakeEmailService{}
	audit := newFakeAuditRepo()
	return NewEmailTemplateUsecase(templates, logRepo, email, audit), templates, logRepo, email, audit
}

func TestUpsertTemplate_ValidatesAndAudits(t *testing.T) {
	uc, templates, _, _, audit := newTemplateUC()
	out, err := uc.UpsertTemplate(context.Background(), uuid.New(), "shift-confirmation", "Hi {{.MemberName}}", "Body {{.ShiftName}}")
	require.NoError(t, err)
	assert.Equal(t, "shift-confirmation", out.Name)
	stored, _ := templates.GetTemplate(context.Background(), "shift-confirmation")
	assert.Equal(t, "Body {{.ShiftName}}", stored.Body)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityEmailTemplate))
}

func TestUpsertTemplate_RejectsInvalidTemplate(t *testing.T) {
	uc, _, _, _, _ := newTemplateUC()
	_, err := uc.UpsertTemplate(context.Background(), uuid.New(), "bad", "ok", "{{.Unclosed")
	assert.Error(t, err)
}

func TestUpsertTemplate_RequiresName(t *testing.T) {
	uc, _, _, _, _ := newTemplateUC()
	_, err := uc.UpsertTemplate(context.Background(), uuid.New(), "  ", "s", "b")
	assert.Error(t, err)
}

func TestListEmailLogAndResend(t *testing.T) {
	uc, _, logRepo, email, audit := newTemplateUC()
	id := uuid.New()
	require.NoError(t, logRepo.Insert(context.Background(), &domain.EmailLogEntry{
		ID: id, To: "x@y.de", Template: "shift-confirmation", Status: domain.EmailLogStatusFailed,
	}))

	list, err := uc.ListEmailLog(context.Background(), 100, 0)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	require.NoError(t, uc.ResendEmail(context.Background(), uuid.New(), id))
	assert.Equal(t, 1, email.countKind("template:shift-confirmation"))
	assert.True(t, audit.has(domain.AuditActionResend, domain.AuditEntityEmailLog))
}
