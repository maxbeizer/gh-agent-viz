package conversation

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// MessageRole identifies who sent a chat message.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
)

// ChatMessage is a single message in the conversation timeline.
type ChatMessage struct {
	Role      MessageRole
	Content   string
	Timestamp string
	Tools     []string // tool names used during this turn
}

// Model is the Bubble Tea model for the conversation view.
type Model struct {
	messages []ChatMessage
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// New creates a new conversation view model.
func New(width, height int) Model {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	return Model{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// SetMessages replaces the displayed messages and re-renders.
func (m *Model) SetMessages(messages []ChatMessage) {
	m.messages = messages
	m.renderContent()
}

// SetSize updates the component dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height)
	if m.ready {
		m.renderContent()
	}
}

// Update handles Bubble Tea messages (viewport passthrough).
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the conversation.
func (m Model) View() string {
	if !m.ready || len(m.messages) == 0 {
		return lipgloss.NewStyle().Faint(true).Render("No conversation events found for this session.")
	}
	return m.viewport.View()
}

// Scrolling helpers

func (m *Model) LineUp()       { m.viewport.ScrollUp(1) }
func (m *Model) LineDown()     { m.viewport.ScrollDown(1) }
func (m *Model) HalfPageUp()   { m.viewport.HalfPageUp() }
func (m *Model) HalfPageDown() { m.viewport.HalfPageDown() }
func (m *Model) GotoTop()      { m.viewport.GotoTop() }
func (m *Model) GotoBottom()   { m.viewport.GotoBottom() }

// ---- rendering ----

// renderContent builds the full conversation view and pushes it into the viewport.
func (m *Model) renderContent() {
	if len(m.messages) == 0 {
		m.ready = true
		return
	}

	bubbleWidth := m.bubbleWidth()
	var sections []string
	var prevTime time.Time

	for _, msg := range m.messages {
		// Insert timestamp separator when gap > 5 minutes
		if ts, ok := parseTimestamp(msg.Timestamp); ok {
			if !prevTime.IsZero() && ts.Sub(prevTime) > 5*time.Minute {
				sep := renderTimeSeparator(ts, m.width)
				sections = append(sections, sep)
			}
			prevTime = ts
		}

		switch msg.Role {
		case RoleUser:
			sections = append(sections, renderUserBubble(msg, bubbleWidth))
		case RoleAssistant:
			sections = append(sections, renderAgentBubble(msg, bubbleWidth, m.width))
		case RoleSystem:
			sections = append(sections, renderSystemBubble(msg, m.width))
		}
	}

	m.viewport.SetContent(strings.Join(sections, "\n\n"))
	m.ready = true
}

func (m *Model) bubbleWidth() int {
	w := m.width * 65 / 100
	if w < 40 {
		w = m.width - 4
	}
	if w < 10 {
		w = 10
	}
	return w
}

// ---- bubble renderers ----

var (
	userBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false).
			BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("27"), Dark: lipgloss.Color("69")}).
			PaddingLeft(1)

	agentBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.ThickBorder()).
				BorderLeft(false).
				BorderRight(true).
				BorderTop(false).
				BorderBottom(false).
				BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("42")}).
				PaddingRight(1)

	headerStyle = lipgloss.NewStyle().Bold(true)

	timestampStyle = lipgloss.NewStyle().Faint(true)

	toolStyle = lipgloss.NewStyle().Faint(true)

	separatorStyle = lipgloss.NewStyle().Faint(true).Align(lipgloss.Center)
)

func renderUserBubble(msg ChatMessage, bubbleWidth int) string {
	ts := formatShortTimestamp(msg.Timestamp)

	hdr := headerStyle.Render("You")
	if ts != "" {
		pad := bubbleWidth - lipgloss.Width(hdr) - lipgloss.Width(ts) - 2
		if pad < 1 {
			pad = 1
		}
		hdr = hdr + strings.Repeat(" ", pad) + timestampStyle.Render(ts)
	}

	body := wordWrap(msg.Content, bubbleWidth-2)
	inner := hdr + "\n" + body

	return userBorderStyle.Width(bubbleWidth).Render(inner)
}

func renderAgentBubble(msg ChatMessage, bubbleWidth, totalWidth int) string {
	ts := formatShortTimestamp(msg.Timestamp)

	hdr := headerStyle.Render("Agent")
	if ts != "" {
		pad := bubbleWidth - lipgloss.Width(hdr) - lipgloss.Width(ts) - 2
		if pad < 1 {
			pad = 1
		}
		hdr = hdr + strings.Repeat(" ", pad) + timestampStyle.Render(ts)
	}

	body := wordWrap(msg.Content, bubbleWidth-2)

	// Append tool line if any
	if len(msg.Tools) > 0 {
		body += "\n" + toolStyle.Render(formatToolLine(msg.Tools))
	}

	inner := hdr + "\n" + body
	bubble := agentBorderStyle.Width(bubbleWidth).Render(inner)

	// Right-align: indent from left
	indent := totalWidth - lipgloss.Width(bubble)
	if indent < 0 {
		indent = 0
	}
	return lipgloss.NewStyle().PaddingLeft(indent).Render(bubble)
}

func renderSystemBubble(msg ChatMessage, totalWidth int) string {
	text := "⚡ " + msg.Content
	return separatorStyle.Width(totalWidth).Render(text)
}

func renderTimeSeparator(ts time.Time, totalWidth int) string {
	text := fmt.Sprintf("── %s ──", ts.Format("15:04"))
	return separatorStyle.Width(totalWidth).Render(text)
}

// ---- tool formatting ----

var toolIcons = map[string]string{
	"bash":   "🔧",
	"edit":   "✏️",
	"grep":   "🔍",
	"view":   "👁️",
	"create": "📄",
	"glob":   "🔍",
}

func formatToolLine(tools []string) string {
	seen := make(map[string]bool)
	var parts []string
	for _, t := range tools {
		if seen[t] {
			continue
		}
		seen[t] = true
		icon, ok := toolIcons[t]
		if !ok {
			icon = "🔧"
		}
		parts = append(parts, icon+" "+t)
	}
	return strings.Join(parts, " • ")
}

// ---- helpers ----

func parseTimestamp(ts string) (time.Time, bool) {
	layouts := []string{time.RFC3339Nano, time.RFC3339}
	for _, l := range layouts {
		if t, err := time.Parse(l, ts); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func formatShortTimestamp(ts string) string {
	if t, ok := parseTimestamp(ts); ok {
		return t.Format("15:04")
	}
	return ""
}

// wordWrap performs a simple word wrap at maxWidth.
func wordWrap(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if lipgloss.Width(line) <= maxWidth {
			if result.Len() > 0 {
				result.WriteByte('\n')
			}
			result.WriteString(line)
			continue
		}
		words := strings.Fields(line)
		cur := ""
		for _, w := range words {
			if cur == "" {
				cur = w
			} else if lipgloss.Width(cur+" "+w) <= maxWidth {
				cur += " " + w
			} else {
				if result.Len() > 0 {
					result.WriteByte('\n')
				}
				result.WriteString(cur)
				cur = w
			}
		}
		if cur != "" {
			if result.Len() > 0 {
				result.WriteByte('\n')
			}
			result.WriteString(cur)
		}
	}
	return result.String()
}
