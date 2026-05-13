package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bluegardenproject/github-butler/internal/config"
	"github.com/bluegardenproject/github-butler/internal/ui/components"
	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// settingItem describes a single editable setting. Each item owns its
// own activation behavior so the list can mix edit-screen items (like
// the poll interval input) with one-shot toggles (like group-by-repo)
// without the caller having to special-case them.
type settingItem struct {
	label    string
	value    func(cfg config.Config) string
	activate func(m Model) (Model, tea.Cmd)
}

var settingItems = []settingItem{
	{
		label:    "Poll interval (seconds)",
		value:    func(cfg config.Config) string { return strconv.Itoa(cfg.PollIntervalSeconds) },
		activate: openIntervalEditor,
	},
	{
		label:    "Group by repo",
		value:    func(cfg config.Config) string { return onOff(cfg.GroupByRepo) },
		activate: toggleGroupByRepo,
	},
	{
		label:    "Dashboard view",
		value:    func(cfg config.Config) string { return cfg.DashboardView },
		activate: openDashboardViewPicker,
	},
	{
		label:    "Theme",
		value:    func(cfg config.Config) string { return cfg.Theme.Selected },
		activate: openThemePicker,
	},
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// openIntervalEditor transitions into the numeric editor for the poll
// interval, pre-filling the current value.
func openIntervalEditor(m Model) (Model, tea.Cmd) {
	m.screen = screenEditInterval
	m.intervalInput.SetValue(strconv.Itoa(m.cfg.PollIntervalSeconds))
	m.intervalInput.Focus()
	m.validationErr = ""
	return m, nil
}

// toggleGroupByRepo flips the group-by-repo flag, re-orders the PRs
// already on screen so the effect is immediate, and persists the
// change to disk.
func toggleGroupByRepo(m Model) (Model, tea.Cmd) {
	m.cfg.GroupByRepo = !m.cfg.GroupByRepo
	m.prs = OrderPRs(m.prs, m.cfg.GroupByRepo, m.cfg.Repos)
	m.selected = 0
	return m, saveConfigCmd(m.cfg)
}

type dashboardViewChoice struct {
	id   string
	name string
}

var dashboardViewChoices = []dashboardViewChoice{
	{id: config.DashboardViewAuto, name: "Auto"},
	{id: config.DashboardViewFull, name: "Full"},
	{id: config.DashboardViewCompact, name: "Compact"},
}

func openDashboardViewPicker(m Model) (Model, tea.Cmd) {
	m.screen = screenDashboardViews
	m.viewCursor = 0
	for i, choice := range dashboardViewChoices {
		if choice.id == m.cfg.DashboardView {
			m.viewCursor = i
			break
		}
	}
	return m, nil
}

func openThemePicker(m Model) (Model, tea.Cmd) {
	m.screen = screenThemes
	m.themeCursor = 0
	for i, choice := range m.themeChoices {
		if choice.ID == m.cfg.Theme.Selected {
			m.themeCursor = i
			break
		}
	}
	return m, nil
}

func (m Model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, keys.Quit):
		return m, tea.Quit
	case key.Matches(km, keys.Back):
		m.screen = screenMenu
	case key.Matches(km, keys.Menu):
		m.screen = screenDashboard
	case key.Matches(km, keys.Up):
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case key.Matches(km, keys.Down):
		if m.settingsCursor < len(settingItems)-1 {
			m.settingsCursor++
		}
	case key.Matches(km, keys.Select):
		if m.settingsCursor >= 0 && m.settingsCursor < len(settingItems) {
			item := settingItems[m.settingsCursor]
			if item.activate != nil {
				return item.activate(m)
			}
		}
	}
	return m, nil
}

func (m Model) updateEditInterval(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(km, keys.Back):
			m.screen = screenSettings
			m.intervalInput.Blur()
			m.validationErr = ""
			return m, nil
		case km.Type == tea.KeyEnter:
			raw := strings.TrimSpace(m.intervalInput.Value())
			n, err := strconv.Atoi(raw)
			if err != nil {
				m.validationErr = "must be a whole number"
				return m, nil
			}
			if n < config.MinPollIntervalSeconds || n > config.MaxPollIntervalSeconds {
				m.validationErr = fmt.Sprintf("must be between %d and %d",
					config.MinPollIntervalSeconds, config.MaxPollIntervalSeconds)
				return m, nil
			}
			m.cfg.PollIntervalSeconds = n
			m.screen = screenSettings
			m.intervalInput.Blur()
			m.validationErr = ""
			return m, saveConfigCmd(m.cfg)
		}
	}
	var cmd tea.Cmd
	m.intervalInput, cmd = m.intervalInput.Update(msg)
	return m, cmd
}

