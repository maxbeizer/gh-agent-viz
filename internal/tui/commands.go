package tui

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/conversation"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/diffview"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/tooltimeline"
)

// Message types
type tasksLoadedMsg struct {
	tasks       []data.Session
	allSessions []data.Session // unfiltered, for kanban/mission
	counts      FilterCounts
}

type taskDetailLoadedMsg struct {
	task *data.Session
}

type taskLogLoadedMsg struct {
	log string
}

type toolTimelineLoadedMsg struct {
	events []tooltimeline.ToolEvent
}

type diffLoadedMsg struct {
	files []diffview.FileDiff
}

type refreshTickMsg struct{}

type animationTickMsg struct{}

type logPollTickMsg struct{}

type logPollResultMsg struct {
	log string
	err error
}

type errMsg struct {
	err error
}

type conversationLoadedMsg struct {
	messages []conversation.ChatMessage
}

// fetchTasks fetches the list of sessions (both agent tasks and local sessions)
func (m Model) fetchTasks() tea.Msg {
	sessions, err := data.FetchAllSessions(m.repo)
	if err != nil {
		return errMsg{err}
	}

	// Exclude dismissed sessions before computing anything
	dismissedIDs := map[string]struct{}{}
	if m.dismissedStore != nil {
		dismissedIDs = m.dismissedStore.IDs()
	}
	visible := make([]data.Session, 0, len(sessions))
	for _, session := range sessions {
		if _, dismissed := dismissedIDs[session.ID]; !dismissed {
			visible = append(visible, session)
		}
	}
	sessions = visible

	// Enrich sessions with token usage from CLI logs
	tokenUsage, _ := data.FetchTokenUsage()
	for i := range sessions {
		if usage, ok := tokenUsage[sessions[i].ID]; ok {
			if sessions[i].Telemetry == nil {
				sessions[i].Telemetry = &data.SessionTelemetry{}
			}
			sessions[i].Telemetry.Model = usage.Model
			sessions[i].Telemetry.InputTokens = usage.InputTokens
			sessions[i].Telemetry.OutputTokens = usage.OutputTokens
			sessions[i].Telemetry.CachedTokens = usage.CachedTokens
			sessions[i].Telemetry.ModelCalls = usage.Calls
		}
	}

	// Compute counts across all visible (non-dismissed) sessions
	counts := FilterCounts{All: len(sessions)}
	for _, session := range sessions {
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

	// Filter sessions based on status filter
	if m.ctx.StatusFilter != "all" {
		filtered := []data.Session{}
		for _, session := range sessions {
			if m.ctx.StatusFilter == "attention" && data.SessionNeedsAttention(session) {
				filtered = append(filtered, session)
			} else if m.ctx.StatusFilter == "active" && (data.StatusIsActive(session.Status) || strings.EqualFold(session.Status, "needs-input")) {
				filtered = append(filtered, session)
			} else if strings.EqualFold(session.Status, m.ctx.StatusFilter) {
				filtered = append(filtered, session)
			}
		}
		sessions = filtered
	}

	return tasksLoadedMsg{sessions, visible, counts}
}

// fetchTaskDetail fetches detailed information for a session
func (m Model) fetchTaskDetail(id string, repo string) tea.Cmd {
	return func() tea.Msg {
		// For now, we only support detail view for agent-task sessions
		// Local sessions don't have a detail API yet
		task, err := data.FetchAgentTaskDetail(id, repo)
		if err != nil {
			return errMsg{err}
		}
		session := data.FromAgentTask(*task)
		return taskDetailLoadedMsg{&session}
	}
}

// fetchTaskLog fetches the log for a task
func (m Model) fetchTaskLog(id string, repo string) tea.Cmd {
	session := m.taskList.SelectedTask()
	return func() tea.Msg {
		// Route to local session log reader for local-copilot sessions
		if session != nil && session.Source == data.SourceLocalCopilot {
			log, err := data.FetchLocalSessionLog(id)
			if err != nil {
				return errMsg{err}
			}
			return taskLogLoadedMsg{log}
		}

		log, err := data.FetchAgentTaskLog(id, repo)
		if err != nil {
			return errMsg{err}
		}
		return taskLogLoadedMsg{log}
	}
}

// fetchToolTimeline fetches tool execution events for the timeline view
func (m Model) fetchToolTimeline(sessionID string) tea.Cmd {
	return func() tea.Msg {
		events, err := data.FetchSessionEvents(sessionID)
		if err != nil {
			return errMsg{err}
		}

		var toolEvents []tooltimeline.ToolEvent
		for _, ev := range events {
			if ev.Type == "tool.execution_start" && ev.ToolName != "" {
				toolEvents = append(toolEvents, tooltimeline.ToolEvent{
					Timestamp: ev.Timestamp,
					ToolName:  ev.ToolName,
					Icon:      tooltimeline.ToolIcon(ev.ToolName),
				})
			}
		}

		return toolTimelineLoadedMsg{events: toolEvents}
	}
}

// fetchConversation loads events for a local session and converts them to chat messages.
func (m Model) fetchConversation(sessionID string) tea.Cmd {
	return func() tea.Msg {
		events, err := data.FetchSessionEvents(sessionID)
		if err != nil {
			return errMsg{err}
		}

		var messages []conversation.ChatMessage
		var pendingTools []string

		for _, ev := range events {
		switch ev.Type {
			case "session.start":
				messages = append(messages, conversation.ChatMessage{
					Role:      conversation.RoleSystem,
					Content:   "Session started",
					Timestamp: ev.Timestamp,
				})
			case "user.message":
				if ev.Content != "" {
					messages = append(messages, conversation.ChatMessage{
						Role:      conversation.RoleUser,
						Content:   ev.Content,
						Timestamp: ev.Timestamp,
					})
				}
			case "tool.execution_start":
				if ev.ToolName != "" {
					pendingTools = append(pendingTools, ev.ToolName)
				}
			case "assistant.message":
				if ev.Content != "" {
					messages = append(messages, conversation.ChatMessage{
						Role:      conversation.RoleAssistant,
						Content:   ev.Content,
						Timestamp: ev.Timestamp,
						Tools:     pendingTools,
					})
					pendingTools = nil
				}
			case "abort":
				messages = append(messages, conversation.ChatMessage{
					Role:      conversation.RoleSystem,
					Content:   "Session aborted",
					Timestamp: ev.Timestamp,
				})
			}
		}

		return conversationLoadedMsg{messages: messages}
	}
}

func (m Model) openSourceRepo() tea.Cmd {
	return func() tea.Msg {
		_, err := data.RunGH("repo", "view", "maxbeizer/gh-agent-viz", "--web")
		if err != nil {
			return errMsg{fmt.Errorf("failed to open repo: %w", err)}
		}
		return nil
	}
}

func (m Model) openFileIssue() tea.Cmd {
	return func() tea.Msg {
		_, err := data.RunGH("issue", "create", "-R", "maxbeizer/gh-agent-viz", "--web")
		if err != nil {
			return errMsg{fmt.Errorf("failed to open issue form: %w", err)}
		}
		return nil
	}
}

func (m Model) openTaskPR(session *data.Session) tea.Cmd {
	return func() tea.Msg {
		if session == nil {
			return errMsg{fmt.Errorf("no session selected")}
		}

		// If we already have PR info, use it
		if session.PRURL != "" {
			output, err := data.RunGH("pr", "view", session.PRURL, "--web")
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
			return nil
		}
		if session.PRNumber > 0 && session.Repository != "" {
			output, err := data.RunGH("pr", "view", fmt.Sprintf("%d", session.PRNumber), "-R", session.Repository, "--web")
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
			return nil
		}

		// Try to discover PR by branch name
		if session.Repository != "" && session.Branch != "" {
			prNumber, prURL, _ := data.FetchPRForBranch(session.Repository, session.Branch)
			if prURL != "" {
				session.PRNumber = prNumber
				session.PRURL = prURL
				output, err := data.RunGH("pr", "view", prURL, "--web")
				if err != nil {
					return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
				}
				return nil
			}
		}

		return errMsg{fmt.Errorf("no PR found for this branch")}
	}
}

// resumeSessionErr returns an error tea.Cmd if the session cannot be resumed,
// or nil if the session is valid for resumption.
func resumeSessionErr(session *data.Session) tea.Cmd {
	if session == nil {
		return func() tea.Msg { return errMsg{fmt.Errorf("no session selected")} }
	}
	if session.Source != data.SourceLocalCopilot {
		return func() tea.Msg { return errMsg{fmt.Errorf("only local Copilot CLI sessions can be resumed")} }
	}
	normalizedStatus := strings.ToLower(strings.TrimSpace(session.Status))
	if normalizedStatus != "running" && normalizedStatus != "queued" && normalizedStatus != "needs-input" {
		return func() tea.Msg {
			return errMsg{fmt.Errorf("cannot resume: session status is '%s' — only running, queued, or needs-input sessions are resumable", session.Status)}
		}
	}
	if session.ID == "" {
		return func() tea.Msg { return errMsg{fmt.Errorf("cannot resume session: session has no ID")} }
	}
	return nil
}

func (m Model) resumeSession(session *data.Session) tea.Cmd {
	if errCmd := resumeSessionErr(session); errCmd != nil {
		return errCmd
	}

	// Use tea.ExecProcess to hand terminal control to the interactive
	// Copilot CLI resume command. This suspends the TUI, lets the child
	// process use stdin/stdout directly, and resumes the TUI on exit.
	c := exec.Command("gh", "copilot", "--", "--resume", session.ID)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{fmt.Errorf("resume session exited with error: %w", err)}
		}
		return nil
	})
}

