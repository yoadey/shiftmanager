package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// SettingsHandler handles application settings, branding, fee tiers, per-member
// fee-tier overrides, logo upload and audit log.
type SettingsHandler struct {
	uc         *usecase.SettingsUsecase
	templates  *usecase.EmailTemplateUsecase
	uploadDir  string
	publicBase string // public base URL used to build served logo URLs
}

// NewSettingsHandler creates a new SettingsHandler. templates may be nil if email
// template CRUD is not wired; uploadDir/publicBase configure logo upload (B-004).
func NewSettingsHandler(uc *usecase.SettingsUsecase, templates *usecase.EmailTemplateUsecase, uploadDir, publicBase string) *SettingsHandler {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	return &SettingsHandler{uc: uc, templates: templates, uploadDir: uploadDir, publicBase: publicBase}
}

// GetSettings returns the current application settings.
// GET /api/v1/settings
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// UpdateSettings persists changed application settings.
// PUT /api/v1/settings
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body domain.AppSettings
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	s, err := h.uc.UpdateSettings(r.Context(), actorID, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// GetBranding returns the current branding configuration.
// GET /api/v1/settings/branding
func (h *SettingsHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	b, err := h.uc.GetBranding(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// UpdateBranding persists a new branding configuration.
// PUT /api/v1/settings/branding
func (h *SettingsHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	var body domain.BrandingConfig
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	res, err := h.uc.UpdateBranding(r.Context(), actorID, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// GetFeeTiers returns the fee tiers for a club year.
// GET /api/v1/settings/fee-tiers?clubYearId=...
func (h *SettingsHandler) GetFeeTiers(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}

	tiers, err := h.uc.GetFeeTiers(r.Context(), clubYearID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

// UpdateFeeTiers replaces all fee tiers for a club year.
// PUT /api/v1/settings/fee-tiers?clubYearId=...
func (h *SettingsHandler) UpdateFeeTiers(w http.ResponseWriter, r *http.Request) {
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}

	var body struct {
		Tiers []*domain.FeeTier `json:"tiers"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	tiers, err := h.uc.UpdateFeeTiers(r.Context(), actorID, clubYearID, body.Tiers)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

// GetAuditLog returns filtered audit log entries.
// GET /api/v1/settings/audit
func (h *SettingsHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	filter := port.AuditFilter{
		Entity:   r.URL.Query().Get("entity"),
		EntityID: r.URL.Query().Get("entityId"),
		Limit:    parseIntQuery(r, "limit", 100),
		Offset:   parseIntQuery(r, "offset", 0),
	}

	if a := r.URL.Query().Get("actorId"); a != "" {
		if id, err := uuid.Parse(a); err == nil {
			filter.ActorID = &id
		}
	}
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &t
		}
	}

	entries, err := h.uc.GetAuditLog(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// GetMemberFeeTiers returns the per-member fee tier overrides for a club year (G-004).
// GET /api/v1/settings/members/{id}/fee-tiers?clubYearId=...
func (h *SettingsHandler) GetMemberFeeTiers(w http.ResponseWriter, r *http.Request) {
	memberID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid member id")
		return
	}
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}
	tiers, err := h.uc.GetMemberFeeTiers(r.Context(), memberID, clubYearID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

// UpdateMemberFeeTiers replaces the per-member fee tier overrides (G-004).
// PUT /api/v1/settings/members/{id}/fee-tiers?clubYearId=...
func (h *SettingsHandler) UpdateMemberFeeTiers(w http.ResponseWriter, r *http.Request) {
	memberID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid member id")
		return
	}
	clubYearID, err := uuid.Parse(r.URL.Query().Get("clubYearId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "clubYearId is required")
		return
	}
	var body struct {
		Tiers []*domain.FeeTier `json:"tiers"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	tiers, err := h.uc.UpdateMemberFeeTiers(r.Context(), actorID, memberID, clubYearID, body.Tiers)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tiers)
}

// UploadLogo accepts a PNG/SVG logo upload, stores it under the uploads dir and
// sets branding.logo_url to the served path (B-004).
// POST /api/v1/settings/logo  (multipart/form-data, field "file")
func (h *SettingsHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	const maxSize = 2 << 20 // 2MB
	if err := r.ParseMultipartForm(maxSize); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	if hdr.Size > maxSize {
		writeError(w, http.StatusRequestEntityTooLarge, "file exceeds 2MB limit")
		return
	}

	ext := strings.ToLower(filepath.Ext(hdr.Filename))
	contentType := hdr.Header.Get("Content-Type")
	if !isAllowedLogo(ext, contentType) {
		writeError(w, http.StatusUnsupportedMediaType, "only PNG and SVG files are allowed")
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "cannot create upload dir")
		return
	}
	name := "logo-" + uuid.New().String() + ext
	dst := filepath.Join(h.uploadDir, name)
	out, err := os.Create(dst)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot store file")
		return
	}
	if _, err := io.Copy(out, io.LimitReader(file, maxSize)); err != nil {
		_ = out.Close()
		writeError(w, http.StatusInternalServerError, "cannot write file")
		return
	}
	_ = out.Close()

	logoURL := strings.TrimRight(h.publicBase, "/") + "/uploads/" + name
	actorID := middleware.GetUserID(r.Context())
	b, err := h.uc.SetLogoURL(r.Context(), actorID, logoURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// isAllowedLogo validates the extension and (when present) content type.
func isAllowedLogo(ext, contentType string) bool {
	switch ext {
	case ".png":
		return contentType == "" || contentType == "image/png"
	case ".svg":
		return contentType == "" || contentType == "image/svg+xml" || strings.HasPrefix(contentType, "text/")
	default:
		return false
	}
}

// ListEmailTemplates returns all email templates (Section 4).
// GET /api/v1/settings/email-templates
func (h *SettingsHandler) ListEmailTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := h.templates.ListTemplates(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// GetEmailTemplate returns a single email template by name.
// GET /api/v1/settings/email-templates/{name}
func (h *SettingsHandler) GetEmailTemplate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	t, err := h.templates.GetTemplate(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// UpdateEmailTemplate upserts an email template.
// PUT /api/v1/settings/email-templates/{name}
func (h *SettingsHandler) UpdateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var body struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	t, err := h.templates.UpsertTemplate(r.Context(), actorID, name, body.Subject, body.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// GetEmailLog returns recent email-log entries (N-004).
// GET /api/v1/settings/email-log
func (h *SettingsHandler) GetEmailLog(w http.ResponseWriter, r *http.Request) {
	entries, err := h.templates.ListEmailLog(r.Context(), parseIntQuery(r, "limit", 100), parseIntQuery(r, "offset", 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// GetBrandingHistory returns recent branding configuration snapshots (B-008).
// GET /api/v1/settings/branding/history
func (h *SettingsHandler) GetBrandingHistory(w http.ResponseWriter, r *http.Request) {
	entries, err := h.uc.GetBrandingHistory(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// RollbackBranding restores a previous branding configuration snapshot (B-008).
// POST /api/v1/settings/branding/rollback/{id}
func (h *SettingsHandler) RollbackBranding(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	b, err := h.uc.RollbackBranding(r.Context(), actorID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// ResendEmail re-sends a logged email by id (N-004).
// POST /api/v1/settings/email-log/{id}/resend
func (h *SettingsHandler) ResendEmail(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	actorID := middleware.GetUserID(r.Context())
	if err := h.templates.ResendEmail(r.Context(), actorID, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resent"})
}
