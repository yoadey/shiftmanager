package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// EventHandler handles HTTP requests for event operations.
type EventHandler struct {
	uc      *usecase.EventUsecase
	storage port.MediaStorage // event attachment upload (V-008)
}

// NewEventHandler creates a new EventHandler. storage configures where event
// attachment uploads (V-008) are written, mirroring the logo upload in
// SettingsHandler.
func NewEventHandler(uc *usecase.EventUsecase, storage port.MediaStorage) *EventHandler {
	return &EventHandler{uc: uc, storage: storage}
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
		Status      domain.EventStatus      `json:"status"`
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
		Status:      body.Status,
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
		Status      domain.EventStatus      `json:"status"`
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
		Status:      body.Status,
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
	deletedAttachments, err := h.uc.DeleteEvent(r.Context(), actorID, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	for _, a := range deletedAttachments {
		_ = h.storage.Delete(r.Context(), a.URL)
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

// Complete transitions a published event to completed and confirms all registered shifts.
// POST /api/v1/events/:id/complete
func (h *EventHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	e, err := h.uc.CompleteEvent(r.Context(), actorID, id)
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

// CopyEvent creates a copy of an event including all its shifts.
// POST /api/v1/events/{id}/copy
func (h *EventHandler) CopyEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	event, err := h.uc.CopyEvent(r.Context(), actorID, eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

// GenerateRecurrence turns an event into a recurring series (V-007),
// creating follow-up occurrences (weekly or monthly, up to "until") as
// full copies of the event including its shifts.
// POST /api/v1/events/{id}/recurrence
func (h *EventHandler) GenerateRecurrence(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var body struct {
		Frequency domain.RecurrenceFrequency `json:"frequency"`
		Until     time.Time                  `json:"until"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	events, err := h.uc.GenerateRecurrence(r.Context(), actorID, id, body.Frequency, body.Until)
	if err != nil {
		if err == domain.ErrEventNotFound {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, events)
}

// maxAttachmentSize is the per-file upload limit for event attachments (V-008).
const maxAttachmentSize = 5 << 20 // 5MB

// UploadAttachment accepts an image or document upload for an event, stores it
// under the uploads dir and records it against the event (V-008).
//
// SVG is deliberately not in the allow-list: unlike the club-logo upload
// (Vorstand-only), this endpoint is reachable by any Veranstaltungsleiter,
// and an SVG can embed a <script> that would execute in the app's own origin
// for anyone who opens the attachment — a stored-XSS path we don't want to
// open up at this broader privilege level.
// POST /api/v1/events/{id}/attachments  (multipart/form-data, field "file")
func (h *EventHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	fileURL, originalName, contentType, size, ok := receiveUpload(
		w, r, h.storage, maxAttachmentSize, "event-attach-",
		isAllowedEventAttachment, "only PNG, JPEG, GIF, WEBP and PDF files are allowed",
	)
	if !ok {
		return
	}

	actorID := middleware.GetUserID(r.Context())
	a, err := h.uc.AddAttachment(r.Context(), actorID, eventID, originalName, fileURL, contentType, size)
	if err != nil {
		_ = h.storage.Delete(r.Context(), fileURL)
		if err == domain.ErrEventNotFound {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

// isAllowedEventAttachment validates the extension and (when present) content type.
func isAllowedEventAttachment(ext, contentType string) bool {
	switch ext {
	case ".png":
		return contentType == "" || contentType == "image/png"
	case ".jpg", ".jpeg":
		return contentType == "" || contentType == "image/jpeg"
	case ".gif":
		return contentType == "" || contentType == "image/gif"
	case ".webp":
		return contentType == "" || contentType == "image/webp"
	case ".pdf":
		return contentType == "" || contentType == "application/pdf"
	default:
		return false
	}
}

// maxHeaderImageSize is the upload limit for an event's header image (V-010),
// matching the general attachment limit (V-008).
const maxHeaderImageSize = 5 << 20 // 5MB

// UploadHeaderImage sets an event's header image (V-010), a single image
// shown at the top of the event detail page, kept separate from the general
// attachments list (V-008) — replacing it (if one is already set) just
// overwrites the field, no cleanup of an old attachment entry needed.
//
// Image types only, no PDF: unlike UploadAttachment, this is specifically a
// picture slot, not a generic document upload. SVG stays excluded for the
// same stored-XSS reason as V-008.
// POST /api/v1/events/{id}/header-image  (multipart/form-data, field "file")
func (h *EventHandler) UploadHeaderImage(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	fileURL, _, _, _, ok := receiveUpload(
		w, r, h.storage, maxHeaderImageSize, "event-header-",
		isAllowedHeaderImage, "only PNG, JPEG, GIF and WEBP images are allowed",
	)
	if !ok {
		return
	}

	actorID := middleware.GetUserID(r.Context())
	ev, err := h.uc.SetHeaderImage(r.Context(), actorID, eventID, fileURL)
	if err != nil {
		_ = h.storage.Delete(r.Context(), fileURL)
		if err == domain.ErrEventNotFound {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ev)
}

// isAllowedHeaderImage validates the extension and (when present) content
// type — the same allow-list as isAllowedEventAttachment minus PDF.
func isAllowedHeaderImage(ext, contentType string) bool {
	switch ext {
	case ".png":
		return contentType == "" || contentType == "image/png"
	case ".jpg", ".jpeg":
		return contentType == "" || contentType == "image/jpeg"
	case ".gif":
		return contentType == "" || contentType == "image/gif"
	case ".webp":
		return contentType == "" || contentType == "image/webp"
	default:
		return false
	}
}

// DeleteHeaderImage clears an event's header image and, best-effort, removes
// the underlying file (V-010).
// DELETE /api/v1/events/{id}/header-image
func (h *EventHandler) DeleteHeaderImage(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	prevURL, err := h.uc.ClearHeaderImage(r.Context(), actorID, eventID)
	if err != nil {
		if err == domain.ErrEventNotFound {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if prevURL != "" {
		_ = h.storage.Delete(r.Context(), prevURL)
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListAttachments returns the files attached to an event (V-008).
// GET /api/v1/events/{id}/attachments
func (h *EventHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	list, err := h.uc.ListAttachments(r.Context(), eventID)
	if err != nil {
		if err == domain.ErrEventNotFound {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// DeleteAttachment removes an attachment from an event and, best-effort, its
// underlying file (V-008).
// DELETE /api/v1/events/{id}/attachments/{attachmentId}
func (h *EventHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	attachmentID, err := parseUUIDParam(r, "attachmentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	deleted, err := h.uc.DeleteAttachment(r.Context(), actorID, eventID, attachmentID)
	if err != nil {
		if err == domain.ErrEventAttachmentNotFound {
			writeError(w, http.StatusNotFound, "attachment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = h.storage.Delete(r.Context(), deleted.URL)

	w.WriteHeader(http.StatusNoContent)
}
