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
	localstorage "github.com/yoadey/shiftmanager/internal/adapter/storage/local"
	s3storage "github.com/yoadey/shiftmanager/internal/adapter/storage/s3"
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

	// One OIDCService per configured provider (A-005). The common
	// single-provider case still yields exactly one entry, named "default".
	oidcSvcs := map[string]port.OIDCService{}
	var oidcProviders []handler.OIDCProviderInfo
	for _, p := range cfg.Providers() {
		if p.Issuer == "" {
			continue
		}
		svc, err := oidcadapter.New(rootCtx, oidcadapter.Config{
			Issuer:       p.Issuer,
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			RedirectURL:  p.RedirectURL,
		})
		if err != nil {
			log.Warn().Err(err).Str("provider", p.Name).Msg("OIDC provider init failed; login via this provider will be unavailable")
			continue
		}
		oidcSvcs[p.Name] = svc
		oidcProviders = append(oidcProviders, handler.OIDCProviderInfo{Name: p.Name, Label: p.Label})
	}

	var mediaStorage port.MediaStorage
	switch cfg.MediaStorage {
	case "s3":
		s3Storage, err := s3storage.New(rootCtx, s3storage.Config{
			Endpoint:        cfg.S3Endpoint,
			Region:          cfg.S3Region,
			Bucket:          cfg.S3Bucket,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			ForcePathStyle:  cfg.S3ForcePathStyle,
			PublicBaseURL:   cfg.S3PublicBaseURL,
		})
		if err != nil {
			return fmt.Errorf("init s3 media storage: %w", err)
		}
		mediaStorage = s3Storage
	default:
		mediaStorage = localstorage.New(cfg.UploadDir, cfg.BaseURL)
	}

	// --- Usecases ---
	memberUC := usecase.NewMemberUsecase(memberRepo, auditRepo)
	eventUC := usecase.NewEventUsecase(eventRepo, shiftRepo, regRepo, auditRepo, emailSvc, memberRepo, eventAttachmentRepo)
	regUC := usecase.NewRegistrationUsecase(regRepo, shiftRepo, eventRepo, memberRepo, emailSvc, auditRepo, settingsRepo)
	hourUC := usecase.NewHourUsecase(hourRepo, memberRepo, shiftRepo, regRepo, auditRepo, emailSvc, eventRepo)
	billingUC := usecase.NewBillingUsecase(hourRepo, memberRepo, settingsRepo, auditRepo, shiftRepo, regRepo)
	settingsUC := usecase.NewSettingsUsecase(settingsRepo, auditRepo, memCache)
	reminderUC := usecase.NewReminderUsecase(shiftRepo, regRepo, eventRepo, memberRepo, emailSvc, auditRepo, settingsRepo)
	templateUC := usecase.NewEmailTemplateUsecase(templateRepo, emailLogRepo, emailSvc, auditRepo)
	statsUC := usecase.NewStatsUsecase(hourRepo, memberRepo, shiftRepo, regRepo)
	privacyUC := usecase.NewMemberPrivacyUsecase(memberRepo, regRepo, hourRepo, auditRepo)
	notificationUC := usecase.NewNotificationUsecase(hourRepo, memberRepo, shiftRepo, eventRepo, regRepo, settingsRepo, emailSvc, auditRepo, cfg.SMTPFrom)
	regUC.SetUnderstaffedNotifier(notificationUC)

	// --- Handlers ---
	handlers := httpadapter.Handlers{
		Auth:     handler.BuildAuthHandler(oidcSvcs, oidcProviders, memberRepo, auditRepo, cfg.JWTSecret, cfg.JWTExpiration, cfg.LoginRedirectURL, cfg.BootstrapAdminEmail),
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
	}
	if cfg.TestMode {
		handlers.DevAuth = handler.NewDevAuthHandler(cfg.JWTSecret)
	}

	// The /uploads/* static route only serves local media storage; with
	// MEDIA_STORAGE=s3 nothing is ever written under UploadDir, so the route
	// is left unmounted (RouterConfig.UploadDir == "" disables it).
	routerUploadDir := cfg.UploadDir
	if cfg.MediaStorage == "s3" {
		routerUploadDir = ""
	}
	router := httpadapter.NewRouter(handlers, httpadapter.RouterConfig{
		JWTSecret:    cfg.JWTSecret,
		RateLimitRPM: 120,
		Logger:       log,
		Ready:        dbReadyFunc(gdb),
		Static:       static.Handler(),
		UploadDir:    routerUploadDir,
		TestMode:     cfg.TestMode,
	})

	// Scheduler uses a raw *sql.DB for advisory locks; skip in SQLite test mode.
	if !cfg.TestMode {
		sqlDB, err := gdb.DB()
		if err != nil {
			return fmt.Errorf("get underlying sql.DB: %w", err)
		}
		sched := scheduler.New(sqlDB, regUC, reminderUC, notificationUC, auditRepo, log)
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
