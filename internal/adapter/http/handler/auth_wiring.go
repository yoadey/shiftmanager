package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// BuildAuthHandler constructs an AuthHandler from the real repository
// interfaces. The AuthHandler's own dependencies (memberGetter / auditWriter)
// are unexported interfaces, so this constructor lives in the handler package
// and adapts the port repositories to them.
func BuildAuthHandler(
	oidc port.OIDCService,
	members port.MemberRepository,
	audit port.AuditRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
	loginRedirect string,
	bootstrapAdminEmail string,
) *AuthHandler {
	return &AuthHandler{
		oidc:                oidc,
		members:             memberGetterAdapter{repo: members},
		audit:               auditWriterAdapter{repo: audit},
		jwtSecret:           jwtSecret,
		jwtExpiry:           jwtExpiry,
		loginRedirect:       loginRedirect,
		bootstrapAdminEmail: bootstrapAdminEmail,
	}
}

// memberGetterAdapter adapts a port.MemberRepository to the handler's
// memberGetter interface (which is keyed on *http.Request).
type memberGetterAdapter struct {
	repo port.MemberRepository
}

func (a memberGetterAdapter) GetByEmail(r *http.Request, email string) (*domain.Member, error) {
	return a.repo.GetByEmail(reqCtx(r), email)
}

func (a memberGetterAdapter) GetByOIDCSubject(r *http.Request, provider, subject string) (*domain.Member, error) {
	return a.repo.GetByOIDCSubject(reqCtx(r), provider, subject)
}

func (a memberGetterAdapter) Create(r *http.Request, m *domain.Member) error {
	return a.repo.Create(reqCtx(r), m)
}

func (a memberGetterAdapter) Update(r *http.Request, m *domain.Member) error {
	return a.repo.Update(reqCtx(r), m)
}

func (a memberGetterAdapter) LinkOIDC(r *http.Request, link *domain.OIDCLink) error {
	return a.repo.LinkOIDC(reqCtx(r), link)
}

// auditWriterAdapter adapts a port.AuditRepository to the handler's auditWriter.
type auditWriterAdapter struct {
	repo port.AuditRepository
}

func (a auditWriterAdapter) WriteAudit(r *http.Request, actorID *uuid.UUID, action, entity, entityID string, before, after interface{}) error {
	entry, err := domain.NewAuditEntry(actorID, action, entity, entityID, before, after)
	if err != nil {
		return err
	}
	return a.repo.Insert(reqCtx(r), entry)
}

func reqCtx(r *http.Request) context.Context {
	if r == nil {
		return context.Background()
	}
	return r.Context()
}
