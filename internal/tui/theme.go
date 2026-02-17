package tui

import "github.com/charmbracelet/lipgloss"

// Theme contains all Lip Gloss styles for the UI
type Theme struct {
	name             string
	StatusRunning    lipgloss.Style
	StatusQueued     lipgloss.Style
	StatusCompleted  lipgloss.Style
	StatusFailed     lipgloss.Style
	TableHeader      lipgloss.Style
	TableRow         lipgloss.Style
	TableRowSelected lipgloss.Style
	Border           lipgloss.Style
	Title            lipgloss.Style
	Footer           lipgloss.Style
	// Tab bar styles
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabCount    lipgloss.Style
	// Focus area styles
	FocusBorder   lipgloss.Style
	RowGutter     lipgloss.Style
	RowGutterSel  lipgloss.Style
	SectionHeader lipgloss.Style
}

// ThemeName returns the name of the active theme.
func (t *Theme) ThemeName() string {
	return t.name
}

// NewTheme creates the default theme using ANSI color numbers.
func NewTheme() *Theme {
	return newDefaultTheme()
}

// NewThemeFromConfig returns the theme matching themeName, or the adaptive
// default when the name is empty or unrecognised.
func NewThemeFromConfig(themeName string) *Theme {
	switch themeName {
	case "catppuccin-mocha":
		return newCatppuccinMochaTheme()
	case "dracula":
		return newDraculaTheme()
	case "tokyo-night":
		return newTokyoNightTheme()
	case "default":
		return newDefaultTheme()
	default:
		return newAdaptiveDefaultTheme()
	}
}

// newDefaultTheme builds the original hardcoded ANSI theme.
func newDefaultTheme() *Theme {
	return &Theme{
		name:            "default",
		StatusRunning:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		StatusQueued:    lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
		StatusCompleted: lipgloss.NewStyle().Foreground(lipgloss.Color("46")),
		StatusFailed:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true),
		TableRow: lipgloss.NewStyle().
			Padding(0, 1),
		TableRowSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("15")).
			Bold(true),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99")).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 2),
		TabCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		FocusBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")),
		RowGutter: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
		RowGutterSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true),
		SectionHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Bold(true).
			Padding(0, 1),
	}
}

// newAdaptiveDefaultTheme uses lipgloss.AdaptiveColor so the palette
// automatically adjusts to light and dark terminals.
// Color philosophy: teal/cyan for structure, warm amber for actions,
// distinct status colors, and clear hierarchy via weight not just color.
func newAdaptiveDefaultTheme() *Theme {
	return &Theme{
		name:            "default",
		StatusRunning:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "42"}),
		StatusQueued:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "178", Dark: "222"}),
		StatusCompleted: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "30", Dark: "72"}),
		StatusFailed:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "203"}),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"}).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "249", Dark: "238"}),
		TableRow: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"}),
		TableRowSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "255"}).
			Bold(true),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "249", Dark: "240"}),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"}).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "245"}).
			Padding(0, 1),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "230"}).
			Background(lipgloss.AdaptiveColor{Light: "24", Dark: "24"}).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "242", Dark: "248"}).
			Padding(0, 2),
		TabCount: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "245"}),
		FocusBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "30", Dark: "73"}),
		RowGutter: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "239"}),
		RowGutterSel: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "30", Dark: "73"}).
			Bold(true),
		SectionHeader: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "30", Dark: "73"}).
			Bold(true).
			Padding(0, 1),
	}
}

// newCatppuccinMochaTheme returns a theme using the Catppuccin Mocha palette.
func newCatppuccinMochaTheme() *Theme {
	return &Theme{
		name:            "catppuccin-mocha",
		StatusRunning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
		StatusQueued:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")),
		StatusCompleted: lipgloss.NewStyle().Foreground(lipgloss.Color("#94e2d5")),
		StatusFailed:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#cba6f7")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true),
		TableRow: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#cdd6f4")),
		TableRowSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#cdd6f4")).
			Bold(true),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#313244")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#cba6f7")).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8")).
			Padding(0, 1),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1e1e2e")).
			Background(lipgloss.Color("#cba6f7")).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8")).
			Padding(0, 2),
		TabCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8")),
		FocusBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")),
		RowGutter: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#313244")),
		RowGutterSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")).
			Bold(true),
		SectionHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true).
			Padding(0, 1),
	}
}

// newDraculaTheme returns a theme using the Dracula palette.
func newDraculaTheme() *Theme {
	return &Theme{
		name:            "dracula",
		StatusRunning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")),
		StatusQueued:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f1fa8c")),
		StatusCompleted: lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")),
		StatusFailed:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bd93f9")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true),
		TableRow: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#f8f8f2")),
		TableRowSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#f8f8f2")).
			Bold(true),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#44475a")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bd93f9")).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Padding(0, 1),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282a36")).
			Background(lipgloss.Color("#bd93f9")).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Padding(0, 2),
		TabCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")),
		FocusBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#8be9fd")),
		RowGutter: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#44475a")),
		RowGutterSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bd93f9")).
			Bold(true),
		SectionHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8be9fd")).
			Bold(true).
			Padding(0, 1),
	}
}

// newTokyoNightTheme returns a theme using the Tokyo Night palette.
func newTokyoNightTheme() *Theme {
	return &Theme{
		name:            "tokyo-night",
		StatusRunning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")),
		StatusQueued:    lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")),
		StatusCompleted: lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff")),
		StatusFailed:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bb9af7")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true),
		TableRow: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#c0caf5")),
		TableRowSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#c0caf5")).
			Bold(true),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#33467c")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bb9af7")).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89")).
			Padding(0, 1),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(lipgloss.Color("#bb9af7")).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89")).
			Padding(0, 2),
		TabCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89")),
		FocusBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7aa2f7")),
		RowGutter: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#33467c")),
		RowGutterSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bb9af7")).
			Bold(true),
		SectionHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7aa2f7")).
			Bold(true).
			Padding(0, 1),
	}
}

// StatusIcon returns the appropriate emoji icon for a given status
func StatusIcon(status string) string {
	switch status {
	case "running":
		return "●"
	case "queued":
		return "○"
	case "needs-input":
		return "🧑"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "⚪"
	}
}

// Running sessions: steady dot, gentle color breathing between two close greens
var runningColors = []string{"42", "42", "42", "36", "36", "42"}

// AnimatedStatusIcon returns a subtly animated icon for running sessions.
// Queued and other statuses use their static icon — only "in progress"
// sessions get the gentle color pulse.
func AnimatedStatusIcon(status string, frame int) string {
	if status == "running" {
		color := runningColors[frame%len(runningColors)]
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("●")
	}
	return StatusIcon(status)
}
