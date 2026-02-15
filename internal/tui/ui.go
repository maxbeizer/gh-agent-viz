package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/config"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/footer"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/header"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/logview"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/taskdetail"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/tasklist"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeDetail
	ViewModeLog
)

// Model represents the main TUI application state
type Model struct {
	ctx        *ProgramContext
	theme      *Theme
	keys       Keybindings
	header     header.Model
	footer     footer.Model
	taskList   tasklist.Model
	taskDetail taskdetail.Model
	logView    logview.Model
	viewMode   ViewMode
	ready      bool
	repo       string
	refreshInt time.Duration
}

// NewModel creates a new TUI model
func NewModel(repo string, debug bool) Model {
	ctx := NewProgramContext()
	ctx.Debug = debug
	cfg, err := config.Load("")
	if err == nil {
		ctx.Config = cfg
	} else {
		ctx.Error = fmt.Errorf("failed to load config: %w", err)
	}

	if repo == "" && len(ctx.Config.Repos) > 0 {
		repo = ctx.Config.Repos[0]
	}
	if isValidFilter(ctx.Config.DefaultFilter) {
		ctx.StatusFilter = ctx.Config.DefaultFilter
	}

	refreshSeconds := ctx.Config.RefreshInterval
	if refreshSeconds <= 0 {
		refreshSeconds = 30
	}

	theme := NewTheme()
	keys := NewKeybindings()

	// Prepare key bindings for footer
	footerKeys := []key.Binding{
		keys.MoveLeft,
		keys.MoveRight,
		keys.MoveUp,
		keys.MoveDown,
		keys.SelectTask,
		keys.ShowLogs,
		keys.OpenInBrowser,
		keys.ResumeSession,
		keys.ToggleFilter,
		keys.RefreshData,
		keys.ExitApp,
	}

	return Model{
		ctx:        ctx,
		theme:      theme,
		keys:       keys,
		header:     header.New(theme.Title, "GitHub Agent Sessions", &ctx.StatusFilter),
		footer:     footer.New(theme.Footer, footerKeys),
		taskList:   tasklist.New(theme.Title, theme.TableHeader, theme.TableRow, theme.TableRowSelected, StatusIcon),
		taskDetail: taskdetail.New(theme.Title, theme.Border, StatusIcon),
		logView:    logview.New(theme.Title, 80, 20),
		viewMode:   ViewModeList,
		ready:      false,
		repo:       repo,
		refreshInt: time.Duration(refreshSeconds) * time.Second,
	}
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchTasks,
		m.refreshCmd(),
		tea.EnterAltScreen,
	)
}

// Update handles incoming messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		m.logView.SetSize(msg.Width-4, msg.Height-8)
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tasksLoadedMsg:
		m.ctx.Error = nil
		m.taskList.SetTasks(msg.tasks)
		return m, nil

	case taskDetailLoadedMsg:
		m.ctx.Error = nil
		m.taskDetail.SetTask(msg.task)
		return m, nil

	case taskLogLoadedMsg:
		m.ctx.Error = nil
		m.logView.SetContent(msg.log)
		return m, nil

	case refreshTickMsg:
		return m, tea.Batch(m.fetchTasks, m.refreshCmd())

	case errMsg:
		m.ctx.Error = msg.err
		return m, nil
	}

	// Update the log view if in log mode
	if m.viewMode == ViewModeLog {
		var cmd tea.Cmd
		m.logView, cmd = m.logView.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Update footer hints based on current context
	m.updateFooterHints()

	headerView := m.header.View()
	footerView := m.footer.View()

	var mainView string
	switch m.viewMode {
	case ViewModeList:
		mainView = m.taskList.View()
	case ViewModeDetail:
		mainView = m.taskDetail.View()
	case ViewModeLog:
		mainView = m.logView.View()
	}

	if m.ctx.Error != nil {
		errorText := fmt.Sprintf("Error: %v", m.ctx.Error)
		if m.ctx.Debug {
			errorText = fmt.Sprintf("%s\nDebug mode enabled. Log file: %s", errorText, data.DebugLogPath())
		}
		mainView = fmt.Sprintf("%s\nPress 'r' to retry or Tab to change filter\n\n%s", errorText, mainView)
	}

	return headerView + mainView + footerView
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	}

	return m, nil
}

// handleListKeys handles keys in list view mode
func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left":
		m.taskList.MoveColumn(-1)
	case "right":
		m.taskList.MoveColumn(1)
	case "j", "down":
		m.taskList.MoveCursor(1)
	case "k", "up":
		m.taskList.MoveCursor(-1)
	case "enter":
		task := m.taskList.SelectedTask()
		if task != nil {
			m.viewMode = ViewModeDetail
			return m, m.fetchTaskDetail(task.ID, task.Repository)
		}
	case "l":
		task := m.taskList.SelectedTask()
		if task != nil {
			m.viewMode = ViewModeLog
			return m, m.fetchTaskLog(task.ID, task.Repository)
		}
	case "o":
		task := m.taskList.SelectedTask()
		if task != nil {
			return m, m.openTaskPR(task)
		}
	case "s":
		task := m.taskList.SelectedTask()
		if task != nil {
			return m, m.resumeSession(task)
		}
	case "r":
		return m, m.fetchTasks
	case "tab":
		m.cycleFilter()
		return m, m.fetchTasks
	}
	return m, nil
}