func (m Model) fetchLogPoll(id string, repo string, source data.SessionSource) tea.Cmd {
	return func() tea.Msg {
		var log string
		var err error
		if source == data.SourceLocalCopilot {
			log, err = data.FetchLocalSessionLog(id)
		} else {
			log, err = data.FetchAgentTaskLog(id, repo)
		}
		return logPollResultMsg{log: log, err: err}
	}
}

// fetchPRDiff fetches the PR diff for a session, discovering the PR by branch if needed
func (m Model) fetchPRDiff(session *data.Session) tea.Cmd {
	prNumber := session.PRNumber
	repo := session.Repository
	branch := session.Branch
	return func() tea.Msg {
		// Try to discover PR if not already known
		if prNumber == 0 && repo != "" && branch != "" {
			n, url, _ := data.FetchPRForBranch(repo, branch)
			if n > 0 {
				prNumber = n
				session.PRNumber = n
				session.PRURL = url
			}
		}
		if prNumber == 0 {
			return errMsg{fmt.Errorf("no PR found for this branch")}
		}
		raw, err := data.FetchPRDiff(prNumber, repo)
		if err != nil {
			return errMsg{err}
		}
		files := diffview.ParseUnifiedDiff(raw)
		return diffLoadedMsg{files}
	}
}
