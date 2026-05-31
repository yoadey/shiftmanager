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

// EventHandler handles HTTP requests for event operations.
type EventHandler struct {
	uc *usecase.EventUsecase
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(uc *usecase.EventUsecase) *EventHandler {
	return &EventHandler{uc: uc}
}

// List returns events filtered by status, date range, etc.
// GET /api/v1/events
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := port.EventFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.EventStatus(s)
		filter.Status = &st
	}
	if v := r.URL.Query().Get("visibility"); v != "" {
		vis := domain.EventVisibility(v)
		filter.Visibility = &vis
	}
	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err == nil {
			filter.FromDate = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err == nil {
			filter.ToDate = &t
		}
	}

	events, err := h.uc.ListEvents(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// Create creates a new event in draft state.
// POST /api/v1/events
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string                  `json:"name"`
		Description string                  `json:"description"`
		Location    string                  `json:"location"`
		Category    string                  `json:"category"`
		StartDate   time.Time               `json:"startDate"`
		EndDate     time.Time               `json:"endDate"`
		Visibility  domain.EventVisibility  `json:"visibility"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	e, err := h.uc.CreateEvent(r.Context(), actorID, usecase.CreateEventInput{
		Name:        body.Name,
		Description: body.Description,
		Location:    body.Location,
		Category:    body.Category,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
		Visibility:  body.Visibility,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// GetWithTimeline returns a single event with its full timeline.
// GET /api/v1/events/:id
func (h *EventHandler) GetWithTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	timeline, err := h.uc.GetEventTimeline(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}

// Update applies changes to an event.
// PUT /api/v1/events/:id
func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var body struct {
		Name        string                  `json:"name"`
		Description string                  `json:"description"`
		Location    string                  `json:"location"`
		Category    string                  `json:"category"`
		StartDate   *time.Time              `json:"startDate"`
		EndDate     *time.Time              `json:"endDate"`
		Visibility  domain.EventVisibility  `json:"visibility"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	e, err := h.uc.UpdateEvent(r.Context(), actorID, id, usecase.UpdateEventInput{
		Name:        body.Name,
		Description: body.Description,
		Location:    body.Location,
		Category:    body.Category,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
		Visibility:  body.Visibility,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// Delete removes an event (only draft/cancelled).
// DELETE /api/v1/events/:id
func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	if err := h.uc.DeleteEvent(r.Context(), actorID, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Publish transitions an event from draft to published.
// POST /api/v1/events/:id/publish
func (h *EventHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	e, err := h.uc.PublishEvent(r.Context(), actorID, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// GetTimeline returns just the timeline for an event (alias).
// GET /api/v1/events/:id/timeline
func (h *EventHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	h.GetWithTimeline(w, r)
}

// unused import guard
var _ uuid.UUID
