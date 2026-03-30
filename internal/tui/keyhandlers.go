package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/mission"
)

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// When help overlay is visible, only ? and esc close it; ignore everything else
	if m.help.Visible() {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "?" || msg.Code == tea.KeyEscape {
			m.help.Toggle()
		}
		return m, nil
	}

	// Search mode: capture text input for filtering
	if m.searchActive {
		// ctrl+c always quits, even in search mode
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch msg.Code {
		case tea.KeyEscape:
			m.searchActive = false
			m.searchQuery = ""
			m.recomputeAndDisplay(m.visibleSessions())
			return m, nil
		case tea.KeyEnter:
			m.searchActive = false
			// Keep the filter active, just stop capturing input
			return m, nil
		case tea.KeyBackspace:
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.recomputeAndDisplay(m.visibleSessions())
			}
			return m, nil
		default:
			if len(msg.Text) > 0 {
				m.searchQuery += msg.Text
				m.recomputeAndDisplay(m.visibleSessions())
				return m, nil
			}
		}
		return m, nil
	}

	// ? toggles help overlay in any mode
	if msg.String() == "?" {
		m.help.Toggle()
		return m, nil
	}

	// Global quit key
	if msg.String() == "q" || msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Debug snapshot (any view)
	if msg.String() == "S" {
		ts := time.Now().UTC().Format("2006-01-02T150405Z")
		path := fmt.Sprintf("/tmp/gh-agent-viz-snapshot-%s.json", ts)
		origPath := m.snapshotPath
		m.snapshotPath = path
		m.writeSnapshot()
		m.snapshotPath = origPath
		m.toast.Push("📸", "Snapshot", path)
		return m, nil
	}

	// / activates search in navigable views
	if msg.String() == "/" {
		if m.viewMode == ViewModeList || m.viewMode == ViewModeKanban || m.viewMode == ViewModeMission || m.viewMode == ViewModeActive {
			m.searchActive = true
			m.searchQuery = ""
			return m, nil
		}
	}

	switch m.viewMode {
	case ViewModeList:
		return m.handleListKeys(msg)
	case ViewModeDetail:
		return m.handleDetailKeys(msg)
	case ViewModeLog:
		return m.handleLogKeys(msg)
	case ViewModeDiff:
		return m.handleDiffKeys(msg)
	case ViewModeKanban:
		return m.handleKanbanKeys(msg)
	case ViewModeToolTimeline:
		return m.handleToolTimelineKeys(msg)
	case ViewModeMission:
		return m.handleMissionKeys(msg)
	case ViewModeActive:
		return m.handleActiveKeys(msg)
	case ViewModeGitActivity:
		return m.handleGitActivityKeys(msg)
	}

	return m, nil
}

