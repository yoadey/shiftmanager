package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Member represents a club member.
type Member struct {
	ID                  uuid.UUID  `json:"id"`
	FirstName           string     `json:"firstName"`
	LastName            string     `json:"lastName"`
	Email               string     `json:"email"`
	JoinedAt            time.Time  `json:"joinedAt"`
	LeftAt              *time.Time `json:"leftAt,omitempty"`
	IsActive            bool       `json:"isActive"`
	IndividualGoalHours *float64   `json:"individualGoalHours,omitempty"`
	OIDCSubject         *string    `json:"oidcSubject,omitempty"`
	Role                string     `json:"role"`
}

// OIDCLink stores the link between a member and an OIDC provider subject.
type OIDCLink struct {
	ID       uuid.UUID `json:"id"`
	MemberID uuid.UUID `json:"memberId"`
	Provider string    `json:"provider"`
	Subject  string    `json:"subject"`
	LinkedAt time.Time `json:"linkedAt"`
}

// NameMode controls how member names are displayed.
type NameMode string

const (
	NameModeAbbrev NameMode = "abbrev"
	NameModeFull   NameMode = "full"
)

// Roles defines RBAC roles in ascending privilege order.
const (
	RoleKiosk               = "kiosk"
	RoleMitglied            = "mitglied"
	RoleVeranstaltungsleiter = "veranstaltungsleiter"
	RoleVorstand            = "vorstand"
	RoleAdmin               = "admin"
)

// RoleLevel maps roles to numeric levels for comparison.
var RoleLevel = map[string]int{
	RoleKiosk:               0,
	RoleMitglied:            1,
	RoleVeranstaltungsleiter: 2,
	RoleVorstand:            3,
	RoleAdmin:               4,
}

// HasRole returns true when the given role meets the minimum required role.
func HasRole(userRole, requiredRole string) bool {
	userLevel, ok := RoleLevel[userRole]
	if !ok {
		return false
	}
	reqLevel, ok := RoleLevel[requiredRole]
	if !ok {
		return false
	}
	return userLevel >= reqLevel
}

// DisplayName returns the member's name according to the given mode.
func (m *Member) DisplayName(mode NameMode) string {
	switch mode {
	case NameModeAbbrev:
		return m.Abbreviate()
	default:
		return m.FullName()
	}
}

// DisplayNameFor returns the member's name as it should be shown to the given viewer.
// Per NM-002 the full name is always returned when:
//   - the viewer is looking at their own name (viewerID == member.ID), or
//   - the viewer is a board member or higher (vorstand/admin),
//   - or the configured name mode is "full".
// Otherwise the abbreviated form ("Maximilian M.") is returned.
func (m *Member) DisplayNameFor(mode NameMode, viewerID uuid.UUID, viewerRole string) string {
	if mode == NameModeFull {
		return m.FullName()
	}
	if viewerID == m.ID {
		return m.FullName()
	}
	if HasRole(viewerRole, RoleVorstand) {
		return m.FullName()
	}
	return m.Abbreviate()
}

// FullName returns "FirstName LastName".
func (m *Member) FullName() string {
	return fmt.Sprintf("%s %s", m.FirstName, m.LastName)
}

// Abbreviate returns "FirstName L." where L is the first letter of the last name
// (NM-002). For example "Maximilian Müller" becomes "Maximilian M.".
func (m *Member) Abbreviate() string {
	if len(m.LastName) == 0 {
		return m.FirstName
	}
	initial := strings.ToUpper(string([]rune(m.LastName)[0]))
	if len(m.FirstName) == 0 {
		return fmt.Sprintf("%s.", initial)
	}
	return fmt.Sprintf("%s %s.", m.FirstName, initial)
}

// Activate marks the member as active.
func (m *Member) Activate() {
	m.IsActive = true
	m.LeftAt = nil
}

// Deactivate marks the member as inactive and records when they left.
func (m *Member) Deactivate(at time.Time) {
	m.IsActive = false
	m.LeftAt = &at
}

// ErrMemberNotFound is returned when a member cannot be found.
var ErrMemberNotFound = fmt.Errorf("member not found")

// ErrMemberEmailConflict is returned when a member email already exists.
var ErrMemberEmailConflict = fmt.Errorf("member email already in use")

// ErrMemberInactive is returned when operating on an inactive member.
var ErrMemberInactive = fmt.Errorf("member is inactive")

// CSVImportRow represents a single row from a CSV import file.
type CSVImportRow struct {
	FirstName           string
	LastName            string
	Email               string
	JoinedAt            *time.Time
	IndividualGoalHours *float64
}

// CSVImportPreview describes what would change during a CSV import.
type CSVImportPreview struct {
	ToCreate []CSVImportRow `json:"toCreate"`
	ToUpdate []CSVImportRow `json:"toUpdate"`
	ToSkip   []CSVImportRow `json:"toSkip"`
}
