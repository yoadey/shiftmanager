package domain

import (
	"math"
	"strconv"
	"strings"
)

// ContrastRatio computes the WCAG 2.x contrast ratio between two colours given as
// hex strings (e.g. "#000000" or "fff"). The ratio is in the range [1, 21].
// Invalid colours are treated as black. The ratio is symmetric.
func ContrastRatio(hexA, hexB string) float64 {
	la := relativeLuminance(hexA)
	lb := relativeLuminance(hexB)
	lighter, darker := la, lb
	if darker > lighter {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05)
}

// MeetsAAContrast reports whether the contrast ratio meets the WCAG AA threshold
// for normal text (4.5:1).
func MeetsAAContrast(hexA, hexB string) bool {
	return ContrastRatio(hexA, hexB) >= 4.5
}

// BestContrastAgainstBlackWhite returns the larger contrast ratio of the given
// colour against pure black and pure white. This is the relevant figure when a
// foreground colour may be placed on either a light or dark surface.
func BestContrastAgainstBlackWhite(hex string) float64 {
	black := ContrastRatio(hex, "#000000")
	white := ContrastRatio(hex, "#FFFFFF")
	return math.Max(black, white)
}

// relativeLuminance computes the WCAG relative luminance of a hex colour.
func relativeLuminance(hex string) float64 {
	r, g, b := parseHexColor(hex)
	rl := linearize(float64(r) / 255.0)
	gl := linearize(float64(g) / 255.0)
	bl := linearize(float64(b) / 255.0)
	return 0.2126*rl + 0.7152*gl + 0.0722*bl
}

// linearize converts an sRGB channel value [0,1] to linear light.
func linearize(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// parseHexColor parses a #RRGGBB or #RGB hex string into 8-bit RGB components.
// Unparsable input yields black (0,0,0).
func parseHexColor(hex string) (uint8, uint8, uint8) {
	h := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 {
		return 0, 0, 0
	}
	r, err1 := strconv.ParseUint(h[0:2], 16, 8)
	g, err2 := strconv.ParseUint(h[2:4], 16, 8)
	b, err3 := strconv.ParseUint(h[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0
	}
	return uint8(r), uint8(g), uint8(b)
}
