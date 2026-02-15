package tui

import "github.com/charmbracelet/bubbles/key"

// Keybindings holds all key bindings for the application
type Keybindings struct {
	MoveLeft       key.Binding
	MoveRight      key.Binding
	MoveUp         key.Binding
	MoveDown       key.Binding
	SelectTask     key.Binding
	ShowLogs       key.Binding
	OpenInBrowser  key.Binding
	ResumeSession  key.Binding
	RefreshData    key.Binding
	FocusAttention key.Binding
	ExitApp        key.Binding
	ToggleFilter   key.Binding
	NavigateBack   key.Binding
}

// NewKeybindings creates the default key bindings for the TUI
func NewKeybindings() Keybindings {
	return Keybindings{
		MoveLeft: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "move column left"),
		),
		MoveRight: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "move column right"),
		),
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
		ResumeSession: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "resume session"),
		),
		RefreshData: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh tasks"),
		),
		FocusAttention: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "needs-action view"),
		),
		ExitApp: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "exit"),
		),
		ToggleFilter: key.NewBinding(
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab/shift+tab", "switch status filter"),
		),
		NavigateBack: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "return to list"),
		),
	}
}
