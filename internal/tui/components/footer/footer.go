package footer

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// Powerline separator characters
const (
	sepRight = "\ue0b0" // 
	sepLeft  = "\ue0b2" // 
)

// Palette colors (Catppuccin Mocha)
var (
	colorBase     = lipgloss.Color("#1e1e2e")
	colorSurface0 = lipgloss.Color("#313244")
	colorSurface1 = lipgloss.Color("#45475a")
	colorSurface2 = lipgloss.Color("#585b70")
	colorText     = lipgloss.Color("#cdd6f4")
	colorLavender = lipgloss.Color("#b4befe")
	colorMauve    = lipgloss.Color("#cba6f7")
	colorGreen    = lipgloss.Color("#a6e3a1")
	colorYellow   = lipgloss.Color("#f9e2af")
	colorRed      = lipgloss.Color("#f38ba8")
	colorTeal     = lipgloss.Color("#94e2d5")
)

// segment holds a powerline segment's content and colors.
type segment struct {
	text string
	fg   color.Color
	bg   color.Color
}

// Model represents the powerline-style footer component.
type Model struct {
	width    int
	hints    []key.Binding
	badge    string   // left-side badge text (e.g. "⚡ Active", "📋 List")
	badgeBg  color.Color
	status   string   // optional status text (e.g. "running", "failed")
	statusBg color.Color
}

// New creates a new powerline footer model.
func New(_ lipgloss.Style, keys []key.Binding) Model {
	return Model{
		hints:   keys,
		badge:   " ⚡ Agent Viz ",
		badgeBg: colorMauve,
	}
}

// SetHints updates the key binding hints.
func (m *Model) SetHints(keys []key.Binding) {
	m.hints = keys
}

// SetBadge sets the left-side badge text and color.
func (m *Model) SetBadge(text string, bg color.Color) {
	m.badge = text
	m.badgeBg = bg
}

// SetStatus sets the optional status segment.
func (m *Model) SetStatus(text string, bg color.Color) {
	m.status = text
	m.statusBg = bg
}

// ClearStatus removes the status segment.
func (m *Model) ClearStatus() {
	m.status = ""
}

// SetWidth updates the available terminal width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// View renders the powerline-style footer.
func (m Model) View() string {
	// Left side: badge + optional status
	leftSegs := []segment{
		{text: m.badge, fg: colorBase, bg: m.badgeBg},
	}
	if m.status != "" {
		leftSegs = append(leftSegs, segment{
			text: m.status, fg: colorBase, bg: m.statusBg,
		})
	}

	left := renderPowerlineLeft(leftSegs)
	leftW := lipgloss.Width(left)

	// Right side: key hints (truncated to fit available space)
	rightSegs := m.buildHintSegments(leftW)
	right := renderPowerlineRight(rightSegs)

	rightW := lipgloss.Width(right)
	gap := m.width - leftW - rightW
	if gap < 0 {
		gap = 0
	}

	mid := lipgloss.NewStyle().
		Background(colorBase).
		Width(gap).
		Render("")

	return "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, mid, right)
}

func (m Model) buildHintSegments(leftWidth int) []segment {
	if len(m.hints) == 0 {
		return nil
	}

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(colorLavender)

	segs := make([]segment, 0, len(m.hints))
	for i, h := range m.hints {
		help := h.Help()
		if help.Key == "" {
			continue
		}
		text := fmt.Sprintf(" %s %s ", keyStyle.Render(help.Key), help.Desc)

		// Alternate between surface1 and surface2 for visual grouping
		bg := colorSurface1
		if i >= len(m.hints)-2 {
			bg = colorSurface2
		}
		segs = append(segs, segment{text: text, fg: colorText, bg: bg})
	}

	// Last hint gets the accent color
	if len(segs) > 0 {
		segs[len(segs)-1].bg = colorMauve
	}

	// Truncate hints that don't fit — reserve space for the left badge + a small gap
	totalW := 0
	for _, s := range segs {
		totalW += lipgloss.Width(s.text) + 1 // +1 for separator
	}
	maxRight := m.width - leftWidth - 4
	if maxRight < 20 {
		maxRight = 20
	}
	for totalW > maxRight && len(segs) > 2 {
		dropped := segs[len(segs)-2]
		totalW -= lipgloss.Width(dropped.text) + 1
		segs = append(segs[:len(segs)-2], segs[len(segs)-1])
	}

	return segs
}

func renderPowerlineLeft(segs []segment) string {
	if len(segs) == 0 {
		return ""
	}
	var b strings.Builder
	for i, seg := range segs {
		body := lipgloss.NewStyle().
			Foreground(seg.fg).
			Background(seg.bg).
			Render(seg.text)
		b.WriteString(body)

		nextBg := colorBase
		if i+1 < len(segs) {
			nextBg = segs[i+1].bg
		}
		arrow := lipgloss.NewStyle().
			Foreground(seg.bg).
			Background(nextBg).
			Render(sepRight)
		b.WriteString(arrow)
	}
	return b.String()
}

func renderPowerlineRight(segs []segment) string {
	if len(segs) == 0 {
		return ""
	}
	var b strings.Builder
	for i, seg := range segs {
		prevBg := colorBase
		if i > 0 {
			prevBg = segs[i-1].bg
		}
		arrow := lipgloss.NewStyle().
			Foreground(seg.bg).
			Background(prevBg).
			Render(sepLeft)
		b.WriteString(arrow)

		body := lipgloss.NewStyle().
			Foreground(seg.fg).
			Background(seg.bg).
			Render(seg.text)
		b.WriteString(body)
	}
	return b.String()
}

// Badge color helpers for view modes
func BadgeBgMission() color.Color  { return colorMauve }
func BadgeBgActive() color.Color   { return colorGreen }
func BadgeBgList() color.Color     { return colorLavender }
func BadgeBgDetail() color.Color   { return colorSurface2 }
func BadgeBgLog() color.Color      { return colorSurface2 }

func StatusBgRunning() color.Color    { return colorTeal }
func StatusBgFailed() color.Color     { return colorRed }
func StatusBgNeedsInput() color.Color { return colorYellow }
