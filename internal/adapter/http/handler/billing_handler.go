package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// BillingHandler handles year-end billing computation and export.
type BillingHandler struct {
	uc *usecase.BillingUsecase
}

// NewBillingHandler creates a new BillingHandler.
func NewBillingHandler(uc *usecase.BillingUsecase) *BillingHandler {
	return &BillingHandler{uc: uc}
}

// Compute calculates billing for all active members for a club year.
// GET /api/v1/billing/:clubYearId
func (h *BillingHandler) Compute(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := parseUUIDParam(r, "clubYearId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid club year id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	report, err := h.uc.ComputeYearBilling(r.Context(), actorID, clubYearID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// ExportCSV returns the billing report as a CSV download.
// GET /api/v1/billing/:clubYearId/export.csv
func (h *BillingHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := parseUUIDParam(r, "clubYearId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid club year id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	data, err := h.uc.ExportBillingCSV(r.Context(), actorID, clubYearID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"billing.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ExportPDF returns the billing report as a PDF (currently CSV-backed).
// GET /api/v1/billing/:clubYearId/export.pdf
func (h *BillingHandler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := parseUUIDParam(r, "clubYearId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid club year id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	data, contentType, err := h.uc.ExportBillingPDF(r.Context(), actorID, clubYearID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\"billing.pdf\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// unused import guard
var _ = uuid.Nil
