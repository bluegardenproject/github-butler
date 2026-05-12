package theme

import "github.com/charmbracelet/lipgloss"

var (
	OuterBorder lipgloss.Style
	Panel       lipgloss.Style
	PanelTitle  lipgloss.Style
	Dimmed      lipgloss.Style
	Bold        lipgloss.Style
	OK          lipgloss.Style
	Fail        lipgloss.Style
	Pending     lipgloss.Style
	Info        lipgloss.Style
	Accent      lipgloss.Style
	SelectedRow lipgloss.Style

	DraftChip     lipgloss.Style
	StaleChip     lipgloss.Style
	CodeOwnerChip lipgloss.Style

	SuccessToast lipgloss.Style
	ErrorToast   lipgloss.Style

	KeyHint  lipgloss.Style
	KeyLabel lipgloss.Style
)

func init() {
	rebuildStyles()
}

func rebuildStyles() {
	OuterBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(NeonMagenta).
		Padding(0, 1)

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(NeonCyan).
		Padding(0, 1)

	PanelTitle = lipgloss.NewStyle().
		Foreground(NeonPink).
		Bold(true).
		MarginBottom(1)

	Dimmed = lipgloss.NewStyle().
		Foreground(Dim)

	Bold = lipgloss.NewStyle().
		Bold(true)

	OK = lipgloss.NewStyle().
		Foreground(NeonLime).
		Bold(true)

	Fail = lipgloss.NewStyle().
		Foreground(HotPink).
		Bold(true)

	Pending = lipgloss.NewStyle().
		Foreground(NeonYellow)

	Info = lipgloss.NewStyle().
		Foreground(NeonCyan)

	Accent = lipgloss.NewStyle().
		Foreground(NeonMagenta).
		Bold(true)

	SelectedRow = lipgloss.NewStyle().
		Foreground(selectedForeground).
		Background(selectedBackground).
		Bold(true)

	// Chips
	DraftChip = lipgloss.NewStyle().
		Foreground(chipForeground).
		Background(NeonCyan).
		Padding(0, 1).
		Bold(true)

	StaleChip = lipgloss.NewStyle().
		Foreground(chipForeground).
		Background(NeonOrange).
		Padding(0, 1).
		Bold(true)

	CodeOwnerChip = lipgloss.NewStyle().
		Foreground(chipForeground).
		Background(NeonYellow).
		Padding(0, 1)

	SuccessToast = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(NeonLime).
		Foreground(NeonLime).
		Padding(0, 1).
		Bold(true)

	ErrorToast = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(HotPink).
		Foreground(HotPink).
		Padding(0, 1).
		Bold(true)

	KeyHint = lipgloss.NewStyle().
		Foreground(NeonMagenta).
		Bold(true)

	KeyLabel = lipgloss.NewStyle().
		Foreground(NeonCyan)
}
