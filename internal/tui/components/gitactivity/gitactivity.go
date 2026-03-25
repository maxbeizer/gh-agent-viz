package gitactivity

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/diffview"
)

// Styles for rendering
var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(compat.AdaptiveColor{Light: lipgloss.Color("24"), Dark: lipgloss.Color("75")})
	statsStyle  = lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("242"), Dark: lipgloss.Color("245")})
	addStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green
	delStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
	hunkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Faint(true)
	headerStyle = lipgloss.NewStyle().Bold(true)
	sepStyle    = lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("249"), Dark: lipgloss.Color("238")})
	emptyStyle  = lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("244"), Dark: lipgloss.Color("245")}).Italic(true)
)

// Model represents the git activity view component
type Model struct {
	viewport viewport.Model
	result   *data.GitDiffResult
	files    []diffview.FileDiff
	width    int
	height   int
	ready    bool
	loading  bool
}

// New creates a new git activity model
func New(width, height int) Model {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	return Model{
		viewport: vp,
		width:    width,
		height:   height,
		loading:  true,
	}
}

// SetSize updates the component dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height)
	if m.ready {
		m.renderContent()
	}
}

// SetLoading sets the loading state
func (m *Model) SetLoading(loading bool) {
	m.loading = loading
}

// SetDiffResult updates the diff data and re-renders
func (m *Model) SetDiffResult(result *data.GitDiffResult) {
	m.result = result
	m.loading = false
	if result != nil && result.Diff != "" {
		m.files = diffview.ParseUnifiedDiff(result.Diff)
	} else {
		m.files = nil
	}
	m.ready = true
	m.renderContent()
}

// Update handles incoming messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the component
func (m Model) View() string {
	if m.loading {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			statsStyle.Render("Loading git changes…"))
	}
	return m.viewport.View()
}

// renderContent builds the viewport content from the current diff result
func (m *Model) renderContent() {
	if m.result == nil || (m.result.Diff == "" && m.result.StatLines == "") {
		m.viewport.SetContent(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			emptyStyle.Render("No uncommitted changes")))
		return
	}

	var sb strings.Builder

	// Header with stats
	sb.WriteString(titleStyle.Render("  📂 Git Activity"))
	sb.WriteString("\n")
	sb.WriteString(statsStyle.Render(fmt.Sprintf("  %d file(s) changed, %s+%d%s %s−%d%s",
		m.result.FileCount,
		addStyle.Render(""), m.result.Additions, statsStyle.Render(""),
		delStyle.Render(""), m.result.Deletions, statsStyle.Render(""))))
	sb.WriteString("\n")
	sb.WriteString(sepStyle.Render(strings.Repeat("─", m.width-2)))
	sb.WriteString("\n")

	// Stat summary (file list with +/- bars)
	if m.result.StatLines != "" {
		for _, line := range strings.Split(m.result.StatLines, "\n") {
			sb.WriteString("  " + line + "\n")
		}
		sb.WriteString(sepStyle.Render(strings.Repeat("─", m.width-2)))
		sb.WriteString("\n\n")
	}

	// Full colored diff
	for _, file := range m.files {
		sb.WriteString(headerStyle.Render(fmt.Sprintf("  %s", file.Path)))
		sb.WriteString(statsStyle.Render(fmt.Sprintf(" (+%d, -%d)", file.Additions, file.Deletions)))
		sb.WriteString("\n")

		for _, line := range strings.Split(file.Patch, "\n") {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				sb.WriteString("  " + addStyle.Render(line) + "\n")
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				sb.WriteString("  " + delStyle.Render(line) + "\n")
			} else if strings.HasPrefix(line, "@@") {
				sb.WriteString("  " + hunkStyle.Render(line) + "\n")
			} else {
				sb.WriteString("  " + line + "\n")
			}
		}
		sb.WriteString("\n")
	}

	m.viewport.SetContent(sb.String())
}

// ScrollUp scrolls up one line
func (m *Model) ScrollUp() {
	m.viewport.ScrollUp(1)
}

// ScrollDown scrolls down one line
func (m *Model) ScrollDown() {
	m.viewport.ScrollDown(1)
}

// HalfPageUp scrolls up half a page
func (m *Model) HalfPageUp() {
	m.viewport.HalfPageUp()
}

// HalfPageDown scrolls down half a page
func (m *Model) HalfPageDown() {
	m.viewport.HalfPageDown()
}
