package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// KioskHandler handles unauthenticated public kiosk registration flows.
type KioskHandler struct {
	regUC      *usecase.RegistrationUsecase
	eventUC    *usecase.EventUsecase
	memberUC   *usecase.MemberUsecase
	settingsUC *usecase.SettingsUsecase
}

// NewKioskHandler creates a new KioskHandler. settingsUC may be nil to disable
// the kiosk-lock check.
func NewKioskHandler(regUC *usecase.RegistrationUsecase, eventUC *usecase.EventUsecase, memberUC *usecase.MemberUsecase, settingsUC *usecase.SettingsUsecase) *KioskHandler {
	return &KioskHandler{regUC: regUC, eventUC: eventUC, memberUC: memberUC, settingsUC: settingsUC}
}

// locked reports whether the public kiosk is currently locked (K-012). When
// locked it writes a 403 response and returns true.
func (h *KioskHandler) locked(w http.ResponseWriter, r *http.Request) bool {
	if h.settingsUC != nil && h.settingsUC.IsKioskLocked(r.Context()) {
		writeError(w, http.StatusForbidden, "kiosk is locked")
		return true
	}
	return false
}

// PublicTimeline returns the timeline of a published event for the kiosk view.
// GET /api/v1/kiosk/events/:id
func (h *KioskHandler) PublicTimeline(w http.ResponseWriter, r *http.Request) {
	if h.locked(w, r) {
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	timeline, err := h.eventUC.GetEventTimeline(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}

// Register creates a guest registration from the kiosk. A confirmation email
// with a double-opt-in link is sent to the supplied address.
// POST /api/v1/kiosk/shifts/:id/register
func (h *KioskHandler) Register(w http.ResponseWriter, r *http.Request) {
	if h.locked(w, r) {
		return
	}
	shiftID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shift id")
		return
	}

	var body struct {
		GuestEmail string     `json:"guestEmail"`
		Email      string     `json:"email"`       // frontend alias for guestEmail
		MemberID   *uuid.UUID `json:"memberId"`
		Comment    string     `json:"comment"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Accept "email" as alias for "guestEmail".
	if body.GuestEmail == "" && body.Email != "" {
		body.GuestEmail = body.Email
	}

	var regInput usecase.RegisterInput
	regInput.ShiftID = shiftID
	regInput.Comment = body.Comment
	if body.MemberID != nil {
		regInput.MemberID = body.MemberID
	} else {
		email := strings.ToLower(strings.TrimSpace(body.GuestEmail))
		if email == "" {
			writeError(w, http.StatusBadRequest, "email or memberId is required")
			return
		}
		regInput.GuestEmail = &email
	}

	reg, err := h.regUC.Register(r.Context(), nil, regInput)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "already registered for this shift":
			status = http.StatusConflict
		case "shift is fully booked":
			status = http.StatusUnprocessableEntity
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, reg)
}

// ListEvents returns all published events for the kiosk selection screen.
// GET /api/v1/kiosk/events
func (h *KioskHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	if h.locked(w, r) {
		return
	}
	published := "published"
	events, err := h.eventUC.ListEvents(r.Context(), port.EventFilter{Status: (*domain.EventStatus)(&published)})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// SearchMembers returns active members matching a search query for kiosk lookup.
// GET /api/v1/kiosk/members?q=...
func (h *KioskHandler) SearchMembers(w http.ResponseWriter, r *http.Request) {
	if h.locked(w, r) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		writeError(w, http.StatusBadRequest, "query must be at least 2 characters")
		return
	}
	active := true
	members, err := h.memberUC.ListMembers(r.Context(), port.MemberFilter{Search: q, IsActive: &active})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// Confirm validates a one-time confirmation token from a kiosk email.
// GET /api/v1/kiosk/confirm/:token
func (h *KioskHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	token, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid token")
		return
	}

	reg, err := h.regUC.ConfirmByLink(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, reg)
}
