package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
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
	kioskUC    *usecase.KioskUsecase
}

// NewKioskHandler creates a new KioskHandler. settingsUC may be nil to disable
// the kiosk-lock check; kioskUC manages and validates kiosk links (K-016).
func NewKioskHandler(regUC *usecase.RegistrationUsecase, eventUC *usecase.EventUsecase, memberUC *usecase.MemberUsecase, settingsUC *usecase.SettingsUsecase, kioskUC *usecase.KioskUsecase) *KioskHandler {
	return &KioskHandler{regUC: regUC, eventUC: eventUC, memberUC: memberUC, settingsUC: settingsUC, kioskUC: kioskUC}
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

	// Anonymous viewer: names only, per name mode (SC-011).
	timeline, err := h.eventUC.GetEventTimelineFor(r.Context(), id, usecase.TimelineViewer{})
	if errors.Is(err, domain.ErrEventNotFound) {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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
		Email      string     `json:"email"` // frontend alias for guestEmail
		Name       string     `json:"name"`  // full or display name (K-014, K-006)
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

	// Picking a member is only possible through the member search (K-002).
	if body.MemberID != nil && !h.kioskSearchEnabled(r) {
		writeError(w, http.StatusForbidden, "kiosk member search is disabled; register with your name")
		return
	}
	if body.MemberID == nil && strings.TrimSpace(body.Name) == "" {
		writeError(w, http.StatusBadRequest, "name or memberId is required")
		return
	}
	if email := strings.TrimSpace(body.GuestEmail); email != "" && !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}

	_, err = h.regUC.KioskRegister(r.Context(), usecase.KioskRegisterInput{
		ShiftID:  shiftID,
		MemberID: body.MemberID,
		Name:     body.Name,
		Email:    body.GuestEmail,
		Comment:  body.Comment,
	})
	switch {
	case errors.Is(err, usecase.ErrInvalidKioskName):
		writeError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, domain.ErrAlreadyRegistered):
		writeError(w, http.StatusConflict, err.Error())
		return
	case errors.Is(err, domain.ErrShiftNotFound), errors.Is(err, domain.ErrEventNotFound):
		writeError(w, http.StatusNotFound, "shift not found")
		return
	case err != nil:
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	// The same answer whether or not the e-mail belongs to a member (K-005).
	writeJSON(w, http.StatusCreated, map[string]string{"message": "registered"})
}

// ListEvents returns all published events for the kiosk selection screen.
// GET /api/v1/kiosk/events
func (h *KioskHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	if h.locked(w, r) {
		return
	}
	published := "published"
	events, _, err := h.eventUC.ListEvents(r.Context(), port.EventFilter{Status: (*domain.EventStatus)(&published)})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// kioskSearchEnabled reports whether the admin allows the kiosk member
// search (K-002); without settings (tests) it is allowed.
func (h *KioskHandler) kioskSearchEnabled(r *http.Request) bool {
	if h.settingsUC == nil {
		return true
	}
	st, err := h.settingsUC.GetSettings(r.Context())
	return err == nil && st.KioskSearch
}

// kioskSearchLimit caps the anonymous kiosk member search.
const kioskSearchLimit = 20

// SearchMembers returns active members matching a search query for kiosk lookup.
// GET /api/v1/kiosk/members?q=...
func (h *KioskHandler) SearchMembers(w http.ResponseWriter, r *http.Request) {
	if h.locked(w, r) {
		return
	}
	// With the search switched off the kiosk only takes e-mail addresses
	// (K-002, K-004); don't answer name queries behind the UI's back.
	if !h.kioskSearchEnabled(r) {
		writeError(w, http.StatusForbidden, "kiosk member search is disabled")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		writeError(w, http.StatusBadRequest, "query must be at least 2 characters")
		return
	}
	// Names only (no e-mail lookups) and a capped result, so an anonymous
	// caller can neither probe addresses nor page through the roster (K-003).
	members, _, err := h.memberUC.ListMembers(r.Context(), port.MemberFilter{
		Search:          q,
		SearchNamesOnly: true,
		Status:          port.MemberStatusActive,
		Limit:           kioskSearchLimit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// The kiosk is unauthenticated: only id and first name plus initial (K-003).
	out := make([]domain.MemberName, 0, len(members))
	for _, m := range members {
		out = append(out, domain.MemberName{ID: m.ID, Name: m.Abbreviate()})
	}
	writeJSON(w, http.StatusOK, out)
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
	result := map[string]string{"message": "confirmed"}
	if shift, err := h.eventUC.GetShift(r.Context(), reg.ShiftID); err == nil {
		result["shiftName"] = shift.Name
		if ev, err := h.eventUC.GetEvent(r.Context(), shift.EventID); err == nil {
			result["eventName"] = ev.Name
		}
	}
	writeJSON(w, http.StatusOK, result)
}

// Config returns what the kiosk screen needs without a login: whether the
// member search is offered and the club's name and logo.
// GET /api/v1/kiosk/config
func (h *KioskHandler) Config(w http.ResponseWriter, r *http.Request) {
	if h.locked(w, r) {
		return
	}
	cfg := map[string]any{"kioskSearch": h.kioskSearchEnabled(r)}
	if h.settingsUC != nil {
		if b, err := h.settingsUC.GetBranding(r.Context()); err == nil {
			cfg["clubName"] = b.ClubName
			cfg["logoUrl"] = b.LogoURL
			cfg["primaryColor"] = b.PrimaryColor
		}
	}
	writeJSON(w, http.StatusOK, cfg)
}

// TokenValid reports whether token unlocks the kiosk (K-016).
func (h *KioskHandler) TokenValid(r *http.Request, token string) bool {
	return h.kioskUC != nil && h.kioskUC.ValidToken(r.Context(), token)
}

// ListLinks returns all kiosk links (Vorstand+, K-016).
// GET /api/v1/settings/kiosk-links
func (h *KioskHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.kioskUC.ListLinks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, links)
}

// CreateLink creates a kiosk link; the token is only in this response.
// POST /api/v1/settings/kiosk-links {label, validDays}
func (h *KioskHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label     string `json:"label"`
		ValidDays int    `json:"validDays"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	link, token, err := h.kioskUC.CreateLink(r.Context(), middleware.GetUserID(r.Context()), body.Label, body.ValidDays)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, struct {
		*domain.KioskLink
		Token string `json:"token"`
	}{link, token})
}

// RevokeLink revokes a kiosk link.
// DELETE /api/v1/settings/kiosk-links/{id}
func (h *KioskHandler) RevokeLink(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.kioskUC.RevokeLink(r.Context(), middleware.GetUserID(r.Context()), id); err != nil {
		if errors.Is(err, domain.ErrKioskLinkNotFound) {
			writeError(w, http.StatusNotFound, "kiosk link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
