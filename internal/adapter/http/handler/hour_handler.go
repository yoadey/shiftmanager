package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// HourHandler handles HTTP requests for hour-tracking operations.
type HourHandler struct {
	uc *usecase.HourUsecase
}

// NewHourHandler creates a new HourHandler.
func NewHourHandler(uc *usecase.HourUsecase) *HourHandler {
	return &HourHandler{uc: uc}
}

// ConfirmShiftHours records confirmed hours for a member after a shift.
// POST /api/v1/hours/confirm
func (h *HourHandler) ConfirmShiftHours(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID    uuid.UUID `json:"memberId"`
		ShiftID     uuid.UUID `json:"shiftId"`
		ClubYearID  uuid.UUID `json:"clubYearId"`
		Hours       float64   `json:"hours"`
		Description string    `json:"description"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	entry, err := h.uc.ConfirmShiftHours(r.Context(), actorID, body.MemberID, body.ShiftID, body.ClubYearID, body.Hours, body.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

// ManualBooking creates a manual hour entry (board only).
// POST /api/v1/hours/manual
func (h *HourHandler) ManualBooking(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID    uuid.UUID              `json:"memberId"`
		ClubYearID  uuid.UUID              `json:"clubYearId"`
		Hours       float64                `json:"hours"`
		Description string                 `json:"description"`
		Status      domain.HourEntryStatus `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	entry, err := h.uc.ManualBooking(r.Context(), actorID, usecase.ManualBookingInput{
		MemberID:    body.MemberID,
		ClubYearID:  body.ClubYearID,
		Hours:       body.Hours,
		Description: body.Description,
		Status:      body.Status,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

// CorrectEntry updates an existing hour entry.
// PUT /api/v1/hours/:id
func (h *HourHandler) CorrectEntry(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	var body struct {
		Hours       float64 `json:"hours"`
		Description string  `json:"description"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	entry, err := h.uc.CorrectEntry(r.Context(), actorID, id, body.Hours, body.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

// DeleteEntry removes an hour entry.
// DELETE /api/v1/hours/:id
func (h *HourHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	if err := h.uc.DeleteEntry(r.Context(), actorID, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetMemberAccount returns a member's hour account for a club year.
// GET /api/v1/hours/account?memberId=...&clubYearId=...
func (h *HourHandler) GetMemberAccount(w http.ResponseWriter, r *http.Request) {
	memberID, err := uuid.Parse(r.URL.Query().Get("memberId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "memberId is required")
		return
	}
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}

	account, err := h.uc.GetMemberAccount(r.Context(), memberID, clubYearID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, account)
}

// GetYearSummary returns the hour summary for all active members for a year.
// GET /api/v1/hours/summary?clubYearId=...
func (h *HourHandler) GetYearSummary(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}

	rows, err := h.uc.GetYearSummary(r.Context(), clubYearID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
