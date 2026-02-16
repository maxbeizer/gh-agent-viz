package tui

import "github.com/charmbracelet/bubbles/key"

// Keybindings holds all key bindings for the application
type Keybindings struct {
	MoveUp         key.Binding
	MoveDown       key.Binding
	SelectTask     key.Binding
	ShowLogs       key.Binding
	OpenInBrowser  key.Binding
	ResumeSession  key.Binding
	DismissSession key.Binding
	RefreshData    key.Binding
	FocusAttention key.Binding
	ExitApp        key.Binding
	ToggleFilter   key.Binding
	NavigateBack   key.Binding
	TogglePreview  key.Binding
	GroupBy        key.Binding
}

// NewKeybindings creates the default key bindings for the TUI
func NewKeybindings() Keybindings {
	return Keybindings{
		MoveUp: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		SelectTask: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "details"),
		),
		ShowLogs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "logs"),
		),
		OpenInBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open PR"),
		),
		ResumeSession: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "resume"),
		),
		DismissSession: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "dismiss"),
		),
		RefreshData: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		FocusAttention: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "action view"),
		),
		ExitApp: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "exit"),
		),
		ToggleFilter: key.NewBinding(
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab", "filter"),
		),
		NavigateBack: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		TogglePreview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "preview"),
		),
		GroupBy: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "group"),
		),
	}
}
