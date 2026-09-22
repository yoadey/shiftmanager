package domain

import (
	"math"
	"testing"
)

func approx(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

func TestContrastRatio_BlackWhite(t *testing.T) {
	r := ContrastRatio("#000000", "#FFFFFF")
	if !approx(r, 21.0, 0.01) {
		t.Fatalf("black/white contrast = %f, want 21", r)
	}
}

func TestContrastRatio_Symmetric(t *testing.T) {
	a := ContrastRatio("#123456", "#abcdef")
	b := ContrastRatio("#abcdef", "#123456")
	if !approx(a, b, 0.0001) {
		t.Fatalf("contrast not symmetric: %f vs %f", a, b)
	}
}

func TestContrastRatio_SameColor(t *testing.T) {
	r := ContrastRatio("#777777", "#777777")
	if !approx(r, 1.0, 0.0001) {
		t.Fatalf("identical colour contrast = %f, want 1", r)
	}
}

func TestContrastRatio_ShortHexAndInvalid(t *testing.T) {
	short := ContrastRatio("#fff", "#000")
	if !approx(short, 21.0, 0.01) {
		t.Fatalf("short hex contrast = %f, want 21", short)
	}
	// Invalid colour treated as black.
	inv := ContrastRatio("not-a-color", "#FFFFFF")
	if !approx(inv, 21.0, 0.01) {
		t.Fatalf("invalid colour contrast = %f, want 21", inv)
	}
}

func TestMeetsAAContrast(t *testing.T) {
	if !MeetsAAContrast("#000000", "#FFFFFF") {
		t.Fatal("black on white should meet AA")
	}
	// The club accent gold on white is low contrast.
	if MeetsAAContrast("#F4B63F", "#FFFFFF") {
		t.Fatal("gold on white should fail AA")
	}
}

func TestBestContrastAgainstBlackWhite(t *testing.T) {
	// Gold contrasts much better against black than white.
	best := BestContrastAgainstBlackWhite("#F4B63F")
	black := ContrastRatio("#F4B63F", "#000000")
	if !approx(best, black, 0.0001) {
		t.Fatalf("best = %f, want black contrast %f", best, black)
	}
}
