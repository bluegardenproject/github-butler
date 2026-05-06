// Package components holds small reusable widgets shared by views.
package components

import (
	"strings"

	"github.com/bluegardenproject/github-butler/internal/ui/theme"
)

// Banner renders the app title with a neon gradient.
func Banner(text string) string {
	return theme.Gradient(text, theme.TitleStops...)
}

// bigLetters is a 3-row block font covering just the characters we need
// for "GITHUB-BUTLER". Every glyph is 4 columns wide so letters line up
// cleanly when BigBanner joins them with a single-column separator.
//
// Each glyph uses Unicode quadrant and half blocks (U+2580…U+259F) so
// that within a 3-text-row height we can effectively draw at 6-pixel
// vertical resolution (two half-rows per text row) and 8-pixel
// horizontal resolution (two half-cols per text column). The corner
// quadrants (▟ ▙ ▜ ▛) give rounded outer corners; half-blocks (▀ ▄)
// let us draw single-pixel-tall horizontal strokes.
//
// If BigBanner is ever asked to render a rune not in this map it falls
// back to a solid block, so the banner still renders rather than
// panicking on unexpected input.
var bigLetters = map[rune][3]string{
	'G': {"▟██▙", "█ ▄▄", "▜▄▄▛"},
	'I': {"████", " ██ ", "████"},
	'T': {"████", " ██ ", " ██ "},
	'H': {"██ █", "████", "██ █"},
	'U': {"██ █", "██ █", "████"},
	'B': {"██▀▙", "██▀▙", "██▄▛"},
	'E': {"█▀▀▀", "█▀▀ ", "████"},
	'L': {"██  ", "██  ", "████"},
	'R': {"██▀▙", "██▄▛", "█ ▜▄"},
	'-': {"    ", "▄▄▄▄", "    "},
	' ': {"    ", "    ", "    "},
}

// BigBanner renders text as a 3-line block-letter banner, gradient-
// coloured with the shared title palette. Input is upper-cased before
// lookup so callers don't have to care about case.
//
// The banner is meant for the top of a view (currently only the
// dashboard uses it). Returned string already contains its own
// newlines; the caller just needs to place it with JoinVertical.
func BigBanner(text string) string {
	runes := []rune(strings.ToUpper(text))

	var rows [3]strings.Builder
	for i, r := range runes {
		glyph, ok := bigLetters[r]
		if !ok {
			glyph = [3]string{"████", "████", "████"}
		}
		for row := 0; row < 3; row++ {
			rows[row].WriteString(glyph[row])
			if i < len(runes)-1 {
				rows[row].WriteString(" ")
			}
		}
	}
	var lines [3]string
	for i := 0; i < 3; i++ {
		lines[i] = theme.Gradient(rows[i].String(), theme.TitleStops...)
	}
	return strings.Join(lines[:], "\n")
}
