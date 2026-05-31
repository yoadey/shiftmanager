package handler

import (
	"net/http"

	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// StatsHandler serves the admin statistics endpoint (D-004).
type StatsHandler struct {
	uc *usecase.StatsUsecase
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(uc *usecase.StatsUsecase) *StatsHandler {
	return &StatsHandler{uc: uc}
}

// GetStats returns system-wide statistics.
// GET /api/v1/stats
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.uc.GetStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// PrivacyHandler serves GDPR export/deletion (DS-003/DS-004) and member
// reminder preferences (N-001).
type PrivacyHandler struct {
	uc *usecase.MemberPrivacyUsecase
}

// NewPrivacyHandler creates a new PrivacyHandler.
func NewPrivacyHandler(uc *usecase.MemberPrivacyUsecase) *PrivacyHandler {
	return &PrivacyHandler{uc: uc}
}

// ExportData returns the GDPR data export for a member (DS-003). A member may
// export their own data; Vorstand+ may export any member's data.
// GET /api/v1/members/{id}/export-data
func (h *PrivacyHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid member id")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if actorID != id && !domain.HasRole(role, domain.RoleVorstand) {
		writeError(w, http.StatusForbidden, "insufficient permissions")
		return
	}
	export, err := h.uc.ExportData(r.Context(), actorID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "member not found")
		return
	}
	writeJSON(w, http.StatusOK, export)
}

// GDPRDelete anonymizes a member (DS-004, Vorstand+ enforced by the router).
// POST /api/v1/members/{id}/gdpr-delete
func (h *PrivacyHandler) GDPRDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid member id")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	if err := h.uc.GDPRDelete(r.Context(), actorID, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "anonymized"})
}

// UpdatePreferences updates the authenticated member's own reminder opt-out (N-001).
// PUT /api/v1/members/me/preferences
func (h *PrivacyHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r.Context())
	var body struct {
		ReminderOptOut bool `json:"reminderOptOut"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.uc.SetReminderOptOut(r.Context(), actorID, actorID, body.ReminderOptOut); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"reminderOptOut": body.ReminderOptOut})
}
