package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When help overlay is visible, only ? and esc close it; ignore everything else
	if m.help.Visible() {
		if msg.String() == "?" || msg.Type == tea.KeyEscape {
			m.help.Toggle()
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
	}

	return m, nil
}

// handleListKeys handles keys in list view mode
func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
	case " ":
		if m.taskList.IsGrouped() {
			m.taskList.ToggleGroupExpand()
			return m, nil
		}
	case "K":
		m.viewMode = ViewModeKanban
		m.kanban.SetSize(m.ctx.Width, m.ctx.Height-4)
		return m, nil
	case "M":
		m.viewMode = ViewModeMission
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-4)
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
		}
	}
	return m, nil
}

// handleDetailKeys handles keys in detail view mode
func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeList
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
		}
	case "d":
		session := m.taskList.SelectedTask()
		if session != nil && canShowDiff(session) {
			m.diffView.SetLoading()
			m.viewMode = ViewModeDiff
			return m, m.fetchPRDiff(session)
		}
	}
	return m, nil
}

// handleDiffKeys handles keys in diff view mode
func (m Model) handleDiffKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeList
		return m, nil
	}

	// Delegate to viewport for scrolling
	var cmd tea.Cmd
	m.diffView, cmd = m.diffView.Update(msg)
	return m, cmd
}

func (m Model) handleLogKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeList
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
func (m Model) handleKanbanKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "K":
		m.viewMode = ViewModeList
	case "h", "left":
		m.kanban.MoveColumn(-1)
	case "l", "right":
		m.kanban.MoveColumn(1)
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
	case "r":
		return m, m.fetchTasks
	}
	return m, nil
}

// handleToolTimelineKeys handles keys in tool timeline view mode
func (m Model) handleToolTimelineKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
func (m Model) handleMissionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "M":
		m.viewMode = ViewModeList
	case "j", "down":
		m.mission.MoveCursor(1)
	case "k", "up":
		m.mission.MoveCursor(-1)
	case "enter":
		// Return to list view (enter feels natural to "drill in")
		m.viewMode = ViewModeList
	case "r":
		return m, m.fetchTasks
	}
	return m, nil
}
