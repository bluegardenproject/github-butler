package ui

import (
	"strings"

	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// pad right-pads/truncates text to a visible width, measuring with
// lipgloss so ANSI sequences and wide runes behave correctly.
func pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w == width {
		return s
	}
	if w > width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-w)
}

// padVisible pads an already-styled string (containing ANSI sequences) to
// `width` visible columns. Uses lipgloss.Width for accurate measurement.
func padVisible(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// truncate shortens s so its visible length is at most max, adding an
// ellipsis when it had to cut. Operates on runes, not bytes.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

// footerHints renders "key label  key label  ..." using the shared theme.
func footerHints(bindings ...key.Binding) string {
	parts := make([]string, 0, len(bindings))
	for _, b := range bindings {
		h := b.Help()
		if h.Key == "" {
			continue
		}
		parts = append(parts,
			theme.KeyHint.Render(h.Key)+" "+theme.KeyLabel.Render(h.Desc),
		)
	}
	return strings.Join(parts, "  ")
}
