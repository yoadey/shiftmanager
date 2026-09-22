package db

import "github.com/google/uuid"

// mustParseUUID parses a UUID string and panics if it is invalid.
// This should only be called with values that were previously stored by this package.
func mustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}