func (m Model) updateDashboardViews(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, keys.Quit):
		return m, tea.Quit
	case key.Matches(km, keys.Back):
		m.screen = screenSettings
	case key.Matches(km, keys.Menu):
		m.screen = screenDashboard
	case key.Matches(km, keys.Up):
		if m.viewCursor > 0 {
			m.viewCursor--
		}
	case key.Matches(km, keys.Down):
		if m.viewCursor < len(dashboardViewChoices)-1 {
			m.viewCursor++
		}
	case key.Matches(km, keys.Select):
		if m.viewCursor < 0 || m.viewCursor >= len(dashboardViewChoices) {
			return m, nil
		}
		m.cfg.DashboardView = dashboardViewChoices[m.viewCursor].id
		m.screen = screenSettings
		return m, saveConfigCmd(m.cfg)
	}
	return m, nil
}

func (m Model) updateThemes(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, keys.Quit):
		return m, tea.Quit
	case key.Matches(km, keys.Back):
		m.screen = screenSettings
	case key.Matches(km, keys.Menu):
		m.screen = screenDashboard
	case key.Matches(km, keys.Up):
		if m.themeCursor > 0 {
			m.themeCursor--
		}
	case key.Matches(km, keys.Down):
		if m.themeCursor < len(m.themeChoices)-1 {
			m.themeCursor++
		}
	case key.Matches(km, keys.Select):
		if m.themeCursor < 0 || m.themeCursor >= len(m.themeChoices) {
			return m, nil
		}
		choice := m.themeChoices[m.themeCursor]
		previous := m.cfg.Theme.Selected
		m.cfg.Theme.Selected = choice.ID
		if err := theme.Activate(theme.OptionsFromConfig(m.cfg.Theme)); err != nil {
			m.cfg.Theme.Selected = previous
			_ = theme.Activate(theme.OptionsFromConfig(m.cfg.Theme))
			return m.showToast("theme failed: "+err.Error(), components.ToastError)
		}
		m.screen = screenSettings
		return m, saveConfigCmd(m.cfg)
	}
	return m, nil
}

func (m Model) viewSettings() string {
	banner := components.Banner(" SETTINGS ")

	var rows []string
	for i, item := range settingItems {
		val := theme.Accent.Render(item.value(m.cfg))
		line := fmt.Sprintf("  %s: %s", item.label, val)
		if i == m.settingsCursor && m.screen == screenSettings {
			line = theme.SelectedRow.Render(
				fmt.Sprintf(" ▸ %s: %s ", item.label, item.value(m.cfg)),
			)
		}
		rows = append(rows, line)
	}
	body := theme.Panel.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	var extra string
	if m.screen == screenEditInterval {
		extra = m.renderEditInterval()
	} else if m.screen == screenDashboardViews {
		extra = m.renderDashboardViewPicker()
	} else if m.screen == screenThemes {
		extra = m.renderThemePicker()
	}

	hints := footerHints(keys.Up, keys.Down, keys.Select, keys.Back, keys.Menu, keys.Quit)

	parts := []string{banner, body}
	if extra != "" {
		parts = append(parts, extra)
	}
	parts = append(parts, hints)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderDashboardViewPicker() string {
	label := theme.PanelTitle.Render("Select dashboard view")
	var rows []string
	for i, choice := range dashboardViewChoices {
		line := "  " + choice.name
		if i == m.viewCursor {
			line = theme.SelectedRow.Render(" ▸ " + choice.name + " ")
		}
		if choice.id == m.cfg.DashboardView {
			line += theme.Accent.Render("  current")
		}
		rows = append(rows, line)
	}
	rows = append(rows, theme.Dimmed.Render("[enter] select  [esc] cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.NeonPink).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, append([]string{label}, rows...)...))
}

func (m Model) renderThemePicker() string {
	label := theme.PanelTitle.Render("Select theme")
	var rows []string
	for i, choice := range m.themeChoices {
		line := fmt.Sprintf("  %s (%s)", choice.Name, choice.Kind)
		if i == m.themeCursor {
			line = theme.SelectedRow.Render(" ▸ " + choice.Name + " ")
		}
		if choice.ID == m.cfg.Theme.Selected {
			line += theme.Accent.Render("  current")
		}
		rows = append(rows, line)
	}
	if len(rows) == 0 {
		rows = append(rows, theme.Dimmed.Render("  no themes found"))
	}
	rows = append(rows, theme.Dimmed.Render("[enter] select  [esc] cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.NeonPink).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, append([]string{label}, rows...)...))
}

func (m Model) renderEditInterval() string {
	label := theme.PanelTitle.Render("Set poll interval")
	input := m.intervalInput.View()
	hint := theme.Dimmed.Render(fmt.Sprintf(
		"min %d, max %d seconds",
		config.MinPollIntervalSeconds, config.MaxPollIntervalSeconds,
	))

	rows := []string{label, input, hint}
	if m.validationErr != "" {
		rows = append(rows, theme.Fail.Render("Error: "+m.validationErr))
	}
	rows = append(rows, theme.Dimmed.Render("[enter] save  [esc] cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.NeonPink).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}
