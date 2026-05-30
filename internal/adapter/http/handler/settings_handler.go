package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// SettingsHandler handles application settings, branding, fee tiers and audit log.
type SettingsHandler struct {
	uc *usecase.SettingsUsecase
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(uc *usecase.SettingsUsecase) *SettingsHandler {
	return &SettingsHandler{uc: uc}
}

// GetSettings returns the current application settings.
// GET /api/v1/settings
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// UpdateSettings persists changed application settings.
// PUT /api/v1/settings
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body domain.AppSettings
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	s, err := h.uc.UpdateSettings(r.Context(), actorID, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// GetBranding returns the current branding configuration.
// GET /api/v1/settings/branding
func (h *SettingsHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	b, err := h.uc.GetBranding(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// UpdateBranding persists a new branding configuration.
// PUT /api/v1/settings/branding
func (h *SettingsHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	var body domain.BrandingConfig
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	b, err := h.uc.UpdateBranding(r.Context(), actorID, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// GetFeeTiers returns the fee tiers for a club year.
// GET /api/v1/settings/fee-tiers?clubYearId=...
func (h *SettingsHandler) GetFeeTiers(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}

	tiers, err := h.uc.GetFeeTiers(r.Context(), clubYearID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

// UpdateFeeTiers replaces all fee tiers for a club year.
// PUT /api/v1/settings/fee-tiers?clubYearId=...
func (h *SettingsHandler) UpdateFeeTiers(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}

	var body struct {
		Tiers []*domain.FeeTier `json:"tiers"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	tiers, err := h.uc.UpdateFeeTiers(r.Context(), actorID, clubYearID, body.Tiers)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

// GetAuditLog returns filtered audit log entries.
// GET /api/v1/settings/audit
func (h *SettingsHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	filter := port.AuditFilter{
		Entity:   r.URL.Query().Get("entity"),
		EntityID: r.URL.Query().Get("entityId"),
		Limit:    parseIntQuery(r, "limit", 100),
		Offset:   parseIntQuery(r, "offset", 0),
	}

	if a := r.URL.Query().Get("actorId"); a != "" {
		if id, err := uuid.Parse(a); err == nil {
			filter.ActorID = &id
		}
	}
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &t
		}
	}

	entries, err := h.uc.GetAuditLog(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
