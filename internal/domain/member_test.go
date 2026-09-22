package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMember_Abbreviate(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{"NM-002 standard", "Maximilian", "Müller", "Maximilian M."},
		{"simple", "Anna", "Schmidt", "Anna S."},
		{"empty last name", "Anna", "", "Anna"},
		{"empty first name", "", "Müller", "M."},
		{"both empty", "", "", ""},
		{"lowercase last name initial uppercased", "Max", "müller", "Max M."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Member{FirstName: tt.firstName, LastName: tt.lastName}
			assert.Equal(t, tt.want, m.Abbreviate())
		})
	}
}

func TestMember_FullName(t *testing.T) {
	m := &Member{FirstName: "Maximilian", LastName: "Müller"}
	assert.Equal(t, "Maximilian Müller", m.FullName())
}

func TestMember_DisplayName(t *testing.T) {
	m := &Member{FirstName: "Maximilian", LastName: "Müller"}
	assert.Equal(t, "Maximilian M.", m.DisplayName(NameModeAbbrev))
	assert.Equal(t, "Maximilian Müller", m.DisplayName(NameModeFull))
	assert.Equal(t, "Maximilian Müller", m.DisplayName(NameMode("unknown")))
}

func TestMember_DisplayNameFor(t *testing.T) {
	memberID := uuid.New()
	otherID := uuid.New()
	m := &Member{ID: memberID, FirstName: "Maximilian", LastName: "Müller"}

	tests := []struct {
		name       string
		mode       NameMode
		viewerID   uuid.UUID
		viewerRole string
		want       string
	}{
		{"full mode always full", NameModeFull, otherID, RoleMitglied, "Maximilian Müller"},
		{"own name full", NameModeAbbrev, memberID, RoleMitglied, "Maximilian Müller"},
		{"board sees full", NameModeAbbrev, otherID, RoleVorstand, "Maximilian Müller"},
		{"admin sees full", NameModeAbbrev, otherID, RoleAdmin, "Maximilian Müller"},
		{"other member sees abbrev", NameModeAbbrev, otherID, RoleMitglied, "Maximilian M."},
		{"veranstaltungsleiter sees abbrev", NameModeAbbrev, otherID, RoleVeranstaltungsleiter, "Maximilian M."},
		{"kiosk sees abbrev", NameModeAbbrev, otherID, RoleKiosk, "Maximilian M."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, m.DisplayNameFor(tt.mode, tt.viewerID, tt.viewerRole))
		})
	}
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		userRole string
		reqRole  string
		want     bool
	}{
		{RoleAdmin, RoleVorstand, true},
		{RoleVorstand, RoleVorstand, true},
		{RoleMitglied, RoleVorstand, false},
		{RoleKiosk, RoleMitglied, false},
		{RoleVeranstaltungsleiter, RoleMitglied, true},
		{"unknown", RoleMitglied, false},
		{RoleAdmin, "unknown", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, HasRole(tt.userRole, tt.reqRole), "%s>=%s", tt.userRole, tt.reqRole)
	}
}

func TestMember_ActivateDeactivate(t *testing.T) {
	m := &Member{IsActive: false}
	m.Activate()
	assert.True(t, m.IsActive)
	assert.Nil(t, m.LeftAt)

	now := time.Now().UTC()
	m.Deactivate(now)
	assert.False(t, m.IsActive)
	if assert.NotNil(t, m.LeftAt) {
		assert.Equal(t, now, *m.LeftAt)
	}
}
