// Package testmode provides a self-contained HTTP server backed by a SQLite
// in-memory database via GORM. It is used exclusively for integration tests —
// no PostgreSQL or OIDC provider required.
package testmode

import (
	"context"
	"net/http/httptest"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/yoadey/shiftmanager/internal/adapter/cache"
	"github.com/yoadey/shiftmanager/internal/adapter/db"
	"github.com/yoadey/shiftmanager/internal/adapter/email"
	httpadapter "github.com/yoadey/shiftmanager/internal/adapter/http"
	"github.com/yoadey/shiftmanager/internal/adapter/http/handler"
	oidcadapter "github.com/yoadey/shiftmanager/internal/adapter/oidc"
	localstorage "github.com/yoadey/shiftmanager/internal/adapter/storage/local"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

const TestJWTSecret = "integration-test-secret"

// Server wraps an httptest.Server ready for integration tests.
type Server struct {
	*httptest.Server
	uploadDir string
}

// Close shuts down the underlying HTTP server and removes the temp upload dir.
func (s *Server) Close() {
	s.Server.Close()
	if s.uploadDir != "" {
		_ = os.RemoveAll(s.uploadDir)
	}
}

// New starts an in-memory GORM/SQLite-backed HTTP server with seeded fixture data.
// Call s.Close() when done.
func New(_ context.Context) (*Server, error) {
	// Uploads (logos, event attachments) go to a throwaway temp dir rather
	// than the repo's ./uploads, so tests that exercise them don't leave
	// files behind or clash when run in parallel.
	uploadDir, err := os.MkdirTemp("", "shiftmanager-uploads-*")
	if err != nil {
		return nil, err
	}
	ready := false
	defer func() {
		// Only reached if New() returns before the server is fully up;
		// once returned successfully, Server.Close() owns removing uploadDir.
		if !ready {
			_ = os.RemoveAll(uploadDir)
		}
	}()

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
	eventAttachmentRepo := db.NewEventAttachmentRepo(gdb)
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
	eventUC := usecase.NewEventUsecase(eventRepo, shiftRepo, regRepo, auditRepo, emailSvc, memberRepo, eventAttachmentRepo)
	regUC := usecase.NewRegistrationUsecase(regRepo, shiftRepo, eventRepo, memberRepo, emailSvc, auditRepo, settingsRepo)
	hourUC := usecase.NewHourUsecase(hourRepo, memberRepo, shiftRepo, regRepo, auditRepo, emailSvc, eventRepo)
	billingUC := usecase.NewBillingUsecase(hourRepo, memberRepo, settingsRepo, auditRepo, shiftRepo, regRepo)
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
	mediaStorage := localstorage.New(uploadDir, "http://localhost")

	handlers := httpadapter.Handlers{
		Auth:     handler.BuildAuthHandler(oidcSvc, memberRepo, auditRepo, TestJWTSecret, 24*time.Hour, "/auth/callback", ""),
		Member:   handler.NewMemberHandler(memberUC),
		Event:    handler.NewEventHandler(eventUC, mediaStorage),
		Shift:    handler.NewShiftHandler(eventUC, regUC),
		Hour:     handler.NewHourHandler(hourUC),
		Kiosk:    handler.NewKioskHandler(regUC, eventUC, memberUC, settingsUC),
		Settings: handler.NewSettingsHandler(settingsUC, templateUC, mediaStorage),
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
		UploadDir: uploadDir,
		TestMode:  true,
	})

	ts := httptest.NewServer(router)
	ready = true
	return &Server{Server: ts, uploadDir: uploadDir}, nil
}
