package components

import (
	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Confirm renders a small y/n prompt line. Views handle the keypresses;
// this helper just renders the prompt.
func Confirm(prompt string) string {
	yes := theme.KeyHint.Render("y")
	no := theme.KeyHint.Render("n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.NeonYellow).
		Padding(0, 1).
		Render(prompt + "  " + yes + theme.KeyLabel.Render(" yes  ") + no + theme.KeyLabel.Render(" no"))
}
