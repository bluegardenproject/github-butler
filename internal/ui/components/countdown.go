package components

import (
	"strings"

	"github.com/bluegardenproject/github-butler/internal/ui/theme"
)

// Countdown renders a width-wide horizontal gradient progress bar showing
// how much time remains until the next refresh. remaining/total is clamped
// to [0,1]. Filled runes fade through CountdownStops; empty runes are dim.
func Countdown(remaining, total float64, width int) string {
	if width <= 0 {
		return ""
	}
	if total <= 0 {
		total = 1
	}
	ratio := remaining / total
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}

	fillText := strings.Repeat("█", filled)
	emptyText := strings.Repeat("░", width-filled)

	return theme.Gradient(fillText, theme.CountdownStops...) +
		theme.Dimmed.Render(emptyText)
}
