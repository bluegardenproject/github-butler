package ui

import (
	"github.com/bluegardenproject/github-butler/internal/ui/components"
	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	label  string
	target screen
}

var menuItems = []menuItem{
	{label: "Repositories", target: screenRepos},
	{label: "Settings", target: screenSettings},
	{label: "Back to dashboard", target: screenDashboard},
}

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, keys.Quit):
		return m, tea.Quit
	case key.Matches(km, keys.Back), key.Matches(km, keys.Menu):
		m.screen = screenDashboard
	case key.Matches(km, keys.Up):
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case key.Matches(km, keys.Down):
		if m.menuCursor < len(menuItems)-1 {
			m.menuCursor++
		}
	case key.Matches(km, keys.Select):
		target := menuItems[m.menuCursor].target
		m.screen = target
		switch target {
		case screenRepos:
			m.reposCursor = 0
		case screenSettings:
			m.settingsCursor = 0
		}
	}
	return m, nil
}

func (m Model) viewMenu() string {
	banner := components.Banner(" MENU ")

	var rows []string
	for i, item := range menuItems {
		line := "  " + item.label
		if i == m.menuCursor {
			line = theme.SelectedRow.Render(" ▸ " + item.label + " ")
		}
		rows = append(rows, line)
	}
	body := theme.Panel.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	hints := footerHints(keys.Up, keys.Down, keys.Select, keys.Back, keys.Quit)
	return lipgloss.JoinVertical(lipgloss.Left, banner, body, hints)
}