// handleDetailKeys handles keys in detail view mode
func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeList
	case "l":
		task := m.taskList.SelectedTask()
		if task != nil {
			m.viewMode = ViewModeLog
			return m, m.fetchTaskLog(task.ID, task.Repository)
		}
	case "o":
		task := m.taskList.SelectedTask()
		if task != nil {
			return m, m.openTaskPR(task)
		}
	}
	return m, nil
}

// handleLogKeys handles keys in log view mode
func (m Model) handleLogKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeList
	case "j", "down":
		m.logView.LineDown()
	case "k", "up":
		m.logView.LineUp()
	case "d":
		m.logView.HalfPageDown()
	case "u":
		m.logView.HalfPageUp()
	case "g":
		m.logView.GotoTop()
	case "G":
		m.logView.GotoBottom()
	}
	return m, nil
}

// cycleFilter cycles through status filters
func (m *Model) cycleFilter() {
	filters := []string{"all", "active", "completed", "failed"}
	for i, f := range filters {
		if f == m.ctx.StatusFilter {
			m.ctx.StatusFilter = filters[(i+1)%len(filters)]
			break
		}
	}
}

// Message types
type tasksLoadedMsg struct {
	tasks []data.Session
}

type taskDetailLoadedMsg struct {
	task *data.Session
}

type taskLogLoadedMsg struct {
	log string
}

type refreshTickMsg struct{}

type errMsg struct {
	err error
}

// fetchTasks fetches the list of sessions (both agent tasks and local sessions)
func (m Model) fetchTasks() tea.Msg {
	sessions, err := data.FetchAllSessions(m.repo)
	if err != nil {
		return errMsg{err}
	}

	// Filter sessions based on status filter
	if m.ctx.StatusFilter != "all" {
		filtered := []data.Session{}
		for _, session := range sessions {
			if m.ctx.StatusFilter == "active" && (session.Status == "running" || session.Status == "queued") {
				filtered = append(filtered, session)
			} else if session.Status == m.ctx.StatusFilter {
				filtered = append(filtered, session)
			}
		}
		sessions = filtered
	}

	return tasksLoadedMsg{sessions}
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
	return func() tea.Msg {
		log, err := data.FetchAgentTaskLog(id, repo)
		if err != nil {
			return errMsg{err}
		}
		return taskLogLoadedMsg{log}
	}
}

func (m Model) openTaskPR(session *data.Session) tea.Cmd {
	return func() tea.Msg {
		if session == nil {
			return errMsg{fmt.Errorf("no session selected")}
		}

		// Local sessions don't have PR URLs
		if session.Source == data.SourceLocalCopilot {
			return errMsg{fmt.Errorf("local sessions don't have associated pull requests")}
		}

		switch {
		case session.PRURL != "":
			output, err := exec.Command("gh", "pr", "view", session.PRURL, "--web").CombinedOutput()
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
		case session.PRNumber > 0 && session.Repository != "":
			output, err := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", session.PRNumber), "-R", session.Repository, "--web").CombinedOutput()
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
		default:
			return errMsg{fmt.Errorf("selected session has no pull request to open")}
		}

		return nil
	}
}

func (m Model) resumeSession(session *data.Session) tea.Cmd {
	return func() tea.Msg {
		if session == nil {
			return errMsg{fmt.Errorf("no session selected")}
		}

		if session.Source != data.SourceLocalCopilot {
			return errMsg{fmt.Errorf("only local Copilot CLI sessions can be resumed")}
		}

		// Only allow resuming active sessions (running or queued)
		if session.Status != "running" && session.Status != "queued" {
			return errMsg{fmt.Errorf("cannot resume session: session status is '%s' (only 'running' or 'queued' sessions can be resumed)", session.Status)}
		}

		if session.ID == "" {
			return errMsg{fmt.Errorf("cannot resume session: session has no ID")}
		}

		cmd := exec.Command("gh", "copilot", "--", "--resume", session.ID)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Provide a user-friendly error message
			outputStr := strings.TrimSpace(string(output))
			if outputStr != "" {
				// Include output only if it provides useful context
				return errMsg{fmt.Errorf("failed to resume session: %s", outputStr)}
			}
			return errMsg{fmt.Errorf("failed to resume session")}
		}

		return nil
	}
}

func (m Model) refreshCmd() tea.Cmd {
	return tea.Tick(m.refreshInt, func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func isValidFilter(filter string) bool {
	switch filter {
	case "all", "active", "completed", "failed":
		return true
	default:
		return false
	}
}

// updateFooterHints updates footer hints based on current view mode and state
func (m *Model) updateFooterHints() {
	switch m.viewMode {
	case ViewModeList:
		m.footer.SetHints([]key.Binding{
			m.keys.MoveUp,
			m.keys.MoveDown,
			m.keys.SelectTask,
			m.keys.ShowLogs,
			m.keys.OpenInBrowser,
			m.keys.ToggleFilter,
			m.keys.RefreshData,
			m.keys.ExitApp,
		})
	case ViewModeDetail:
		m.footer.SetHints([]key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			m.keys.ShowLogs,
			m.keys.OpenInBrowser,
			m.keys.ExitApp,
		})
	case ViewModeLog:
		m.footer.SetHints([]key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			key.NewBinding(key.WithKeys("↑/k"), key.WithHelp("↑/k", "up")),
			key.NewBinding(key.WithKeys("↓/j"), key.WithHelp("↓/j", "down")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "page down")),
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "page up")),
			key.NewBinding(key.WithKeys("g/G"), key.WithHelp("g/G", "top/bottom")),
			m.keys.ExitApp,
		})
	}
}
