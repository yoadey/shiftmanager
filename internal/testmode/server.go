// Package testmode provides a self-contained HTTP server backed by a SQLite
// in-memory database via GORM. It is used exclusively for integration tests —
// no PostgreSQL or OIDC provider required.
package testmode

import (
	"context"
	"net/http/httptest"

	"github.com/rs/zerolog"
	"github.com/yoadey/shiftmanager/internal/adapter/cache"
	"github.com/yoadey/shiftmanager/internal/adapter/db"
	"github.com/yoadey/shiftmanager/internal/adapter/email"
	httpadapter "github.com/yoadey/shiftmanager/internal/adapter/http"
	"github.com/yoadey/shiftmanager/internal/adapter/http/handler"
	oidcadapter "github.com/yoadey/shiftmanager/internal/adapter/oidc"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

const TestJWTSecret = "integration-test-secret"

// Server wraps an httptest.Server ready for integration tests.
type Server struct {
	*httptest.Server
}

// New starts an in-memory GORM/SQLite-backed HTTP server with seeded fixture data.
// Call s.Close() when done.
func New(_ context.Context) (*Server, error) {
	// SQLite in-memory DB — a fresh schema on every test run.
	gdb, err := db.Open("sqlite::memory:")
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(gdb); err != nil {
		return nil, err
	}
	if err := db.SeedTestData(gdb); err != nil {
		return nil, err
	}
	if err := db.SeedDefaultTemplates(gdb); err != nil {
		return nil, err
	}

	// --- Repositories ---
	memberRepo := db.NewMemberRepo(gdb)
	eventRepo := db.NewEventRepo(gdb)
	shiftRepo := db.NewShiftRepo(gdb)
	regRepo := db.NewRegistrationRepo(gdb)
	hourRepo := db.NewHourRepo(gdb)
	auditRepo := db.NewAuditRepo(gdb)
	settingsRepo := db.NewSettingsRepo(gdb)
	templateRepo := db.NewEmailTemplateRepo(gdb)
	emailLogRepo := db.NewEmailLogRepo(gdb)

	// --- Services ---
	memCache := cache.NewMemoryCache()
	emailSvc, _ := email.NewSMTPService(email.Config{
		Host:    "localhost",
		Port:    25,
		From:    "noreply@test.local",
		BaseURL: "http://localhost",
	}, templateRepo, emailLogRepo)

	// --- Usecases ---
	memberUC := usecase.NewMemberUsecase(memberRepo, auditRepo)
	eventUC := usecase.NewEventUsecase(eventRepo, shiftRepo, regRepo, auditRepo, emailSvc)
	regUC := usecase.NewRegistrationUsecase(regRepo, shiftRepo, eventRepo, memberRepo, emailSvc, auditRepo, settingsRepo)
	hourUC := usecase.NewHourUsecase(hourRepo, memberRepo, shiftRepo, auditRepo, emailSvc, eventRepo)
	billingUC := usecase.NewBillingUsecase(hourRepo, memberRepo, settingsRepo, auditRepo)
	settingsUC := usecase.NewSettingsUsecase(settingsRepo, auditRepo, memCache)
	templateUC := usecase.NewEmailTemplateUsecase(templateRepo, emailLogRepo, emailSvc, auditRepo)
	statsUC := usecase.NewStatsUsecase(hourRepo, memberRepo, shiftRepo, regRepo)
	privacyUC := usecase.NewMemberPrivacyUsecase(memberRepo, regRepo, hourRepo, auditRepo)
	notifUC := usecase.NewNotificationUsecase(hourRepo, memberRepo, shiftRepo, eventRepo, regRepo, settingsRepo, emailSvc, auditRepo, "noreply@test.local")
	reminderUC := usecase.NewReminderUsecase(shiftRepo, regRepo, eventRepo, memberRepo, emailSvc, auditRepo, settingsRepo)
	_ = reminderUC
	regUC.SetUnderstaffedNotifier(notifUC)

	// --- Handlers ---
	var oidcSvc port.OIDCService
	_ = oidcadapter.Config{} // ensure import used

	handlers := httpadapter.Handlers{
		Auth:     handler.BuildAuthHandler(oidcSvc, memberRepo, auditRepo, TestJWTSecret, 0, "/auth/callback", ""),
		Member:   handler.NewMemberHandler(memberUC),
		Event:    handler.NewEventHandler(eventUC),
		Shift:    handler.NewShiftHandler(eventUC, regUC),
		Hour:     handler.NewHourHandler(hourUC),
		Kiosk:    handler.NewKioskHandler(regUC, eventUC, memberUC, settingsUC),
		Settings: handler.NewSettingsHandler(settingsUC, templateUC, "", "http://localhost"),
		Billing:  handler.NewBillingHandler(billingUC),
		Stats:    handler.NewStatsHandler(statsUC),
		Privacy:  handler.NewPrivacyHandler(privacyUC),
		OpenAPI:  handler.NewOpenAPIHandler(),
		DevAuth:  handler.NewDevAuthHandler(TestJWTSecret),
	}

	router := httpadapter.NewRouter(handlers, httpadapter.RouterConfig{
		JWTSecret: TestJWTSecret,
		Logger:    zerolog.Nop(),
		Ready:     func() bool { return true },
		TestMode:  true,
	})

	ts := httptest.NewServer(router)
	return &Server{Server: ts}, nil
}
