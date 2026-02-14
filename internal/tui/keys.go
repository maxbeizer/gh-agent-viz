package tui

import "github.com/charmbracelet/bubbles/key"

// Keybindings holds all key bindings for the application
type Keybindings struct {
	MoveUp         key.Binding
	MoveDown       key.Binding
	SelectTask     key.Binding
	ShowLogs       key.Binding
	OpenInBrowser  key.Binding
	RefreshData    key.Binding
	ExitApp        key.Binding
	ToggleFilter   key.Binding
	NavigateBack   key.Binding
}

// NewKeybindings creates the default key bindings for the TUI
func NewKeybindings() Keybindings {
	return Keybindings{
		MoveUp: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "navigate up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "navigate down"),
		),
		SelectTask: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show task details"),
		),
		ShowLogs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "show task logs"),
		),
		OpenInBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open PR URL"),
		),
		RefreshData: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh tasks"),
		),
		ExitApp: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "exit"),
		),
		ToggleFilter: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch status filter"),
		),
		NavigateBack: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "return to list"),
		),
	}
}