// handleListKeys handles keys in list view mode
func (m Model) handleListKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
	case "j", "down":
		m.taskList.MoveCursor(1)
	case "k", "up":
		m.taskList.MoveCursor(-1)
	case "enter":
		// If cursor is on a collapsed group header, expand it instead of opening detail
		if m.taskList.IsCursorOnCollapsedGroup() {
			m.taskList.ToggleGroupExpand()
			return m, nil
		}
		session := m.taskList.SelectedTask()
		if session != nil {
			if session.Source == data.SourceLocalCopilot {
				m.ctx.Error = nil
				m.viewMode = ViewModeDetail
				m.taskDetail.SetTask(session)
				return m, nil
			}
			m.viewMode = ViewModeDetail
			return m, m.fetchTaskDetail(session.ID, session.Repository)
		}
	case "l":
		session := m.taskList.SelectedTask()
		if session != nil {
			m.viewMode = ViewModeLog
			if isSessionRunning(session) {
				m.logView.SetLive(true)
				m.logView.SetFollowMode(true)
				return m, tea.Batch(m.fetchTaskLog(session.ID, session.Repository), m.logPollTick())
			}
			return m, m.fetchTaskLog(session.ID, session.Repository)
		}
	case "c":
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.HasLog {
			m.viewMode = ViewModeLog
			m.showConversation = true
			return m, m.fetchConversation(session.ID)
		} else if session != nil {
			m.toast.Push("ℹ️", "Conversation", "only available for local Copilot sessions")
		}
	case "o":
		session := m.taskList.SelectedTask()
		if session != nil {
			return m, m.openTaskPR(session)
		}
	case "s":
		session := m.taskList.SelectedTask()
		if session != nil {
			return m, m.resumeSession(session)
		}
	case "x":
		m.taskList.DismissSelected()
		m.lastFingerprint = ""
		m.lastSplitTaskID = ""
		m.recomputeAndDisplay(m.visibleSessions())
	case "X":
		count := m.taskList.DismissCompleted()
		if count > 0 {
			m.lastFingerprint = ""
			m.recomputeAndDisplay(m.visibleSessions())
			m.toast.Push("🧹", "Dismissed", fmt.Sprintf("%d completed session(s) cleared", count))
		} else {
			m.toast.Push("ℹ️", "Nothing to dismiss", "no completed sessions found")
		}
	case "r":
		return m, m.fetchTasks
	case "p":
		m.showPreview = !m.showPreview
		m.updateSplitLayout()
		return m, nil
	case "a":
		m.ctx.StatusFilter = "attention"
		m.showPreview = false
		m.taskList.SetLoading(true)
		return m, m.fetchTasks
	case "tab":
		m.cycleFilter(1)
		m.taskList.SetLoading(true)
		return m, m.fetchTasks
	case "shift+tab", "backtab":
		m.cycleFilter(-1)
		m.taskList.SetLoading(true)
		return m, m.fetchTasks
	case "g":
		m.taskList.CycleGroupBy()
		return m, nil
	case "space":
		if m.taskList.IsGrouped() {
			m.taskList.ToggleGroupExpand()
			return m, nil
		}
	case "K":
		m.viewMode = ViewModeKanban
		m.kanban.SetSessions(m.visibleSessions())
		m.kanban.SetSize(m.ctx.Width, m.ctx.Height-6)
		return m, nil
	case "M":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
		return m, nil
	case "A":
		m.viewMode = ViewModeActive
		m.activeView.SetSessions(m.visibleSessions())
		m.activeView.SetSize(m.ctx.Width, m.ctx.Height-6)
		return m, nil
	case "!":
		return m, m.openSessionRepo()
	case "@":
		return m, m.openFileIssue()
	case "d":
		session := m.taskList.SelectedTask()
		if session != nil && canShowDiff(session) {
			m.diffView.SetLoading()
			m.viewMode = ViewModeDiff
			return m, m.fetchPRDiff(session)
		} else if session != nil {
			m.toast.Push("⚠️", session.Title, "no PR — session is on "+session.Branch)
		}
	case "t":
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.HasLog {
			m.viewMode = ViewModeToolTimeline
			m.toolTimeline.SetSize(m.ctx.Width-4, m.ctx.Height-8)
			return m, m.fetchToolTimeline(session.ID)
		} else if session != nil {
			m.toast.Push("ℹ️", "Tool Timeline", "only available for local Copilot sessions")
		}
	case "G":
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.WorkDir != "" {
			m.gitActivity.SetLoading(true)
			m.gitActivity.SetSize(m.ctx.Width-4, m.ctx.Height-8)
			m.viewMode = ViewModeGitActivity
			return m, tea.Batch(m.fetchGitDiff(session.WorkDir), m.gitDiffPollTick())
		} else if session != nil {
			m.toast.Push("ℹ️", "Git Activity", "only available for local sessions with a working directory")
		}
	}
	return m, nil
}

