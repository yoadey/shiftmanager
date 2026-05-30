package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/yoadey/shiftmanager/internal/adapter/http/handler"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/infrastructure/logger"
)

// Handlers groups every HTTP handler the router needs to mount.
type Handlers struct {
	Auth     *handler.AuthHandler
	Member   *handler.MemberHandler
	Event    *handler.EventHandler
	Shift    *handler.ShiftHandler
	Hour     *handler.HourHandler
	Kiosk    *handler.KioskHandler
	Settings *handler.SettingsHandler
	Billing  *handler.BillingHandler
}

// RouterConfig holds the cross-cutting configuration for the router.
type RouterConfig struct {
	JWTSecret      string
	AllowedOrigins []string
	RateLimitRPM   int
	Logger         zerolog.Logger
	// Ready reports whether the application's dependencies (e.g. the database)
	// are healthy. May be nil, in which case /readyz always returns 200.
	Ready func() bool
	// Static serves the embedded SPA. May be nil to disable static serving.
	Static http.Handler
}

// NewRouter builds the chi router with the full middleware stack and all routes.
func NewRouter(h Handlers, cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// --- Global middleware ---
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(logger.RequestLogger(cfg.Logger))

	allowedOrigins := cfg.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- Health probes (no auth, no rate limit) ---
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if cfg.Ready != nil && !cfg.Ready() {
			http.Error(w, `{"status":"unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	// --- API ---
	r.Route("/api/v1", func(api chi.Router) {
		// Public auth routes (rate-limited, no JWT).
		api.Group(func(pub chi.Router) {
			pub.Use(middleware.RateLimit(cfg.RateLimitRPM))
			pub.Get("/auth/login", h.Auth.Login)
			pub.Get("/auth/callback", h.Auth.Callback)
		})

		// Public kiosk routes (rate-limited, no JWT).
		api.Group(func(kiosk chi.Router) {
			kiosk.Use(middleware.RateLimit(cfg.RateLimitRPM))
			kiosk.Get("/kiosk/events/{id}", h.Kiosk.PublicTimeline)
			kiosk.Post("/kiosk/shifts/{id}/register", h.Kiosk.Register)
			kiosk.Get("/kiosk/confirm/{token}", h.Kiosk.Confirm)
		})

		// Authenticated routes.
		api.Group(func(auth chi.Router) {
			auth.Use(middleware.JWTAuth(cfg.JWTSecret))

			// Auth/session.
			auth.Get("/auth/me", h.Auth.Me)
			auth.Post("/auth/logout", h.Auth.Logout)

			// Members (board+ for write operations).
			auth.Route("/members", func(m chi.Router) {
				m.Get("/", h.Member.List)
				m.Get("/export", h.Member.Export)
				m.Get("/{id}", h.Member.Get)
				m.Group(func(w chi.Router) {
					w.Use(middleware.RequireRole(domain.RoleVorstand))
					w.Post("/", h.Member.Create)
					w.Post("/import", h.Member.Import)
					w.Put("/{id}", h.Member.Update)
					w.Delete("/{id}", h.Member.Deactivate)
				})
			})

			// Events and shifts.
			auth.Route("/events", func(e chi.Router) {
				e.Get("/", h.Event.List)
				e.Get("/{id}", h.Event.GetWithTimeline)
				e.Get("/{id}/timeline", h.Event.GetTimeline)
				e.Group(func(w chi.Router) {
					w.Use(middleware.RequireRole(domain.RoleVeranstaltungsleiter))
					w.Post("/", h.Event.Create)
					w.Put("/{id}", h.Event.Update)
					w.Delete("/{id}", h.Event.Delete)
					w.Post("/{id}/publish", h.Event.Publish)
					w.Post("/{eventId}/shifts", h.Shift.CreateShift)
				})
			})

			auth.Route("/shifts", func(sh chi.Router) {
				sh.Post("/{id}/register", h.Shift.Register)
				sh.Delete("/{id}/register", h.Shift.Deregister)
				sh.Group(func(w chi.Router) {
					w.Use(middleware.RequireRole(domain.RoleVeranstaltungsleiter))
					w.Put("/{id}", h.Shift.UpdateShift)
					w.Delete("/{id}", h.Shift.DeleteShift)
					w.Post("/{id}/confirm", h.Shift.ConfirmRegistration)
				})
			})

			// Hours.
			auth.Route("/hours", func(hr chi.Router) {
				hr.Get("/account", h.Hour.GetMemberAccount)
				hr.Group(func(w chi.Router) {
					w.Use(middleware.RequireRole(domain.RoleVeranstaltungsleiter))
					w.Post("/confirm", h.Hour.ConfirmShiftHours)
					w.Get("/summary", h.Hour.GetYearSummary)
				})
				hr.Group(func(w chi.Router) {
					w.Use(middleware.RequireRole(domain.RoleVorstand))
					w.Post("/manual", h.Hour.ManualBooking)
					w.Put("/{id}", h.Hour.CorrectEntry)
					w.Delete("/{id}", h.Hour.DeleteEntry)
				})
			})

			// Settings, branding, fee tiers, audit (board/admin).
			auth.Route("/settings", func(st chi.Router) {
				st.Get("/", h.Settings.GetSettings)
				st.Get("/branding", h.Settings.GetBranding)
				st.Get("/fee-tiers", h.Settings.GetFeeTiers)
				st.Group(func(w chi.Router) {
					w.Use(middleware.RequireRole(domain.RoleVorstand))
					w.Put("/", h.Settings.UpdateSettings)
					w.Put("/branding", h.Settings.UpdateBranding)
					w.Put("/fee-tiers", h.Settings.UpdateFeeTiers)
					w.Get("/audit", h.Settings.GetAuditLog)
				})
			})

			// Billing (board/admin).
			auth.Route("/billing", func(b chi.Router) {
				b.Use(middleware.RequireRole(domain.RoleVorstand))
				b.Get("/{clubYearId}", h.Billing.Compute)
				b.Get("/{clubYearId}/export.csv", h.Billing.ExportCSV)
				b.Get("/{clubYearId}/export.pdf", h.Billing.ExportPDF)
			})
		})
	})

	// --- SPA fallback for everything that is not an API route ---
	if cfg.Static != nil {
		r.NotFound(func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/api/") {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			cfg.Static.ServeHTTP(w, req)
		})
	}

	return r
}
