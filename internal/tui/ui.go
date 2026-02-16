package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	ctx         *ProgramContext
	theme       *Theme
	keys        Keybindings
	header      header.Model
	footer      footer.Model
	taskList    tasklist.Model
	taskDetail  taskdetail.Model
	logView     logview.Model
	viewMode    ViewMode
	showPreview bool
	ready       bool
	repo        string
	refreshInt  time.Duration
	animFrame   int
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
		keys.MoveUp,
		keys.MoveDown,
		keys.SelectTask,
		keys.ToggleFilter,
		keys.FocusAttention,
		keys.RefreshData,
		keys.ExitApp,
	}

	dismissedStore := data.NewDismissedStore()

	var animIconFunc func(string, int) string
	if ctx.Config.AnimationsEnabled() {
		animIconFunc = AnimatedStatusIcon
	}

	return Model{
		ctx:         ctx,
		theme:       theme,
		keys:        keys,
		header:      header.New(theme.Title, theme.TabActive, theme.TabInactive, theme.TabCount, "⚡ Agent Sessions", &ctx.StatusFilter, ctx.Config.AsciiHeaderEnabled()),
		footer:      footer.New(theme.Footer, footerKeys),
		taskList:    tasklist.NewWithStore(theme.Title, theme.TableHeader, theme.TableRow, theme.TableRowSelected, StatusIcon, animIconFunc, dismissedStore),
		taskDetail:  taskdetail.New(theme.Title, theme.Border, StatusIcon),
		logView:     logview.New(theme.Title, 80, 20),
		viewMode:    ViewModeList,
		showPreview: false,
		ready:       false,
		repo:        repo,
		refreshInt:  time.Duration(refreshSeconds) * time.Second,
	}
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.fetchTasks,
		m.refreshCmd(),
		tea.EnterAltScreen,
	}
	if m.ctx.Config.AnimationsEnabled() {
		cmds = append(cmds, m.animationTickCmd())
	}
	return tea.Batch(cmds...)
}

// Update handles incoming messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		m.header.SetSize(msg.Width, msg.Height)
		m.logView.SetSize(msg.Width-4, msg.Height-8)
		m.updateSplitLayout()
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tasksLoadedMsg:
		m.ctx.Error = nil
		m.ctx.Counts = msg.counts
		m.header.SetCounts(header.FilterCounts{
			All:       msg.counts.All,
			Attention: msg.counts.Attention,
			Active:    msg.counts.Active,
			Completed: msg.counts.Completed,
			Failed:    msg.counts.Failed,
		})
		m.taskList.SetLoading(false)
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

	case animationTickMsg:
		m.animFrame++
		m.taskList.SetAnimFrame(m.animFrame)
		return m, m.animationTickCmd()

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
		return "Loading sessions..."
	}

	// Update footer hints based on current context
	m.updateFooterHints()

	headerView := m.header.View()
	footerView := m.footer.View()

	var mainView string
	switch m.viewMode {
	case ViewModeList:
		if m.previewVisible() {
			selected := m.taskList.SelectedTask()
			if selected != nil {
				m.taskDetail.SetTask(selected)
			}
			leftContent := m.taskList.View()
			rightContent := m.taskDetail.ViewSplit()
			mainView = lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
		} else {
			mainView = m.taskList.View()
		}
	case ViewModeDetail:
		mainView = m.taskDetail.View()
	case ViewModeLog:
		mainView = m.logView.View()
	}

	debugBanner := ""
	if m.ctx.Debug {
		debugBanner = fmt.Sprintf("DEBUG ON • command logs: %s\n", data.DebugLogPath())
	}

	if m.ctx.Error != nil {
		errorText := fmt.Sprintf("Error: %v", m.ctx.Error)
		if m.ctx.Debug {
			errorText = fmt.Sprintf("%s\nInspect debug log for command output.", errorText)
		}
		mainView = fmt.Sprintf("%s%s\nPress 'r' to retry or Tab/Shift+Tab to change filter\n\n%s", debugBanner, errorText, mainView)
	} else if debugBanner != "" {
		mainView = debugBanner + mainView
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
	case "j", "down":
		m.taskList.MoveCursor(1)
	case "k", "up":
		m.taskList.MoveCursor(-1)
	case "enter":
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
			if session.Source == data.SourceLocalCopilot {
				m.ctx.Error = fmt.Errorf("logs are only available for remote agent-task sessions")
				return m, nil
			}
			m.viewMode = ViewModeLog
			return m, m.fetchTaskLog(session.ID, session.Repository)
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
		m.cycleFilter(1)
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
			if session.Source == data.SourceLocalCopilot {
				m.ctx.Error = fmt.Errorf("logs are only available for remote agent-task sessions")
				return m, nil
			}
			m.viewMode = ViewModeLog
			return m, m.fetchTaskLog(session.ID, session.Repository)
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
	case "s":
		session := m.taskList.SelectedTask()
		if session != nil {
			return m, m.resumeSession(session)
		}
	}
	return m, nil
}

