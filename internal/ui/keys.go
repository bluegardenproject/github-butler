package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap groups every key binding in one place so the help footer and
// handlers stay in sync.
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Select  key.Binding
	Back    key.Binding
	Quit    key.Binding
	Refresh key.Binding
	Menu    key.Binding
	OpenPR  key.Binding
	AddRepo key.Binding
	DelRepo key.Binding
	ConfYes key.Binding
	ConfNo  key.Binding
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Select:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Menu:    key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "menu")),
	OpenPR:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open in browser")),
	AddRepo: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	DelRepo: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	ConfYes: key.NewBinding(key.WithKeys("y")),
	ConfNo:  key.NewBinding(key.WithKeys("n")),
}
