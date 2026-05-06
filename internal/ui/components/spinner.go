package components

import (
	"time"

	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// spinnerFrames is a classic braille spinner — small, monospace-friendly,
// and renders well in every terminal we care about.
var spinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

// spinnerFrameInterval is how long each frame is shown. Matches the UI
// repaint cadence (uiTickInterval) so we advance roughly one frame per
// repaint and the animation looks smooth without extra timers.
const spinnerFrameInterval = 250 * time.Millisecond

// Spinner returns the current spinner glyph styled in NeonCyan. It is
// stateless: the frame is derived from the wall clock, so callers don't
// need to track an index — they just call Spinner() on every render and
// the existing UI tick will naturally advance the animation.
func Spinner() string {
	idx := int(time.Now().UnixNano()/int64(spinnerFrameInterval)) % len(spinnerFrames)
	if idx < 0 {
		idx += len(spinnerFrames)
	}
	return lipgloss.NewStyle().
		Foreground(theme.NeonCyan).
		Bold(true).
		Render(spinnerFrames[idx])
}

// Loading returns a spinner glyph followed by a styled label, e.g.
// "⠋ refreshing". Use this for inline "something is happening" hints
// next to titles or in the footer.
func Loading(label string) string {
	return Spinner() + " " + theme.Info.Render(label)
}
