package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// KioskHandler handles unauthenticated public kiosk registration flows.
type KioskHandler struct {
	regUC    *usecase.RegistrationUsecase
	eventUC  *usecase.EventUsecase
	memberUC *usecase.MemberUsecase
}

// NewKioskHandler creates a new KioskHandler.
func NewKioskHandler(regUC *usecase.RegistrationUsecase, eventUC *usecase.EventUsecase, memberUC *usecase.MemberUsecase) *KioskHandler {
	return &KioskHandler{regUC: regUC, eventUC: eventUC, memberUC: memberUC}
}

// PublicTimeline returns the timeline of a published event for the kiosk view.
// GET /api/v1/kiosk/events/:id
func (h *KioskHandler) PublicTimeline(w http.ResponseWriter, r *http.Request) {
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
	shiftID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shift id")
		return
	}

	var body struct {
		GuestEmail string `json:"guestEmail"`
		Comment    string `json:"comment"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.ToLower(strings.TrimSpace(body.GuestEmail))
	if email == "" {
		writeError(w, http.StatusBadRequest, "guestEmail is required")
		return
	}

	reg, err := h.regUC.Register(r.Context(), nil, usecase.RegisterInput{
		ShiftID:    shiftID,
		GuestEmail: &email,
		Comment:    body.Comment,
	})
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
