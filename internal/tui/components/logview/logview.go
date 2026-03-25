package logview

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"charm.land/lipgloss/v2"
)

// maxLogBytes caps the raw log content retained in memory.
// Content beyond this limit is truncated from the beginning.
const maxLogBytes = 2 * 1024 * 1024 // 2 MB

// truncateLog keeps the tail of content when it exceeds maxLogBytes.
func truncateLog(content string) string {
	if len(content) <= maxLogBytes {
		return content
	}
	// Find the first newline after the cut point to avoid splitting a line
	cut := len(content) - maxLogBytes
	if idx := strings.Index(content[cut:], "\n"); idx >= 0 {
		cut += idx + 1
	}
	return "... (truncated) ...\n" + content[cut:]
}

// Model represents the log view component state
type Model struct {
	titleStyle     lipgloss.Style
	viewport       viewport.Model
	rawContent     string // Original unrendered content (may be truncated)
	rawLen         int    // Length of original content before truncation
	content        string // Rendered content currently displayed
	lineCount      int    // Cache line count for performance
	ready          bool
	followMode     bool // whether auto-scroll is active
	liveSession    bool // whether the session is running (enables LIVE indicator)
	cachedRenderer *glamour.TermRenderer // reusable renderer
	cachedWidth    int                    // width the renderer was built for
}

// New creates a new log view model
func New(titleStyle lipgloss.Style, width, height int) Model {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	return Model{
		titleStyle: titleStyle,
		viewport:   vp,
		ready:      false,
	}
}

// Init initializes the log view
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the log view
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the log view
func (m Model) View() string {
	if !m.ready {
		return "Loading logs..."
	}

	if m.rawContent == "" && !m.liveSession {
		return "No log content available. Press esc to go back."
	}

	if m.content == "" {
		return m.titleStyle.Render("No logs available")
	}

	if m.liveSession {
		indicator := " PAUSED ⏸ "
		style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")) // yellow
		if m.followMode {
			indicator = " LIVE 🔴 "
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")) // red
		}
		return style.Render(indicator) + "\n" + m.viewport.View()
	}

	return m.viewport.View()
}

// SetContent updates the log content
func (m *Model) SetContent(content string) {
	m.rawLen = len(content)
	content = truncateLog(content)
	m.rawContent = content
	rendered := m.renderWithCache(content)
	m.content = rendered
	m.lineCount = len(strings.Split(rendered, "\n"))
	m.viewport.SetContent(rendered)
	m.ready = true
}

// SetSize updates the viewport size
func (m *Model) SetSize(width, height int) {
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height)
	if m.rawContent != "" {
		rendered := m.renderWithCache(m.rawContent)
		m.content = rendered
		m.lineCount = len(strings.Split(rendered, "\n"))
		m.viewport.SetContent(rendered)
	}
}

// renderMarkdown renders content through glamour for rich markdown display.
// Falls back to raw content if rendering fails.
func renderMarkdown(content string, width int) string {
	if width <= 0 {
		width = 80
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	rendered, err := renderer.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimRight(rendered, "\n")
}

// renderWithCache renders markdown using a cached glamour renderer,
// re-creating it only when the viewport width changes.
func (m *Model) renderWithCache(content string) string {
	width := m.viewport.Width()
	if width <= 0 {
		width = 80
	}
	if m.cachedRenderer == nil || m.cachedWidth != width {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			return content
		}
		m.cachedRenderer = r
		m.cachedWidth = width
	}
	rendered, err := m.cachedRenderer.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimRight(rendered, "\n")
}

// GotoTop scrolls to the top
func (m *Model) GotoTop() {
	m.viewport.GotoTop()
}

// GotoBottom scrolls to the bottom
func (m *Model) GotoBottom() {
	m.viewport.GotoBottom()
}

// PageDown scrolls down one page
func (m *Model) PageDown() {
	m.viewport.PageDown()
}

// PageUp scrolls up one page
func (m *Model) PageUp() {
	m.viewport.PageUp()
}

// HalfPageDown scrolls down half a page
func (m *Model) HalfPageDown() {
	m.viewport.HalfPageDown()
}

// HalfPageUp scrolls up half a page
func (m *Model) HalfPageUp() {
	m.viewport.HalfPageUp()
}

// LineDown scrolls down one line
func (m *Model) LineDown() {
	if m.viewport.YOffset() < m.lineCount-m.viewport.Height() {
		m.viewport.ScrollDown(1)
	}
}

// LineUp scrolls up one line
func (m *Model) LineUp() {
	if m.viewport.YOffset() > 0 {
		m.viewport.ScrollUp(1)
	}
}

// SetFollowMode toggles follow mode
func (m *Model) SetFollowMode(on bool) {
	m.followMode = on
}

// FollowMode returns whether follow mode is active
func (m Model) FollowMode() bool {
	return m.followMode
}

// SetLive marks whether the session being viewed is running
func (m *Model) SetLive(live bool) {
	m.liveSession = live
}

// IsLive returns whether the session is live
func (m Model) IsLive() bool {
	return m.liveSession
}

// AppendOrReplace updates log content if the new content is longer than the
// current raw content (poll-and-replace strategy). The comparison uses the
// incoming content length before truncation so that growing logs are always
// accepted even after the buffer has been capped.
// When followMode is true it auto-scrolls to the bottom.
func (m *Model) AppendOrReplace(content string) {
	if len(content) <= m.rawLen {
		return
	}
	m.rawLen = len(content)
	m.SetContent(content)
	if m.followMode {
		m.viewport.GotoBottom()
	}
}
