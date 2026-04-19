package tui

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/footer"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/header"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/statsbar"
)

// maxSessions is the upper bound on accumulated sessions kept in memory.
// When exceeded, the oldest sessions (by UpdatedAt) are discarded.
const maxSessions = 500

// sessionFingerprint returns a hash summarising session IDs, statuses, and
// update timestamps so callers can cheaply detect whether a refresh actually
// changed anything.
func sessionFingerprint(sessions []data.Session) string {
	h := sha256.New()
	for _, s := range sessions {
		fmt.Fprintf(h, "%s|%s|%d\n", s.ID, s.Status, s.UpdatedAt.Unix())
	}
	return hex.EncodeToString(h.Sum(nil))
}

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
	if counts.Attention > 0 || counts.Warning > 0 {
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
}

// updateFooterHints updates footer hints based on current view mode and state
func (m *Model) updateFooterHints() {
	switch m.viewMode {
	case ViewModeList:
		m.footer.SetBadge(" 📋 List ", footer.BadgeBgList())
		m.footer.ClearStatus()
		hints := []key.Binding{
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "navigate")),
			m.keys.SelectTask,
			m.keys.ToggleFilter,
			m.keys.SearchFilter,
			m.keys.ToggleMission,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(hints)
	case ViewModeDetail:
		m.footer.SetBadge(" 🔍 Detail ", footer.BadgeBgDetail())
		m.footer.ClearStatus()
		hints := []key.Binding{
			m.keys.NavigateBack,
			m.keys.ShowLogs,
		}
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.HasLog {
			hints = append(hints, key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "tools")))
		}
		if canShowDiff(session) {
			hints = append(hints, m.keys.ShowDiff)
		}
		if session != nil && session.Source == data.SourceLocalCopilot && session.WorkDir != "" {
			hints = append(hints, m.keys.ShowGitActivity)
		}
		hints = append(hints, m.keys.DismissSession, m.keys.ShowHelp, m.keys.ExitApp)
		m.footer.SetHints(hints)
	case ViewModeLog:
		m.footer.SetBadge(" 📜 Logs ", footer.BadgeBgLog())
		m.footer.ClearStatus()
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
	case ViewModeToolTimeline:
		m.footer.SetBadge(" 🔧 Timeline ", footer.BadgeBgDetail())
		m.footer.ClearStatus()
		timelineHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "scroll")),
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(timelineHints)
	case ViewModeMission:
		m.footer.SetBadge(" 🚀 Mission ", footer.BadgeBgMission())
		m.footer.ClearStatus()
		missionHints := []key.Binding{
			key.NewBinding(key.WithKeys("1-5"), key.WithHelp("1-5", "panel")),
			key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "navigate")),
			m.keys.SelectTask,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(missionHints)
	case ViewModeDiff:
		m.footer.SetBadge(" 📝 Diff ", footer.BadgeBgDetail())
		m.footer.ClearStatus()
		diffHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "scroll")),
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(diffHints)
	case ViewModeGitActivity:
		m.footer.SetBadge(" 🌿 Git ", footer.BadgeBgDetail())
		m.footer.ClearStatus()
		gitHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "scroll")),
			m.keys.RefreshData,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(gitHints)
	case ViewModeActive:
		m.footer.SetBadge(" ⚡ Active ", footer.BadgeBgActive())
		// Set status from selected session
		s := m.activeView.SelectedSession()
		if s != nil {
			st := strings.ToLower(strings.TrimSpace(s.Status))
			statusBg := footer.StatusBgRunning()
			if st == "failed" {
				statusBg = footer.StatusBgFailed()
			} else if st == "needs-input" {
				statusBg = footer.StatusBgNeedsInput()
			}
			m.footer.SetStatus(fmt.Sprintf(" %s ", s.Status), statusBg)
		} else {
			m.footer.ClearStatus()
		}
		activeHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "navigate")),
			m.keys.SelectTask,
			m.keys.OpenInBrowser,
			m.keys.ShowLogs,
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy ID")),
			m.keys.DismissSession,
			m.keys.RefreshData,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(activeHints)
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

// visibleSessions returns allSessions minus dismissed ones. Used to push
// fresh data to components when the user switches views.
func (m Model) visibleSessions() []data.Session {
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
	return visible
}

func (m Model) refreshCmd() tea.Cmd {
	return tea.Tick(m.refreshInt, func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func (m Model) needsAnimation() bool {
	return m.ctx.Counts.Active > 0 || m.toast.HasToasts()
}

func (m Model) animationTickCmd() tea.Cmd {
	// Stop the tick loop entirely when nothing needs animating.
	// The loop is restarted by refreshTickMsg when conditions change.
	if !m.needsAnimation() {
		return nil
	}
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

func (m Model) logPollTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return logPollTickMsg{}
	})
}

func (m Model) gitDiffPollTick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return gitDiffPollTickMsg{}
	})
}