// handleDetailKeys handles keys in detail view mode
func (m Model) handleDetailKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
	case "l":
		session := m.taskList.SelectedTask()
		if session != nil {
			m.viewMode = ViewModeLog
			if isSessionRunning(session) {
				m.logView.SetLive(true)
				m.logView.SetFollowMode(true)
				return m, tea.Batch(m.fetchTaskLog(session.ID, session.Repository), m.logPollTick())
			}
			return m, m.fetchTaskLog(session.ID, session.Repository)
		}
	case "c":
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.HasLog {
			m.viewMode = ViewModeLog
			m.showConversation = true
			return m, m.fetchConversation(session.ID)
		} else if session != nil {
			m.toast.Push("ℹ️", "Conversation", "only available for local Copilot sessions")
		}
	case "o":
		session := m.taskList.SelectedTask()
		if session != nil {
			return m, m.openTaskPR(session)
		}
	case "s":
		session := m.taskList.SelectedTask()
		if session != nil {
			return m, m.resumeSession(session)
		}
	case "t":
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.HasLog {
			m.viewMode = ViewModeToolTimeline
			m.toolTimeline.SetSize(m.ctx.Width-4, m.ctx.Height-8)
			return m, m.fetchToolTimeline(session.ID)
		} else if session != nil {
			m.toast.Push("ℹ️", "Tool Timeline", "only available for local Copilot sessions")
		}
	case "d":
		session := m.taskList.SelectedTask()
		if session != nil && canShowDiff(session) {
			m.diffView.SetLoading()
			m.viewMode = ViewModeDiff
			return m, m.fetchPRDiff(session)
		}
	case "G":
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot && session.WorkDir != "" {
			m.gitActivity.SetLoading(true)
			m.gitActivity.SetSize(m.ctx.Width-4, m.ctx.Height-8)
			m.viewMode = ViewModeGitActivity
			return m, tea.Batch(m.fetchGitDiff(session.WorkDir), m.gitDiffPollTick())
		} else if session != nil {
			m.toast.Push("ℹ️", "Git Activity", "only available for local sessions with a working directory")
		}
	case "x":
		session := m.taskDetail.Session()
		if session != nil && session.ID != "" && m.dismissedStore != nil {
			m.dismissedStore.Add(session.ID)
			m.taskList.DismissByID(session.ID)
		}
		m.lastFingerprint = "" // force recompute
		m.lastSplitTaskID = ""
		m.recomputeAndDisplay(m.visibleSessions())
		m.viewMode = ViewModeMission
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
	}
	return m, nil
}

// handleDiffKeys handles keys in diff view mode
func (m Model) handleDiffKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
		return m, nil
	}

	// Delegate to viewport for scrolling
	var cmd tea.Cmd
	m.diffView, cmd = m.diffView.Update(msg)
	return m, cmd
}

// handleGitActivityKeys handles keys in git activity view mode
func (m Model) handleGitActivityKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
		return m, nil
	case "r":
		// Manual refresh
		session := m.taskList.SelectedTask()
		if session != nil && session.WorkDir != "" {
			m.gitActivity.SetLoading(true)
			return m, m.fetchGitDiff(session.WorkDir)
		}
		return m, nil
	}

	// Delegate to viewport for scrolling
	var cmd tea.Cmd
	m.gitActivity, cmd = m.gitActivity.Update(msg)
	return m, cmd
}

func (m Model) handleLogKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
		m.logView.SetLive(false)
		m.logView.SetFollowMode(false)
		m.showConversation = false
	case "c":
		session := m.taskList.SelectedTask()
		if session != nil && session.Source == data.SourceLocalCopilot {
			if m.showConversation {
				m.showConversation = false
			} else {
				return m, m.fetchConversation(session.ID)
			}
		}
	case "j", "down":
		if m.showConversation {
			m.conversationView.LineDown()
		} else {
			m.logView.SetFollowMode(false)
			m.logView.LineDown()
		}
	case "k", "up":
		if m.showConversation {
			m.conversationView.LineUp()
		} else {
			m.logView.SetFollowMode(false)
			m.logView.LineUp()
		}
	case "d":
		if m.showConversation {
			m.conversationView.HalfPageDown()
		} else {
			m.logView.SetFollowMode(false)
			m.logView.HalfPageDown()
		}
	case "u":
		if m.showConversation {
			m.conversationView.HalfPageUp()
		} else {
			m.logView.SetFollowMode(false)
			m.logView.HalfPageUp()
		}
	case "g":
		if m.showConversation {
			m.conversationView.GotoTop()
		} else {
			m.logView.SetFollowMode(false)
			m.logView.GotoTop()
		}
	case "G":
		if m.showConversation {
			m.conversationView.GotoBottom()
		} else {
			m.logView.SetFollowMode(true)
			m.logView.GotoBottom()
		}
	case "f":
		if !m.showConversation {
			m.logView.SetFollowMode(!m.logView.FollowMode())
			if m.logView.FollowMode() {
				m.logView.GotoBottom()
			}
		}
	case "s":
		session := m.taskList.SelectedTask()
		if session != nil {
			return m, m.resumeSession(session)
		}
	}
	return m, nil
}

