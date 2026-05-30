package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/port"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// MemberHandler handles HTTP requests for member operations.
type MemberHandler struct {
	uc *usecase.MemberUsecase
}

// NewMemberHandler creates a new MemberHandler.
func NewMemberHandler(uc *usecase.MemberUsecase) *MemberHandler {
	return &MemberHandler{uc: uc}
}

// List returns a paginated list of members.
// GET /api/v1/members
func (h *MemberHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := port.MemberFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
		Search: r.URL.Query().Get("search"),
	}

	if active := r.URL.Query().Get("active"); active != "" {
		b := active == "true"
		filter.IsActive = &b
	}

	members, err := h.uc.ListMembers(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// Create creates a new member.
// POST /api/v1/members
func (h *MemberHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FirstName           string     `json:"firstName"`
		LastName            string     `json:"lastName"`
		Email               string     `json:"email"`
		JoinedAt            *time.Time `json:"joinedAt"`
		IndividualGoalHours *float64   `json:"individualGoalHours"`
		Role                string     `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	m, err := h.uc.CreateMember(r.Context(), actorID, usecase.CreateMemberInput{
		FirstName:           body.FirstName,
		LastName:            body.LastName,
		Email:               body.Email,
		JoinedAt:            body.JoinedAt,
		IndividualGoalHours: body.IndividualGoalHours,
		Role:                body.Role,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "member email already in use" {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// Get returns a single member by ID.
// GET /api/v1/members/:id
func (h *MemberHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid member id")
		return
	}

	m, err := h.uc.GetMember(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "member not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// Update applies changes to a member.
// PUT /api/v1/members/:id
func (h *MemberHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid member id")
		return
	}

	var body struct {
		FirstName           string   `json:"firstName"`
		LastName            string   `json:"lastName"`
		Email               string   `json:"email"`
		IndividualGoalHours *float64 `json:"individualGoalHours"`
		Role                string   `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	m, err := h.uc.UpdateMember(r.Context(), actorID, id, usecase.UpdateMemberInput{
		FirstName:           body.FirstName,
		LastName:            body.LastName,
		Email:               body.Email,
		IndividualGoalHours: body.IndividualGoalHours,
		Role:                body.Role,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// Deactivate soft-deletes a member.
// DELETE /api/v1/members/:id
func (h *MemberHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid member id")
		return
	}

	actorID := middleware.GetUserID(r.Context())
	if err := h.uc.DeactivateMember(r.Context(), actorID, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Import handles CSV member import.
// POST /api/v1/members/import
func (h *MemberHandler) Import(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	actorID := middleware.GetUserID(r.Context())
	preview := r.URL.Query().Get("preview") == "true"

	if preview {
		result, err := h.uc.ImportCSVPreview(r.Context(), file)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	result, err := h.uc.ImportCSV(r.Context(), actorID, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Export returns all active members as a CSV file.
// GET /api/v1/members/export
func (h *MemberHandler) Export(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r.Context())
	data, err := h.uc.ExportCSV(r.Context(), actorID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"members.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// helper functions

func parseUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

func parseIntQuery(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// contextKey re-export so handler package is self-contained.
type ctxKey string

// withContext is a utility to attach a value to a context.
func withContext(ctx context.Context, key ctxKey, val interface{}) context.Context {
	return context.WithValue(ctx, key, val)
}
