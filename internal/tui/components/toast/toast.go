package toast

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	defaultTTL       = 5 * time.Second
	defaultMaxToasts = 3
	defaultWidth     = 30
	maxTitleLen      = 20
)

// Toast represents a single notification.
type Toast struct {
	Icon    string
	Title   string
	Message string
	Created time.Time
}

// Model manages a stack of active toasts.
type Model struct {
	toasts    []Toast
	ttl       time.Duration
	maxToasts int
	width     int
}

// New creates a new toast model with default settings.
func New() Model {
	return Model{
		ttl:       defaultTTL,
		maxToasts: defaultMaxToasts,
		width:     defaultWidth,
	}
}

// Push adds a new toast to the stack. If the stack is full, the oldest
// toast is evicted to make room.
func (m *Model) Push(icon, title, message string) {
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-1] + "…"
	}
	t := Toast{
		Icon:    icon,
		Title:   title,
		Message: message,
		Created: time.Now(),
	}
	m.toasts = append(m.toasts, t)
	if len(m.toasts) > m.maxToasts {
		m.toasts = m.toasts[len(m.toasts)-m.maxToasts:]
	}
}

// Tick removes expired toasts. Call this on every animation tick or refresh.
func (m *Model) Tick() {
	now := time.Now()
	alive := m.toasts[:0]
	for _, t := range m.toasts {
		if now.Sub(t.Created) < m.ttl {
			alive = append(alive, t)
		}
	}
	m.toasts = alive
}

// View renders the toast stack as minimal single-line notifications.
// Returns empty string when no active toasts.
func (m Model) View() string {
	if len(m.toasts) == 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	titleStyle := lipgloss.NewStyle().Bold(true)
	msgStyle := lipgloss.NewStyle().Faint(true)

	var lines []string
	for _, t := range m.toasts {
		line := fmt.Sprintf("%s %s %s",
			t.Icon,
			titleStyle.Render(t.Title),
			msgStyle.Render(t.Message))
		lines = append(lines, style.Render(line))
	}

	return strings.Join(lines, "\n")
}

// HasToasts returns true if there are active toasts to display.
func (m Model) HasToasts() bool {
	return len(m.toasts) > 0
}

// SetWidth updates the rendering width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// Count returns the number of active toasts (useful for testing).
func (m Model) Count() int {
	return len(m.toasts)
}
