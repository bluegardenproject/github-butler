package theme

import "github.com/charmbracelet/lipgloss"

// Pre-built styles. Views should compose these rather than re-creating
// them inline.
var (
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
			Foreground(NeonYellow).
			Background(NeonPurple).
			Bold(true)

	// Chips
	DraftChip = lipgloss.NewStyle().
			Foreground(Black).
			Background(NeonCyan).
			Padding(0, 1).
			Bold(true)

	StaleChip = lipgloss.NewStyle().
			Foreground(Black).
			Background(NeonOrange).
			Padding(0, 1).
			Bold(true)

	CodeOwnerChip = lipgloss.NewStyle().
			Foreground(Black).
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
)
