package ui

import (
	"fmt"

	"github.com/bluegardenproject/github-butler/internal/github"
	"github.com/bluegardenproject/github-butler/internal/ui/components"
	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateRepos(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.reposCursor > 0 {
			m.reposCursor--
		}
	case key.Matches(km, keys.Down):
		if m.reposCursor < len(m.cfg.Repos)-1 {
			m.reposCursor++
		}
	case key.Matches(km, keys.AddRepo):
		m.screen = screenAddRepo
		m.addInput.SetValue("")
		m.addInput.Focus()
		m.validationErr = ""
		m.validating = false
		return m, nil
	case key.Matches(km, keys.DelRepo):
		if len(m.cfg.Repos) > 0 {
			m.screen = screenConfirmRemove
		}
	}
	return m, nil
}

func (m Model) updateAddRepo(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(km, keys.Back):
			m.screen = screenRepos
			m.addInput.Blur()
			m.validating = false
			m.validationErr = ""
			return m, nil
		case km.Type == tea.KeyEnter:
			raw := m.addInput.Value()
			slug, err := github.ParseRepoRef(raw)
			if err != nil {
				m.validationErr = err.Error()
				return m, nil
			}
			for _, existing := range m.cfg.Repos {
				if existing == slug {
					m.validationErr = "already tracked"
					return m, nil
				}
			}
			m.validating = true
			m.validationErr = ""
			return m, validateRepoCmd(m.client, slug)
		}
	}
	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m Model) handleRepoValidated(msg repoValidatedMsg) (tea.Model, tea.Cmd) {
	m.validating = false
	if msg.Err != nil {
		m.validationErr = msg.Err.Error()
		return m, nil
	}
	m.cfg.Repos = append(m.cfg.Repos, msg.Slug)
	m.screen = screenRepos
	m.addInput.Blur()
	return m, saveConfigCmd(m.cfg)
}

func (m Model) updateConfirmRemove(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, keys.ConfYes):
		if m.reposCursor >= 0 && m.reposCursor < len(m.cfg.Repos) {
			m.cfg.Repos = append(m.cfg.Repos[:m.reposCursor], m.cfg.Repos[m.reposCursor+1:]...)
			if m.reposCursor >= len(m.cfg.Repos) {
				m.reposCursor = len(m.cfg.Repos) - 1
				if m.reposCursor < 0 {
					m.reposCursor = 0
				}
			}
			m.screen = screenRepos
			return m, saveConfigCmd(m.cfg)
		}
		m.screen = screenRepos
	case key.Matches(km, keys.ConfNo), key.Matches(km, keys.Back):
		m.screen = screenRepos
	case key.Matches(km, keys.Menu):
		m.screen = screenDashboard
	}
	return m, nil
}

func (m Model) viewRepos() string {
	banner := components.Banner(" REPOSITORIES ")

	list := m.renderRepoList()
	body := theme.Panel.Render(list)

	var extra string
	switch m.screen {
	case screenAddRepo:
		extra = m.renderAddRepo()
	case screenConfirmRemove:
		if m.reposCursor >= 0 && m.reposCursor < len(m.cfg.Repos) {
			extra = components.Confirm(fmt.Sprintf("Remove %q?", m.cfg.Repos[m.reposCursor]))
		}
	}

	hints := footerHints(keys.AddRepo, keys.DelRepo, keys.Up, keys.Down, keys.Back, keys.Menu, keys.Quit)

	parts := []string{banner, body}
	if extra != "" {
		parts = append(parts, extra)
	}
	parts = append(parts, hints)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderRepoList() string {
	if len(m.cfg.Repos) == 0 {
		return theme.Dimmed.Render("  (none — press ") +
			theme.KeyHint.Render("a") +
			theme.Dimmed.Render(" to add one)")
	}
	var rows []string
	for i, r := range m.cfg.Repos {
		line := "  " + r
		if i == m.reposCursor && m.screen == screenRepos {
			line = theme.SelectedRow.Render(" ▸ " + r + " ")
		} else if i == m.reposCursor {
			line = theme.Accent.Render(" ▸ " + r)
		}
		rows = append(rows, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderAddRepo() string {
	label := theme.PanelTitle.Render("Add repository")
	input := m.addInput.View()
	hint := theme.Dimmed.Render("Accepts: owner/repo · https://github.com/owner/repo · git@github.com:owner/repo.git")

	var status string
	switch {
	case m.validating:
		status = theme.Pending.Render("Validating with gh api…")
	case m.validationErr != "":
		status = theme.Fail.Render("Error: " + m.validationErr)
	}

	rows := []string{label, input, hint}
	if status != "" {
		rows = append(rows, status)
	}
	rows = append(rows, theme.Dimmed.Render("[enter] submit  [esc] cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.NeonPink).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}
