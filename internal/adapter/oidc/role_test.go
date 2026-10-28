package oidc

import (
	"errors"
	"testing"
)

func TestRoleMapper(t *testing.T) {
	const grafanaStyle = "contains(groups[*], 'sm-admin') && 'admin' || contains(groups[*], 'vorstand') && 'Vorstand' || 'mitglied'"
	cases := []struct {
		name   string
		path   string
		claims map[string]any
		want   string
	}{
		{"admin group", grafanaStyle, map[string]any{"groups": []any{"x", "sm-admin"}}, "admin"},
		{"case-insensitive role", grafanaStyle, map[string]any{"groups": []any{"vorstand"}}, "vorstand"},
		{"fallback role", grafanaStyle, map[string]any{"groups": []any{"other"}}, "mitglied"},
		{"missing claim errors to no role", grafanaStyle, map[string]any{}, ""},
		{"nested keycloak claim", "resource_access.shiftmanager.roles[0]",
			map[string]any{"resource_access": map[string]any{"shiftmanager": map[string]any{"roles": []any{"veranstaltungsleiter"}}}}, "veranstaltungsleiter"},
		{"unknown role", "role", map[string]any{"role": "Viewer"}, ""},
		{"non-string result", "groups", map[string]any{"groups": []any{"admin"}}, ""},
		{"null result", "role", map[string]any{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := NewRoleMapper(tc.path)
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			if got := m.Resolve(tc.claims); got != tc.want {
				t.Errorf("Resolve = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRoleMapper_EmptyPathNeverResolves(t *testing.T) {
	m, err := NewRoleMapper("")
	if err != nil || m != nil {
		t.Fatalf("want nil mapper, got %v, %v", m, err)
	}
	if got := m.Resolve(map[string]any{"role": "admin"}); got != "" {
		t.Errorf("nil mapper resolved %q", got)
	}
}

func TestRoleMapper_InvalidExpression(t *testing.T) {
	if _, err := NewRoleMapper("contains(groups[*], "); !errors.Is(err, ErrInvalidRoleExpression) {
		t.Fatalf("want ErrInvalidRoleExpression, got %v", err)
	}
}
