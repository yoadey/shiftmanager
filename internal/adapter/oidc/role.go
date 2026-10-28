package oidc

import (
	"errors"
	"fmt"

	"github.com/jmespath/go-jmespath"
	"github.com/yoadey/shiftmanager/internal/domain"
)

// ErrInvalidRoleExpression is returned by New when RoleAttributePath does not
// compile. Unlike a failed discovery it is a configuration error that a retry
// cannot fix.
var ErrInvalidRoleExpression = errors.New("invalid OIDC role attribute path")

// RoleMapper evaluates a JMESPath expression against a provider's claims to
// derive the application role, in the style of Grafana's role_attribute_path:
//
//	contains(groups[*], 'shiftmanager-admin') && 'admin' || contains(groups[*], 'vorstand') && 'vorstand' || 'mitglied'
//
// The expression must yield one of the role names (kiosk, mitglied,
// veranstaltungsleiter, vorstand, admin; case-insensitive). Anything else —
// null, an unknown name, a non-string or an evaluation error — means no role.
type RoleMapper struct {
	expr *jmespath.JMESPath
}

// NewRoleMapper compiles path. An empty path yields a nil mapper, which
// never resolves a role.
func NewRoleMapper(path string) (*RoleMapper, error) {
	if path == "" {
		return nil, nil
	}
	expr, err := jmespath.Compile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRoleExpression, err)
	}
	return &RoleMapper{expr: expr}, nil
}

// Resolve returns the role the expression selects from claims, or "".
func (m *RoleMapper) Resolve(claims map[string]any) string {
	if m == nil || claims == nil {
		return ""
	}
	v, err := m.expr.Search(claims)
	if err != nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return domain.ParseRole(s)
}
