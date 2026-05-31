package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	// automaxprocs sets GOMAXPROCS to match the container CPU quota.
	_ "go.uber.org/automaxprocs"

	"github.com/yoadey/shiftmanager/internal/adapter/cache"
	"github.com/yoadey/shiftmanager/internal/adapter/email"
	httpadapter "github.com/yoadey/shiftmanager/internal/adapter/http"
	"github.com/yoadey/shiftmanager/internal/adapter/http/handler"
	oidcadapter "github.com/yoadey/shiftmanager/internal/adapter/oidc"
	"github.com/yoadey/shiftmanager/internal/adapter/postgres"
	"github.com/yoadey/shiftmanager/internal/config"
	"github.com/yoadey/shiftmanager/internal/infrastructure/logger"
	"github.com/yoadey/shiftmanager/internal/infrastructure/scheduler"
	"github.com/yoadey/shiftmanager/internal/infrastructure/static"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

const migrationsPath = "file://migrations"

func main() {
	// The Go runtime reads GOMEMLIMIT from the environment automatically. When it
	// is unset we keep the runtime default (no hard limit). automaxprocs (blank
	// imported above) sets GOMAXPROCS to match the container CPU quota.
	if v := os.Getenv("GOMEMLIMIT"); v == "" {
		// Ensure the soft memory limit stays at its default of math.MaxInt64.
		debug.SetMemoryLimit(-1)
	}

	// Load .env when present; ignore the error in production where env vars are set.
	_ = godotenv.Load()

	// Subcommand: `migrate` runs migrations and exits (used by an init container).
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
	log.Info().Str("port", cfg.Port).Msg("starting shiftmanager")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Database ---
	pool, err := newPool(rootCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// --- Optional Redis ---
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return fmt.Errorf("parse redis url: %w", err)
		}
		redisClient = redis.NewClient(opt)
		if err := redisClient.Ping(rootCtx).Err(); err != nil {
			log.Warn().Err(err).Msg("redis ping failed; continuing without redis")
			_ = redisClient.Close()
			redisClient = nil
		} else {
			defer func() { _ = redisClient.Close() }()
			log.Info().Msg("connected to redis")
		}
	}

	// --- Repositories ---
	memberRepo := postgres.NewMemberRepo(pool)
	eventRepo := postgres.NewEventRepo(pool)
	shiftRepo := postgres.NewShiftRepo(pool)
	regRepo := postgres.NewRegistrationRepo(pool)
	hourRepo := postgres.NewHourRepo(pool)
	auditRepo := postgres.NewAuditRepo(pool)
	settingsRepo := postgres.NewSettingsRepo(pool)
	templateRepo := postgres.NewEmailTemplateRepo(pool)
	emailLogRepo := postgres.NewEmailLogRepo(pool)

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

	router := httpadapter.NewRouter(handlers, httpadapter.RouterConfig{
		JWTSecret:    cfg.JWTSecret,
		RateLimitRPM: 120,
		Logger:       log,
		Ready: func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return pool.Ping(ctx) == nil
		},
		Static:    static.Handler(),
		UploadDir: cfg.UploadDir,
	})

	// --- Scheduler ---
	sched := scheduler.New(pool, regUC, reminderUC, notificationUC, log)
	sched.Start(rootCtx)
	defer sched.Stop()

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

	// Graceful shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info().Msg("server stopped cleanly")
	return nil
}

// newPool creates a pgx connection pool with at most 20 connections.
func newPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	poolCfg.MaxConns = 20
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = time.Minute

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// runMigrations applies all pending up migrations from the migrations directory.
// It opens a dedicated database/sql connection via the pgx stdlib driver, which
// golang-migrate's postgres driver requires.
func runMigrations(dsn string) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open sql db for migrations: %w", err)
	}
	defer sqlDB.Close()

	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// runMigrateCommand runs migrations for the `migrate` subcommand and exits.
func runMigrateCommand() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required for the migrate command")
		os.Exit(1)
	}
	if err := runMigrations(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("migrations applied")
}
