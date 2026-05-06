package components

import (
	"time"

	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
)

// ToastLevel controls toast styling.
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastError
)

// Toast is an ephemeral message pinned to a corner of the screen.
type Toast struct {
	Message string
	Level   ToastLevel
	Expires time.Time
}

// toastExpiredMsg is emitted when a toast should disappear.
type toastExpiredMsg struct{ id int64 }

// Show returns a new Toast plus a tea.Cmd that will emit an expiry message
// after the given duration. The returned int64 id lets callers correlate
// the expiry message with this specific toast (so a later toast doesn't
// get dismissed by an older timer).
func Show(message string, level ToastLevel, ttl time.Duration) (Toast, int64, tea.Cmd) {
	id := time.Now().UnixNano()
	t := Toast{
		Message: message,
		Level:   level,
		Expires: time.Now().Add(ttl),
	}
	return t, id, tea.Tick(ttl, func(time.Time) tea.Msg {
		return toastExpiredMsg{id: id}
	})
}

// ExpiryID returns the id from a toastExpiredMsg if the message matches.
// Views call this from their Update to know whether to dismiss the toast.
func ExpiryID(msg tea.Msg) (int64, bool) {
	if m, ok := msg.(toastExpiredMsg); ok {
		return m.id, true
	}
	return 0, false
}

// Render returns the styled toast string.
func (t Toast) Render() string {
	if t.Message == "" {
		return ""
	}
	switch t.Level {
	case ToastSuccess:
		return theme.SuccessToast.Render("✓ " + t.Message)
	case ToastError:
		return theme.ErrorToast.Render("✗ " + t.Message)
	default:
		return theme.Info.Render(t.Message)
	}
}
