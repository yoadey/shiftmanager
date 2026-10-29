package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameSimilarity(t *testing.T) {
	high := []struct{ a, b string }{
		{"Max Mustermann", "Max Mustermann"},
		{"Max Mustermann", "max mustermann"},
		{"Mustermann Max", "Max Mustermann"},
		{"Mustermann, Max", "Max Mustermann"},
		{"Max Mustermann", "Maximilian Mustermann"},
		{"Max Muller", "Maximilian Müller"},
		{"Jörg Mueller", "Joerg Müller"},
		{"Max Musterman", "Max Mustermann"},
		{"René Lefèvre", "Rene Lefevre"},
	}
	for _, c := range high {
		assert.GreaterOrEqual(t, NameSimilarity(c.a, c.b), 0.75, "%q vs %q", c.a, c.b)
	}
	low := []struct{ a, b string }{
		{"Max Mustermann", "Erika Schmidt"},
		{"Anna Bauer", "Peter Klein"},
	}
	for _, c := range low {
		assert.Less(t, NameSimilarity(c.a, c.b), 0.5, "%q vs %q", c.a, c.b)
	}
	assert.Equal(t, 1.0, NameSimilarity("Max Mustermann", "Max Mustermann"))
	assert.Equal(t, 0.0, NameSimilarity("", "Max"))
	// A close match outranks a loose one.
	assert.Greater(t, NameSimilarity("Max Müller", "Max Mueller"), NameSimilarity("Max Müller", "Moritz Möller"))
}

func TestNameSimilarity_PartialNames(t *testing.T) {
	// A bare first name is a lead (shown), but below a full match.
	s := NameSimilarity("Erika", "Erika Schmidt")
	assert.InDelta(t, 0.75, s, 0.01)
	assert.Less(t, s, NameSimilarity("Erika Schmit", "Erika Schmidt"))
	// A middle name doesn't sink a match.
	assert.GreaterOrEqual(t, NameSimilarity("Hans Müller", "Hans Peter Müller"), 0.75)
}
