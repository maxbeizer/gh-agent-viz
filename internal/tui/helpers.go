package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/header"
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

// mergeSessions adds new sessions to the model, deduplicating by ID,
// then recomputes counts and applies the current filter.
func (m *Model) mergeSessions(newSessions []data.Session) {
	// Build set of existing IDs
	existing := map[string]struct{}{}
	for _, s := range m.allSessions {
		existing[s.ID] = struct{}{}
	}

	// Add non-duplicate sessions
	for _, s := range newSessions {
		if _, ok := existing[s.ID]; !ok {
			m.allSessions = append(m.allSessions, s)
			existing[s.ID] = struct{}{}
		}
	}

	// Filter dismissed
	dismissedIDs := map[string]struct{}{}
	if m.dismissedStore != nil {
		dismissedIDs = m.dismissedStore.IDs()
	}
	visible := make([]data.Session, 0, len(m.allSessions))
	for _, s := range m.allSessions {
		if _, dismissed := dismissedIDs[s.ID]; !dismissed {
			visible = append(visible, s)
		}
	}

	m.recomputeAndDisplay(visible)
}

// enrichTokenUsage applies token usage data to accumulated sessions and re-displays.
func (m *Model) enrichTokenUsage(usage map[string]*data.TokenUsage) {
	for i := range m.allSessions {
		if u, ok := usage[m.allSessions[i].ID]; ok {
			if m.allSessions[i].Telemetry == nil {
				m.allSessions[i].Telemetry = &data.SessionTelemetry{}
			}
			m.allSessions[i].Telemetry.Model = u.Model
			m.allSessions[i].Telemetry.InputTokens = u.InputTokens
			m.allSessions[i].Telemetry.OutputTokens = u.OutputTokens
			m.allSessions[i].Telemetry.CachedTokens = u.CachedTokens
			m.allSessions[i].Telemetry.ModelCalls = u.Calls
		}
	}

	// Re-display with enriched data
	dismissedIDs := map[string]struct{}{}
	if m.dismissedStore != nil {
		dismissedIDs = m.dismissedStore.IDs()
	}
	visible := make([]data.Session, 0, len(m.allSessions))
	for _, s := range m.allSessions {
		if _, dismissed := dismissedIDs[s.ID]; !dismissed {
			visible = append(visible, s)
		}
	}
	m.recomputeAndDisplay(visible)
}

// recomputeAndDisplay recomputes filter counts from visible sessions,
// applies the current status filter, picks smart defaults on first load,
// and updates all display components.
func (m *Model) recomputeAndDisplay(visible []data.Session) {
	// Compute counts
	counts := FilterCounts{All: len(visible)}
	for _, session := range visible {
		if data.SessionNeedsAttention(session) {
			counts.Attention++
		}
		if data.StatusIsActive(session.Status) || strings.EqualFold(session.Status, "needs-input") {
			counts.Active++
		}
		if strings.EqualFold(session.Status, "completed") {
			counts.Completed++
		}
		if strings.EqualFold(session.Status, "failed") {
			counts.Failed++
		}
	}
	m.ctx.Counts = counts
	m.ctx.Error = nil

	m.header.SetCounts(header.FilterCounts{
		All:       counts.All,
		Attention: counts.Attention,
		Active:    counts.Active,
		Completed: counts.Completed,
		Failed:    counts.Failed,
	})

	// On first render, pick the best default tab
	if !m.initialLoadDone {
		m.ctx.StatusFilter = smartDefaultFilter(counts)
	}

	// Apply status filter
	filtered := visible
	if m.ctx.StatusFilter != "all" {
		filtered = []data.Session{}
		for _, session := range visible {
			if m.ctx.StatusFilter == "attention" && data.SessionNeedsAttention(session) {
				filtered = append(filtered, session)
			} else if m.ctx.StatusFilter == "active" && (data.StatusIsActive(session.Status) || strings.EqualFold(session.Status, "needs-input")) {
				filtered = append(filtered, session)
			} else if strings.EqualFold(session.Status, m.ctx.StatusFilter) {
				filtered = append(filtered, session)
			}
		}
	}

	// Update display components
	m.taskList.SetLoading(false)
	m.taskList.SetTasks(filtered)
	m.taskDetail.SetAllSessions(visible)
	m.kanban.SetSessions(visible)
	m.mission.SetSessions(visible)
}
