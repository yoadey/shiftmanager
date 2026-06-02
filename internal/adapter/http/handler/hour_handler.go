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
// Accepts optional clubYearId; falls back to the active club year.
func (h *HourHandler) ManualBooking(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID    uuid.UUID              `json:"memberId"`
		ClubYearID  uuid.UUID              `json:"clubYearId"`
		Hours       float64                `json:"hours"`
		Description string                 `json:"description"`
		Desc        string                 `json:"desc"`        // frontend alias
		Status      domain.HourEntryStatus `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Accept "desc" as an alias for "description" (frontend compatibility).
	if body.Description == "" && body.Desc != "" {
		body.Description = body.Desc
	}
	// If no clubYearId supplied, resolve from the active year.
	if body.ClubYearID == uuid.Nil {
		year, err := h.uc.GetActiveClubYear(r.Context())
		if err != nil {
			writeError(w, http.StatusBadRequest, "no active club year found; supply clubYearId")
			return
		}
		body.ClubYearID = year.ID
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

// GetMyAccount returns the authenticated member's hour account using the active club year.
// GET /api/v1/hours/me
func (h *HourHandler) GetMyAccount(w http.ResponseWriter, r *http.Request) {
	memberID := middleware.GetUserID(r.Context())
	if memberID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	full, err := h.uc.GetMemberAccountFull(r.Context(), memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, full)
}

// GetMemberAccountByID returns any member's hour account using the active club year.
// GET /api/v1/hours/{memberId}
func (h *HourHandler) GetMemberAccountByID(w http.ResponseWriter, r *http.Request) {
	memberID, err := parseUUIDParam(r, "memberId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid memberId")
		return
	}
	full, err := h.uc.GetMemberAccountFull(r.Context(), memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, full)
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

// ListClubYears returns all club years, newest first.
// GET /api/v1/hours/club-years
func (h *HourHandler) ListClubYears(w http.ResponseWriter, r *http.Request) {
	years, err := h.uc.ListClubYears(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, years)
}

// CreateClubYear creates a new club year.
// POST /api/v1/hours/club-years
func (h *HourHandler) CreateClubYear(w http.ResponseWriter, r *http.Request) {
	var body usecase.CreateClubYearInput
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Label == "" {
		writeError(w, http.StatusBadRequest, "label is required")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	year, err := h.uc.CreateClubYear(r.Context(), actorID, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, year)
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
