package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// ShiftHandler handles HTTP requests for shift operations.
type ShiftHandler struct {
	eventUC *usecase.EventUsecase
	regUC   *usecase.RegistrationUsecase
}

// NewShiftHandler creates a new ShiftHandler.
func NewShiftHandler(eventUC *usecase.EventUsecase, regUC *usecase.RegistrationUsecase) *ShiftHandler {
	return &ShiftHandler{eventUC: eventUC, regUC: regUC}
}

// CreateShift adds a new shift to an event.
// POST /api/v1/events/:eventId/shifts
func (h *ShiftHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseUUIDParam(r, "eventId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var body struct {
		Name                  string    `json:"name"`
		StartAt               time.Time `json:"startAt"`
		EndAt                 time.Time `json:"endAt"`
		MinHelpers            int       `json:"minHelpers"`
		MaxHelpers            int       `json:"maxHelpers"`
		RequiredQualification string    `json:"requiredQualification"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	s, err := h.eventUC.CreateShift(r.Context(), actorID, usecase.CreateShiftInput{
		EventID:               eventID,
		Name:                  body.Name,
		StartAt:               body.StartAt,
		EndAt:                 body.EndAt,
		MinHelpers:            body.MinHelpers,
		MaxHelpers:            body.MaxHelpers,
		RequiredQualification: body.RequiredQualification,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

// UpdateShift applies changes to a shift.
// PUT /api/v1/shifts/:id
func (h *ShiftHandler) UpdateShift(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shift id")
		return
	}

	var body struct {
		Name                  string     `json:"name"`
		StartAt               *time.Time `json:"startAt"`
		EndAt                 *time.Time `json:"endAt"`
		MinHelpers            *int       `json:"minHelpers"`
		MaxHelpers            *int       `json:"maxHelpers"`
		RequiredQualification string     `json:"requiredQualification"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	s, err := h.eventUC.UpdateShift(r.Context(), actorID, id, usecase.UpdateShiftInput{
		Name:                  body.Name,
		StartAt:               body.StartAt,
		EndAt:                 body.EndAt,
		MinHelpers:            body.MinHelpers,
		MaxHelpers:            body.MaxHelpers,
		RequiredQualification: body.RequiredQualification,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// DeleteShift removes a shift.
// DELETE /api/v1/shifts/:id
func (h *ShiftHandler) DeleteShift(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shift id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	if err := h.eventUC.DeleteShift(r.Context(), actorID, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Register registers the authenticated member for a shift.
// POST /api/v1/shifts/:id/register
func (h *ShiftHandler) Register(w http.ResponseWriter, r *http.Request) {
	shiftID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shift id")
		return
	}

	var body struct {
		Comment string `json:"comment"`
	}
	_ = decodeJSON(r, &body)

	memberID := middleware.GetUserID(r.Context())
	if memberID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	reg, err := h.regUC.Register(r.Context(), &memberID, usecase.RegisterInput{
		ShiftID:  shiftID,
		MemberID: &memberID,
		Comment:  body.Comment,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case nil:
		default:
			if err.Error() == "already registered for this shift" {
				status = http.StatusConflict
			} else if err.Error() == "shift is fully booked" {
				status = http.StatusUnprocessableEntity
			}
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, reg)
}

// Deregister removes the authenticated member's registration for a shift.
// DELETE /api/v1/shifts/:id/register
func (h *ShiftHandler) Deregister(w http.ResponseWriter, r *http.Request) {
	shiftID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shift id")
		return
	}

	// The registration ID is expected as a query parameter or derived from shiftID + member.
	regIDStr := r.URL.Query().Get("registrationId")
	regID, err := uuid.Parse(regIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "registrationId query param is required")
		return
	}

	memberID := middleware.GetUserID(r.Context())
	force := r.URL.Query().Get("force") == "true"

	_ = shiftID // used for context; the regID already identifies the registration

	if err := h.regUC.Deregister(r.Context(), &memberID, regID, force); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ConfirmRegistration transitions a registration to confirmed (veranstaltungsleiter+).
// POST /api/v1/shifts/:id/confirm
func (h *ShiftHandler) ConfirmRegistration(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RegistrationID uuid.UUID `json:"registrationId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "registrationId is required")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	reg, err := h.regUC.ConfirmShiftRegistration(r.Context(), actorID, body.RegistrationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, reg)
}

// ConfirmByToken handles the one-time confirmation link from kiosk emails.
// GET /api/v1/shifts/confirm/:token
func (h *ShiftHandler) ConfirmByToken(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		// Try URL param.
		tokenStr = parseStringParam(r, "token")
	}

	token, err := uuid.Parse(tokenStr)
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

func parseStringParam(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}
