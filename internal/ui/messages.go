package ui

import (
	"time"

	"github.com/bluegardenproject/github-butler/internal/github"
)

// Shared Bubble Tea messages. Grouped here so any view or command can
// reference them without worrying about import cycles.

type tickMsg time.Time

// uiTickMsg is a fast, display-only heartbeat used to animate the
// countdown bar and relative timestamps. It never triggers a fetch.
type uiTickMsg time.Time

type prsFetchedMsg struct {
	PRs     []github.PR
	Fetched time.Time
}

type errMsg struct{ Err error }

func (e errMsg) Error() string { return e.Err.Error() }

type configSavedMsg struct{}

type repoValidatedMsg struct {
	Slug string
	Err  error
}
