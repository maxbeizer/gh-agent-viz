package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// cycleFilter cycles through status filters by delta (+1 forward, -1 backward)
func (m *Model) cycleFilter(delta int) {
	filters := []string{"active", "completed", "failed", "all", "attention"}
	for i, f := range filters {
		if f == m.ctx.StatusFilter {
			next := (i + delta) % len(filters)
			if next < 0 {
				next += len(filters)
			}
			m.ctx.StatusFilter = filters[next]
			m.showPreview = false
			break
		}
	}
}

func isValidFilter(filter string) bool {
	switch filter {
	case "all", "attention", "active", "completed", "failed":
		return true
	default:
		return false
	}
}

// smartDefaultFilter picks the best starting tab based on actual session counts.
func smartDefaultFilter(counts FilterCounts) string {
	if counts.Active > 0 {
		return "active"
	}
	if counts.Attention > 0 {
		return "attention"
	}
	return "all"
}

// previewVisible returns true when the split-pane detail preview should render.
func (m Model) previewVisible() bool {
	return m.showPreview && m.ctx.Width >= 80 && m.ctx.Height > 20
}

// updateSplitLayout recalculates component dimensions for the current layout.
func (m *Model) updateSplitLayout() {
	if m.previewVisible() {
		leftWidth := m.ctx.Width * 2 / 5
		rightWidth := m.ctx.Width - leftWidth
		contentHeight := m.ctx.Height - 4 // header + footer chrome
		m.taskList.SetSize(leftWidth, contentHeight)
		m.taskList.SetSplitMode(true)
		m.taskDetail.SetSize(rightWidth, contentHeight)
	} else {
		m.taskList.SetSize(m.ctx.Width, m.ctx.Height)
		m.taskList.SetSplitMode(false)
	}
	m.kanban.SetSize(m.ctx.Width, m.ctx.Height-4)
}

// updateFooterHints updates footer hints based on current view mode and state
func (m *Model) updateFooterHints() {
	switch m.viewMode {
	case ViewModeList:
		hints := []key.Binding{
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "navigate")),
			m.keys.SelectTask,
			m.keys.ToggleFilter,
			m.keys.ToggleKanban,
			m.keys.ToggleMission,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(hints)
	case ViewModeDetail:
		hints := []key.Binding{
			m.keys.NavigateBack,
			m.keys.ShowLogs,
		}
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.HasLog {
			hints = append(hints, key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "tools")))
		}
		// Show diff hint when session has a PR
		if canShowDiff(session) {
			hints = append(hints, m.keys.ShowDiff)
		}
		hints = append(hints, m.keys.ShowHelp, m.keys.ExitApp)
		m.footer.SetHints(hints)
	case ViewModeLog:
		logHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "scroll")),
			m.keys.ToggleFollow,
		}
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot {
			logHints = append(logHints, key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "convo")))
		}
		logHints = append(logHints, m.keys.ShowHelp, m.keys.ExitApp)
		m.footer.SetHints(logHints)
	case ViewModeKanban:
		kanbanHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("h/l"), key.WithHelp("h/l", "column")),
			key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "card")),
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(kanbanHints)
	case ViewModeToolTimeline:
		timelineHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "scroll")),
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(timelineHints)
	case ViewModeMission:
		missionHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "navigate")),
			m.keys.SelectTask,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(missionHints)
	case ViewModeDiff:
		diffHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "scroll")),
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(diffHints)
	}
}

func canShowLogs(session *data.Session) bool {
	if session == nil || strings.TrimSpace(session.ID) == "" {
		return false
	}
	if session.Source == data.SourceAgentTask {
		return true
	}
	// Local sessions can show logs if they have an events.jsonl file
	return session.HasLog
}

func canOpenPR(session *data.Session) bool {
	if session == nil {
		return false
	}
	// Has explicit PR info
	if strings.TrimSpace(session.PRURL) != "" {
		return true
	}
	if session.PRNumber > 0 && strings.TrimSpace(session.Repository) != "" {
		return true
	}
	// Can discover PR by branch lookup
	return strings.TrimSpace(session.Repository) != "" && strings.TrimSpace(session.Branch) != ""
}

func canResumeLocalSession(session *data.Session) bool {
	if session == nil || session.Source != data.SourceLocalCopilot || strings.TrimSpace(session.ID) == "" {
		return false
	}
	status := strings.ToLower(strings.TrimSpace(session.Status))
	return status == "running" || status == "queued" || status == "needs-input"
}

// canShowDiff returns true when the session has a PR or can discover one
func canShowDiff(session *data.Session) bool {
	if session == nil {
		return false
	}
	if session.PRNumber > 0 && strings.TrimSpace(session.Repository) != "" {
		return true
	}
	// Can discover PR by branch
	return strings.TrimSpace(session.Repository) != "" && strings.TrimSpace(session.Branch) != ""
}

func isSessionRunning(session *data.Session) bool {
	return session != nil && data.StatusIsActive(session.Status)
}

func (m Model) refreshCmd() tea.Cmd {
	return tea.Tick(m.refreshInt, func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func (m Model) animationTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

func (m Model) logPollTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return logPollTickMsg{}
	})
}
