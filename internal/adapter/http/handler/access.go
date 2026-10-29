package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/adapter/http/middleware"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/usecase"
)

// isSelfOrBoard reports whether the caller is the member memberID or at
// least Vorstand — the rule for reading another member's data (ML-010,
// SA-011, DS-003).
func isSelfOrBoard(r *http.Request, memberID uuid.UUID) bool {
	return memberID == middleware.GetUserID(r.Context()) ||
		domain.HasRole(middleware.GetUserRole(r.Context()), domain.RoleVorstand)
}

// viewerOf identifies the authenticated caller for per-viewer event views.
func viewerOf(r *http.Request) usecase.TimelineViewer {
	return usecase.TimelineViewer{
		ID:   middleware.GetUserID(r.Context()),
		Role: middleware.GetUserRole(r.Context()),
	}
}
