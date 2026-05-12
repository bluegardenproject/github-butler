// Package theme centralises colors, styles, and the gradient helper.
// Everything here depends only on Lipgloss so it can be reused by
// unrelated tools.
package theme

import "github.com/charmbracelet/lipgloss"

var (
	NeonPink    = lipgloss.Color("#FF10F0")
	NeonCyan    = lipgloss.Color("#00F0FF")
	NeonMagenta = lipgloss.Color("#FF00FF")
	NeonLime    = lipgloss.Color("#39FF14")
	NeonPurple  = lipgloss.Color("#BF00FF")
	NeonOrange  = lipgloss.Color("#FF6A00")
	NeonYellow  = lipgloss.Color("#F5FF00")
	NeonBlue    = lipgloss.Color("#1B03FF")
	HotPink     = lipgloss.Color("#FF2A6D")

	Black  = lipgloss.Color("#000000")
	White  = lipgloss.Color("#FFFFFF")
	Dim    = lipgloss.Color("#6C6C80")
	DarkBg = lipgloss.Color("#120018")
)

// TitleStops is the color sequence for the app banner gradient.
var TitleStops = []lipgloss.Color{NeonPink, NeonMagenta, NeonPurple, NeonCyan}

// CountdownStops colors the "next refresh" progress bar.
var CountdownStops = []lipgloss.Color{NeonPink, NeonPurple, NeonCyan}

// HeaderStops colors the table header row.
var HeaderStops = []lipgloss.Color{NeonCyan, NeonPink}

var (
	selectedForeground = NeonYellow
	selectedBackground = NeonPurple
	chipForeground     = Black
)