// handleKanbanKeys handles keys in kanban view mode
func (m Model) handleKanbanKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "K":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
	case "h", "left":
		m.kanban.MoveColumn(-1)
	case "l", "right":
		m.kanban.MoveColumn(1)
	case "tab":
		m.kanban.MoveColumn(1)
	case "shift+tab", "backtab":
		m.kanban.MoveColumn(-1)
	case "j", "down":
		m.kanban.MoveRow(1)
	case "k", "up":
		m.kanban.MoveRow(-1)
	case "enter":
		session := m.kanban.SelectedSession()
		if session != nil {
			if session.Source == data.SourceLocalCopilot {
				m.ctx.Error = nil
				m.viewMode = ViewModeDetail
				m.taskDetail.SetTask(session)
				return m, nil
			}
			m.viewMode = ViewModeDetail
			return m, m.fetchTaskDetail(session.ID, session.Repository)
		}
	case "X":
		// Dismiss completed/failed sessions from all sessions
		count := 0
		if m.dismissedStore != nil {
			m.toast.Push("🧹", "Sweeping", "clearing the decks...")
			for _, s := range m.allSessions {
				status := strings.ToLower(strings.TrimSpace(s.Status))
				if status == "completed" || status == "failed" {
					m.dismissedStore.Add(s.ID)
					count++
				}
			}
		}
		if count > 0 {
			m.toast.Push("✨", "Spotless", fmt.Sprintf("%d session(s) swept away", count))
			m.kanban.SetSessions(m.visibleSessions())
		} else {
			m.toast.Push("🤷", "Already clean", "nothing to dismiss")
		}
	case "r":
		return m, m.fetchTasks
	case "A":
		m.viewMode = ViewModeActive
		m.activeView.SetSessions(m.visibleSessions())
		m.activeView.SetSize(m.ctx.Width, m.ctx.Height-6)
	}
	return m, nil
}

// handleToolTimelineKeys handles keys in tool timeline view mode
func (m Model) handleToolTimelineKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeDetail
	case "j", "down":
		m.toolTimeline.LineDown()
	case "k", "up":
		m.toolTimeline.LineUp()
	case "d":
		m.toolTimeline.HalfPageDown()
	case "u":
		m.toolTimeline.HalfPageUp()
	}
	return m, nil
}

// handleMissionKeys handles keys in mission control view mode
func (m Model) handleMissionKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.mission.MoveCursor(1)
	case "k", "up":
		m.mission.MoveCursor(-1)
	case "tab":
		m.mission.CyclePanel(1)
	case "shift+tab", "backtab":
		m.mission.CyclePanel(-1)
	case "enter":
		// Drill into selected item based on focused panel
		switch m.mission.Focus() {
		case mission.PanelActive, mission.PanelAttention, mission.PanelRecent, mission.PanelIdle:
			session := m.mission.SelectedSession()
			if session != nil {
				if session.Source == data.SourceLocalCopilot {
					m.ctx.Error = nil
					m.viewMode = ViewModeDetail
					m.taskDetail.SetTask(session)
					return m, nil
				}
				m.viewMode = ViewModeDetail
				return m, m.fetchTaskDetail(session.ID, session.Repository)
			}
		case mission.PanelRepos:
			// Filter list view to show only this repo's sessions
			repo := m.mission.SelectedRepo()
			if repo != "" {
				m.viewMode = ViewModeList
				filtered := []data.Session{}
				for _, s := range m.visibleSessions() {
					r := s.Repository
					if r == "" { r = "local" }
					if r == repo {
						filtered = append(filtered, s)
					}
				}
				m.taskList.SetTasks(filtered)
			}
		}
	case "K":
		m.viewMode = ViewModeKanban
		m.kanban.SetSessions(m.visibleSessions())
		m.kanban.SetSize(m.ctx.Width, m.ctx.Height-6)
	case "A":
		m.viewMode = ViewModeActive
		m.activeView.SetSessions(m.visibleSessions())
		m.activeView.SetSize(m.ctx.Width, m.ctx.Height-6)
	case "r":
		return m, m.fetchTasks
	}
	return m, nil
}

