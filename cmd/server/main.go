package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	// automaxprocs sets GOMAXPROCS to match the container CPU quota.
	_ "go.uber.org/automaxprocs"

	"github.com/yoadey/shiftmanager/internal/adapter/cache"
	"github.com/yoadey/shiftmanager/internal/adapter/db"
	"github.com/yoadey/shiftmanager/internal/adapter/email"
	httpadapter "github.com/yoadey/shiftmanager/internal/adapter/http"
	"github.com/yoadey/shiftmanager/internal/adapter/http/handler"
	oidcadapter "github.com/yoadey/shiftmanager/internal/adapter/oidc"
	"github.com/yoadey/shiftmanager/internal/config"
	"github.com/yoadey/shiftmanager/internal/infrastructure/logger"
	"github.com/yoadey/shiftmanager/internal/infrastructure/scheduler"
	"github.com/yoadey/shiftmanager/internal/infrastructure/static"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

func main() {
	if v := os.Getenv("GOMEMLIMIT"); v == "" {
		debug.SetMemoryLimit(-1)
	}
	_ = godotenv.Load()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrateCommand()
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.LogLevel)
	if cfg.TestMode {
		log.Warn().Msg("⚠️  TEST MODE — SQLite in-memory, /dev/token active")
	}
	log.Info().Str("port", cfg.Port).Msg("starting shiftmanager")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Database via GORM (SQLite for test mode, PostgreSQL otherwise) ---
	dsn := cfg.DatabaseURL
	if cfg.TestMode {
		dsn = "sqlite::memory:"
	}
	gdb, err := db.Open(dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	if err := db.Migrate(gdb); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if cfg.TestMode {
		if err := db.SeedTestData(gdb); err != nil {
			return fmt.Errorf("seed test data: %w", err)
		}
	}

	// --- Optional Redis ---
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return fmt.Errorf("parse redis url: %w", err)
		}
		rc := redis.NewClient(opt)
		if err := rc.Ping(rootCtx).Err(); err != nil {
			log.Warn().Err(err).Msg("redis ping failed; continuing without redis")
			_ = rc.Close()
		} else {
			defer func() { _ = rc.Close() }()
			log.Info().Msg("connected to redis")
		}
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
	defer memCache.Close()

	emailSvc, err := email.NewSMTPService(email.Config{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPass,
		From:     cfg.SMTPFrom,
		BaseURL:  cfg.BaseURL,
	}, templateRepo, emailLogRepo)
	if err != nil {
		return fmt.Errorf("init email service: %w", err)
	}

	var oidcSvc port.OIDCService
	if cfg.OIDCIssuer != "" {
		svc, err := oidcadapter.New(rootCtx, oidcadapter.Config{
			Issuer:       cfg.OIDCIssuer,
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			RedirectURL:  cfg.OIDCRedirectURL,
		})
		if err != nil {
			log.Warn().Err(err).Msg("OIDC provider init failed; auth login will be unavailable")
		} else {
			oidcSvc = svc
		}
	}

	// --- Usecases ---
	memberUC := usecase.NewMemberUsecase(memberRepo, auditRepo)
	eventUC := usecase.NewEventUsecase(eventRepo, shiftRepo, regRepo, auditRepo)
	regUC := usecase.NewRegistrationUsecase(regRepo, shiftRepo, eventRepo, memberRepo, emailSvc, auditRepo, settingsRepo)
	hourUC := usecase.NewHourUsecase(hourRepo, memberRepo, shiftRepo, auditRepo)
	billingUC := usecase.NewBillingUsecase(hourRepo, memberRepo, settingsRepo, auditRepo)
	settingsUC := usecase.NewSettingsUsecase(settingsRepo, auditRepo, memCache)
	reminderUC := usecase.NewReminderUsecase(shiftRepo, regRepo, eventRepo, memberRepo, emailSvc, auditRepo, settingsRepo)
	templateUC := usecase.NewEmailTemplateUsecase(templateRepo, emailLogRepo, emailSvc, auditRepo)
	statsUC := usecase.NewStatsUsecase(hourRepo, memberRepo, shiftRepo, regRepo)
	privacyUC := usecase.NewMemberPrivacyUsecase(memberRepo, regRepo, hourRepo, auditRepo)
	notificationUC := usecase.NewNotificationUsecase(hourRepo, memberRepo, shiftRepo, eventRepo, regRepo, settingsRepo, emailSvc, auditRepo, cfg.SMTPFrom)
	regUC.SetUnderstaffedNotifier(notificationUC)

	// --- Handlers ---
	handlers := httpadapter.Handlers{
		Auth:     handler.BuildAuthHandler(oidcSvc, memberRepo, auditRepo, cfg.JWTSecret, cfg.JWTExpiration, cfg.LoginRedirectURL, cfg.BootstrapAdminEmail),
		Member:   handler.NewMemberHandler(memberUC),
		Event:    handler.NewEventHandler(eventUC),
		Shift:    handler.NewShiftHandler(eventUC, regUC),
		Hour:     handler.NewHourHandler(hourUC),
		Kiosk:    handler.NewKioskHandler(regUC, eventUC, memberUC, settingsUC),
		Settings: handler.NewSettingsHandler(settingsUC, templateUC, cfg.UploadDir, cfg.BaseURL),
		Billing:  handler.NewBillingHandler(billingUC),
		Stats:    handler.NewStatsHandler(statsUC),
		Privacy:  handler.NewPrivacyHandler(privacyUC),
		OpenAPI:  handler.NewOpenAPIHandler(),
	}
	if cfg.TestMode {
		handlers.DevAuth = handler.NewDevAuthHandler(cfg.JWTSecret)
	}

	router := httpadapter.NewRouter(handlers, httpadapter.RouterConfig{
		JWTSecret:    cfg.JWTSecret,
		RateLimitRPM: 120,
		Logger:       log,
		Ready:        dbReadyFunc(gdb),
		Static:       static.Handler(),
		UploadDir:    cfg.UploadDir,
		TestMode:     cfg.TestMode,
	})

	// Scheduler uses a raw *sql.DB for advisory locks; skip in SQLite test mode.
	if !cfg.TestMode {
		sqlDB, err := gdb.DB()
		if err != nil {
			return fmt.Errorf("get underlying sql.DB: %w", err)
		}
		sched := scheduler.New(sqlDB, regUC, reminderUC, notificationUC, log)
		sched.Start(rootCtx)
		defer sched.Stop()
	}

	// --- HTTP server ---
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("http server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-rootCtx.Done():
		log.Info().Msg("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info().Msg("server stopped cleanly")
	return nil
}

func dbReadyFunc(gdb *gorm.DB) func() bool {
	return func() bool {
		sqlDB, err := gdb.DB()
		if err != nil {
			return false
		}
		return sqlDB.Ping() == nil
	}
}

// runMigrateCommand runs GORM AutoMigrate for the `migrate` subcommand.
// For production PostgreSQL this replaces the old golang-migrate approach.
func runMigrateCommand() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required for the migrate command")
		os.Exit(1)
	}
	gdb, err := db.Open(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	if err := db.Migrate(gdb); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("migrations applied")
}
