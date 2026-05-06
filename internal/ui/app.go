package ui

import (
	"time"

	"github.com/bluegardenproject/github-butler/internal/config"
	"github.com/bluegardenproject/github-butler/internal/github"
	"github.com/bluegardenproject/github-butler/internal/ui/components"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// screen identifies the currently active view.
type screen int

const (
	screenDashboard screen = iota
	screenMenu
	screenRepos
	screenAddRepo
	screenConfirmRemove
	screenSettings
	screenEditInterval
)

// Model is the single Bubble Tea model backing every screen. Per-screen
// transient state (list cursors, text inputs) lives here too so switching
// screens is a simple enum change.
type Model struct {
	cfg    config.Config
	client *github.Client

	// data
	prs         []github.PR
	lastFetched time.Time
	nextTick    time.Time
	loading     bool
	err         error

	// dashboard state
	selected int

	// screen stack (only one level deep is needed)
	screen screen

	// menu / repos / settings cursors
	menuCursor     int
	reposCursor    int
	settingsCursor int

	// inputs
	addInput      textinput.Model
	intervalInput textinput.Model

	// validation
	validating    bool
	validationErr string

	// toast
	toast   components.Toast
	toastID int64

	// layout
	width, height int
}

// NewModel constructs the root model.
func NewModel(cfg config.Config, client *github.Client) Model {
	addIn := textinput.New()
	addIn.Placeholder = "owner/repo  or  https://github.com/owner/repo"
	addIn.CharLimit = 200
	addIn.Width = 60

	intervalIn := textinput.New()
	intervalIn.CharLimit = 5
	intervalIn.Width = 10
	intervalIn.Placeholder = "seconds"

	return Model{
		cfg:           cfg,
		client:        client,
		screen:        screenDashboard,
		addInput:      addIn,
		intervalInput: intervalIn,
		// Seed the countdown so the bar/"next: Ns" hint render correctly
		// on the very first frame, before the first tickMsg arrives.
		// Init() can't do this because it has a value receiver.
		nextTick: time.Now().Add(cfg.PollInterval()),
	}
}

// Init kicks off the first fetch and the recurring ticks: a slow poll
// tick that drives data refresh, and a fast UI tick that just repaints
// the countdown bar and relative timestamps.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchPRsCmd(m.client, m.cfg.Repos),
		tickCmd(m.cfg.PollInterval()),
		uiTickCmd(),
	)
}

// Update dispatches messages to per-screen handlers, with a small set of
// always-on globals (quit, size, tick, fetched, errors, toast expiry).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tickMsg:
		if m.loading {
			return m, tickCmd(m.cfg.PollInterval())
		}
		m.loading = true
		m.nextTick = time.Now().Add(m.cfg.PollInterval())
		return m, tea.Batch(
			fetchPRsCmd(m.client, m.cfg.Repos),
			tickCmd(m.cfg.PollInterval()),
		)

	case uiTickMsg:
		return m, uiTickCmd()

	case prsFetchedMsg:
		m.loading = false
		m.prs = OrderPRs(msg.PRs, m.cfg.GroupByRepo, m.cfg.Repos)
		m.lastFetched = msg.Fetched
		m.err = nil
		if m.selected >= len(m.prs) {
			m.selected = len(m.prs) - 1
		}
		if m.selected < 0 {
			m.selected = 0
		}
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg.Err
		return m.showToast("fetch failed: "+msg.Err.Error(), components.ToastError)

	case configSavedMsg:
		m2, cmd := m.showToast("saved", components.ToastSuccess)
		return m2, tea.Batch(cmd, fetchPRsCmd(m.client, m.cfg.Repos))

	case repoValidatedMsg:
		return m.handleRepoValidated(msg)
	}

	// toast expiry
	if id, ok := components.ExpiryID(msg); ok {
		if id == m.toastID {
			m.toast = components.Toast{}
		}
		return m, nil
	}

	switch m.screen {
	case screenDashboard:
		return m.updateDashboard(msg)
	case screenMenu:
		return m.updateMenu(msg)
	case screenRepos:
		return m.updateRepos(msg)
	case screenAddRepo:
		return m.updateAddRepo(msg)
	case screenConfirmRemove:
		return m.updateConfirmRemove(msg)
	case screenSettings:
		return m.updateSettings(msg)
	case screenEditInterval:
		return m.updateEditInterval(msg)
	}
	return m, nil
}

// View renders the current screen wrapped in the outer border.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var body string
	switch m.screen {
	case screenDashboard:
		body = m.viewDashboard()
	case screenMenu:
		body = m.viewMenu()
	case screenRepos, screenAddRepo, screenConfirmRemove:
		body = m.viewRepos()
	case screenSettings, screenEditInterval:
		body = m.viewSettings()
	}

	if m.toast.Message != "" {
		body = lipgloss.JoinVertical(lipgloss.Left, body, m.toast.Render())
	}
	return body
}

func (m Model) showToast(msg string, level components.ToastLevel) (Model, tea.Cmd) {
	toast, id, cmd := components.Show(msg, level, 2500*time.Millisecond)
	m.toast = toast
	m.toastID = id
	return m, cmd
}