// handleActiveKeys handles keys in active sessions view mode
func (m Model) handleActiveKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "A":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
	case "j", "down":
		m.activeView.MoveCursor(1)
	case "k", "up":
		m.activeView.MoveCursor(-1)
	case "enter":
		session := m.activeView.SelectedSession()
		if session != nil {
			if session.Source == data.SourceLocalCopilot {
				m.ctx.Error = nil
				m.viewMode = ViewModeDetail
				m.taskDetail.SetTask(session)
				return m, nil
			}
			m.viewMode = ViewModeDetail
			return m, m.fetchTaskDetail(session.ID, session.Repository)
		}
	case "o":
		session := m.activeView.SelectedSession()
		if session != nil {
			return m, m.openTaskPR(session)
		}
	case "l":
		session := m.activeView.SelectedSession()
		if session != nil {
			m.viewMode = ViewModeLog
			if isSessionRunning(session) {
				m.logView.SetLive(true)
				m.logView.SetFollowMode(true)
				return m, tea.Batch(m.fetchTaskLog(session.ID, session.Repository), m.logPollTick())
			}
			return m, m.fetchTaskLog(session.ID, session.Repository)
		}
	case "c":
		session := m.activeView.SelectedSession()
		if session != nil {
			return m, m.copyToClipboard(session.ID)
		}
	case "x":
		m.activeView.DismissSelected()
		m.lastFingerprint = ""
		m.recomputeAndDisplay(m.visibleSessions())
	case "r":
		return m, m.fetchTasks
	case "K":
		m.viewMode = ViewModeKanban
		m.kanban.SetSessions(m.visibleSessions())
		m.kanban.SetSize(m.ctx.Width, m.ctx.Height-6)
	case "M":
		m.viewMode = ViewModeMission
		m.mission.SetSessions(m.visibleSessions())
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-6)
	}
	return m, nil
}

// handleMouse processes mouse events for scrolling and navigation.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			switch m.viewMode {
			case ViewModeList:
				m.taskList.MoveCursor(-3)
			case ViewModeKanban:
				m.kanban.MoveRow(-1)
			case ViewModeMission:
				m.mission.MoveCursor(-1)
			case ViewModeActive:
				m.activeView.MoveCursor(-1)
			case ViewModeLog:
				if m.showConversation {
					m.conversationView.LineUp()
					m.conversationView.LineUp()
					m.conversationView.LineUp()
				} else {
					m.logView.SetFollowMode(false)
					m.logView.LineUp()
					m.logView.LineUp()
					m.logView.LineUp()
				}
			}
		} else if msg.Button == tea.MouseWheelDown {
			switch m.viewMode {
			case ViewModeList:
				m.taskList.MoveCursor(3)
			case ViewModeKanban:
				m.kanban.MoveRow(1)
			case ViewModeMission:
				m.mission.MoveCursor(1)
			case ViewModeActive:
				m.activeView.MoveCursor(1)
			case ViewModeLog:
				if m.showConversation {
					m.conversationView.LineDown()
					m.conversationView.LineDown()
					m.conversationView.LineDown()
				} else {
					m.logView.SetFollowMode(false)
					m.logView.LineDown()
					m.logView.LineDown()
					m.logView.LineDown()
				}
			}
		}
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft && m.viewMode == ViewModeMission {
			mouse := msg.Mouse()
			midX := m.ctx.Width / 2
			if mouse.X < midX {
				leftMid := m.ctx.Height / 2
				if mouse.Y < leftMid {
					m.mission.SetFocus(mission.PanelActive)
				} else {
					m.mission.SetFocus(mission.PanelAttention)
				}
			} else {
				m.mission.SetFocus(mission.PanelRepos)
			}
		}
	}
	return m, nil
}