// cycleFilter cycles through status filters by delta (+1 forward, -1 backward)
func (m *Model) cycleFilter(delta int) {
	filters := []string{"attention", "active", "completed", "failed", "all"}
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

// Message types
type tasksLoadedMsg struct {
	tasks  []data.Session
	counts FilterCounts
}

type taskDetailLoadedMsg struct {
	task *data.Session
}

type taskLogLoadedMsg struct {
	log string
}

type refreshTickMsg struct{}

type animationTickMsg struct{}

type errMsg struct {
	err error
}

// fetchTasks fetches the list of sessions (both agent tasks and local sessions)
func (m Model) fetchTasks() tea.Msg {
	sessions, err := data.FetchAllSessions(m.repo)
	if err != nil {
		return errMsg{err}
	}

	// Compute counts across ALL sessions before filtering
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

	return tasksLoadedMsg{sessions, counts}
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
			output, err := data.RunGH("pr", "view", session.PRURL, "--web")
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
		case session.PRNumber > 0 && session.Repository != "":
			output, err := data.RunGH("pr", "view", fmt.Sprintf("%d", session.PRNumber), "-R", session.Repository, "--web")
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
		default:
			return errMsg{fmt.Errorf("selected session has no pull request to open")}
		}

		return nil
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

func isValidFilter(filter string) bool {
	switch filter {
	case "all", "attention", "active", "completed", "failed":
		return true
	default:
		return false
	}
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
		hints := []key.Binding{
			m.keys.MoveUp,
			m.keys.MoveDown,
			m.keys.SelectTask,
		}
		selected := m.taskList.SelectedTask()
		if canShowLogs(selected) {
			hints = append(hints, m.keys.ShowLogs)
		}
		if canOpenPR(selected) {
			hints = append(hints, m.keys.OpenInBrowser)
		}
		if canResumeLocalSession(selected) {
			hints = append(hints, m.keys.ResumeSession)
		}
		if selected != nil {
			hints = append(hints, m.keys.DismissSession)
		}
		hints = append(hints, m.keys.TogglePreview, m.keys.ToggleFilter, m.keys.FocusAttention, m.keys.RefreshData, m.keys.ExitApp)
		m.footer.SetHints(hints)
	case ViewModeDetail:
		hints := []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
		selected := m.taskList.SelectedTask()
		if canShowLogs(selected) {
			hints = append(hints, m.keys.ShowLogs)
		}
		if canOpenPR(selected) {
			hints = append(hints, m.keys.OpenInBrowser)
		}
		if canResumeLocalSession(selected) {
			hints = append(hints, m.keys.ResumeSession)
		}
		hints = append(hints, m.keys.ExitApp)
		m.footer.SetHints(hints)
	case ViewModeLog:
		logHints := []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			key.NewBinding(key.WithKeys("↑/k"), key.WithHelp("↑/k", "up")),
			key.NewBinding(key.WithKeys("↓/j"), key.WithHelp("↓/j", "down")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "page down")),
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "page up")),
			key.NewBinding(key.WithKeys("g/G"), key.WithHelp("g/G", "top/bottom")),
		}
		selected := m.taskList.SelectedTask()
		if canResumeLocalSession(selected) {
			logHints = append(logHints, m.keys.ResumeSession)
		}
		logHints = append(logHints, m.keys.ExitApp)
		m.footer.SetHints(logHints)
	}
}

func canShowLogs(session *data.Session) bool {
	return session != nil && session.Source == data.SourceAgentTask && strings.TrimSpace(session.ID) != ""
}

func canOpenPR(session *data.Session) bool {
	if session == nil || session.Source != data.SourceAgentTask {
		return false
	}
	if strings.TrimSpace(session.PRURL) != "" {
		return true
	}
	return session.PRNumber > 0 && strings.TrimSpace(session.Repository) != ""
}

func canResumeLocalSession(session *data.Session) bool {
	if session == nil || session.Source != data.SourceLocalCopilot || strings.TrimSpace(session.ID) == "" {
		return false
	}
	status := strings.ToLower(strings.TrimSpace(session.Status))
	return status == "running" || status == "queued" || status == "needs-input"
}