// mergeSessions merges new sessions into the model by updating existing
// entries in-place and appending truly new ones, then recomputes counts
// and applies the current filter.
func (m *Model) mergeSessions(newSessions []data.Session) {
	// Invalidate the split-view cache so updated data is reflected
	m.lastSplitTaskID = ""

	// Build index of existing sessions by ID
	existingIdx := map[string]int{}
	for i, s := range m.allSessions {
		existingIdx[s.ID] = i
	}

	// Update existing, append new
	for _, s := range newSessions {
		if idx, ok := existingIdx[s.ID]; ok {
			m.allSessions[idx] = s // update in place
		} else {
			m.allSessions = append(m.allSessions, s)
			existingIdx[s.ID] = len(m.allSessions) - 1
		}
	}

	// Cap session count to prevent unbounded memory growth
	if len(m.allSessions) > maxSessions {
		sort.SliceStable(m.allSessions, func(i, j int) bool {
			return m.allSessions[i].UpdatedAt.After(m.allSessions[j].UpdatedAt)
		})
		m.allSessions = m.allSessions[:maxSessions]
	}

	// Filter dismissed — but auto-undismiss sessions whose status changed
	// to urgent since they were last seen (e.g. was "running", now "failed").
	// Sessions explicitly dismissed while already urgent stay dismissed.
	dismissedIDs := map[string]struct{}{}
	if m.dismissedStore != nil {
		dismissedIDs = m.dismissedStore.IDs()
	}
	visible := make([]data.Session, 0, len(m.allSessions))
	for _, s := range m.allSessions {
		if _, dismissed := dismissedIDs[s.ID]; dismissed {
			if data.SessionAttentionLevel(s) >= data.AttentionWarning {
				prevStatus, seen := m.prevSessions[s.ID]
				statusChanged := seen && !strings.EqualFold(prevStatus, s.Status)
				if statusChanged {
					if m.dismissedStore != nil {
						m.dismissedStore.Remove(s.ID)
					}
					visible = append(visible, s)
					continue
				}
			}
			continue
		}
		visible = append(visible, s)
	}

	m.recomputeAndDisplay(visible)
}

// enrichTokenUsage applies token usage data to accumulated sessions and re-displays.
func (m *Model) enrichTokenUsage(usage map[string]*data.TokenUsage) {
	m.tokenUsageMap = usage
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
// and updates all display components. Skips component updates when the
// session data has not changed (fingerprint match).
func (m *Model) recomputeAndDisplay(visible []data.Session) {
	// Fast-path: skip when data and active filter haven't changed
	fp := sessionFingerprint(visible) + "|" + m.ctx.StatusFilter + "|" + m.searchQuery
	unchanged := m.lastFingerprint == fp && m.initialLoadDone
	if !unchanged {
		m.lastFingerprint = fp
	}

	// Compute counts
	counts := FilterCounts{All: len(visible)}
	for _, session := range visible {
		level := data.SessionAttentionLevel(session)
		if level >= data.AttentionUrgent {
			counts.Attention++
		}
		if level == data.AttentionWarning {
			counts.Warning++
		}
		if data.StatusIsActive(session.Status) || strings.EqualFold(session.Status, "needs-input") {
			if data.SessionIsActiveNotIdle(session) {
				counts.Active++
			} else {
				counts.Idle++
			}
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
		Warning:   counts.Warning,
		Active:    counts.Active,
		Completed: counts.Completed,
		Failed:    counts.Failed,
	})

	// Update stats bar with aggregate metrics
	totalTokens := int64(0)
	for _, session := range visible {
		if session.Telemetry != nil {
			totalTokens += session.Telemetry.InputTokens
		}
	}
	m.statsBar.SetCounts(statsbar.Counts{
		Active:      counts.Active,
		Idle:        counts.Idle,
		Attention:   counts.Attention,
		Warning:     counts.Warning,
		Completed:   counts.Completed,
		TotalTokens: totalTokens,
		TotalCost:   data.TotalCost(m.tokenUsageMap),
	})

	// On first render, pick the best default tab
	if !m.initialLoadDone {
		m.ctx.StatusFilter = smartDefaultFilter(counts)
	}

	// If data is unchanged, counts/header are still fresh — skip the
	// expensive per-component SetTasks / SetSessions calls.
	if unchanged {
		return
	}

	// Apply status filter
	filtered := visible
	if m.ctx.StatusFilter != "all" {
		filtered = []data.Session{}
		for _, session := range visible {
			if m.ctx.StatusFilter == "attention" && data.SessionNeedsAnyAttention(session) {
				filtered = append(filtered, session)
			} else if m.ctx.StatusFilter == "active" && (data.StatusIsActive(session.Status) || strings.EqualFold(session.Status, "needs-input")) {
				filtered = append(filtered, session)
			} else if strings.EqualFold(session.Status, m.ctx.StatusFilter) {
				filtered = append(filtered, session)
			}
		}
	}

	// Apply search filter
	if m.searchQuery != "" {
		q := strings.ToLower(m.searchQuery)
		searchFiltered := make([]data.Session, 0, len(filtered))
		for _, session := range filtered {
			if strings.Contains(strings.ToLower(session.Title), q) ||
				strings.Contains(strings.ToLower(session.Repository), q) ||
				strings.Contains(strings.ToLower(session.Branch), q) ||
				strings.Contains(strings.ToLower(session.Status), q) {
				searchFiltered = append(searchFiltered, session)
			}
		}
		filtered = searchFiltered
	}

	// Update display components — only push data to the active view to avoid
	// wasted work on invisible components. They'll receive fresh data when
	// the user switches to them.
	m.taskList.SetLoading(false)
	m.taskList.SetTasks(filtered)
	switch m.viewMode {
	case ViewModeMission:
		m.mission.SetSessions(visible)
	case ViewModeActive:
		m.activeView.SetSessions(visible)
	default:
		m.taskDetail.SetAllSessions(visible)
	}
}
