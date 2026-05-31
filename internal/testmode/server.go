// Package testmode provides a self-contained HTTP server wired with in-memory
// repositories and an auto-auth token endpoint. It is used by integration tests
// to exercise the full request→usecase→repo round-trip without a real database
// or OIDC provider.
package testmode

import (
	"context"
	"net/http/httptest"

	"github.com/yoadey/shiftmanager/internal/adapter/cache"
	"github.com/yoadey/shiftmanager/internal/adapter/email"
	httpadapter "github.com/yoadey/shiftmanager/internal/adapter/http"
	"github.com/yoadey/shiftmanager/internal/adapter/http/handler"
	"github.com/yoadey/shiftmanager/internal/adapter/memory"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
	"github.com/rs/zerolog"
)

const TestJWTSecret = "integration-test-secret"

// Server wraps an httptest.Server pre-wired for integration tests.
type Server struct {
	*httptest.Server
	Members  *memory.MemberRepo
	Events   *memory.EventRepo
	Shifts   *memory.ShiftRepo
	Regs     *memory.RegistrationRepo
	Hours    *memory.HourRepo
	Settings *memory.SettingsRepo
}

// New starts a test HTTP server with in-memory repos and seeded data.
// Call s.Close() when done.
func New(ctx context.Context) (*Server, error) {
	// --- Repos ---
	members := memory.NewMemberRepo()
	events := memory.NewEventRepo()
	shifts := memory.NewShiftRepo()
	regs := memory.NewRegistrationRepo()
	hours := memory.NewHourRepo()
	audit := memory.NewAuditRepo()
	settings := memory.NewSettingsRepo()
	emailRepo := memory.NewEmailRepo()

	if err := memory.Seed(ctx, members, events, shifts, hours); err != nil {
		return nil, err
	}

	// --- Services ---
	memCache := cache.NewMemoryCache()
	emailSvc, _ := email.NewSMTPService(email.Config{
		Host:    "localhost",
		Port:    25,
		From:    "noreply@test.local",
		BaseURL: "http://localhost",
	}, emailRepo, emailRepo)

	// --- Usecases ---
	memberUC := usecase.NewMemberUsecase(members, audit)
	eventUC := usecase.NewEventUsecase(events, shifts, regs, audit)
	regUC := usecase.NewRegistrationUsecase(regs, shifts, events, members, emailSvc, audit, settings)
	hourUC := usecase.NewHourUsecase(hours, members, shifts, audit)
	billingUC := usecase.NewBillingUsecase(hours, members, settings, audit)
	settingsUC := usecase.NewSettingsUsecase(settings, audit, memCache)
	templateUC := usecase.NewEmailTemplateUsecase(
		port.EmailTemplateRepository(emailRepo),
		port.EmailLogRepository(emailRepo),
		emailSvc, audit,
	)
	statsUC := usecase.NewStatsUsecase(hours, members, shifts, regs)
	privacyUC := usecase.NewMemberPrivacyUsecase(members, regs, hours, audit)
	notifUC := usecase.NewNotificationUsecase(hours, members, shifts, events, regs, settings, emailSvc, audit, "noreply@test.local")
	regUC.SetUnderstaffedNotifier(notifUC)

	// --- Handlers ---
	log := zerolog.Nop()
	handlers := httpadapter.Handlers{
		Auth:     handler.BuildAuthHandler(nil, members, audit, TestJWTSecret, 0, "/auth/callback", ""),
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
		Logger:    log,
		Ready:     func() bool { return true },
		TestMode:  true,
	})

	ts := httptest.NewServer(router)
	return &Server{
		Server:   ts,
		Members:  members,
		Events:   events,
		Shifts:   shifts,
		Regs:     regs,
		Hours:    hours,
		Settings: settings,
	}, nil
}
