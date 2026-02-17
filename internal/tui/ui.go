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
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/help"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/kanban"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/logview"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/taskdetail"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/tasklist"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/toast"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeDetail
	ViewModeLog
	ViewModeKanban
)

// Model represents the main TUI application state
type Model struct {
	ctx         *ProgramContext
	theme       *Theme
	keys        Keybindings
	header      header.Model
	footer      footer.Model
	help        help.Model
	taskList    tasklist.Model
	taskDetail  taskdetail.Model
	logView     logview.Model
	kanban         kanban.Model
	dismissedStore *data.DismissedStore
	viewMode       ViewMode
	showPreview  bool
	ready        bool
	repo         string
	refreshInt   time.Duration
	animFrame    int
	toast        toast.Model
	prevSessions map[string]string // session ID → previous status
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

	theme := NewThemeFromConfig(ctx.Config.Theme)
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
		help:        help.New(),
		taskList:       tasklist.NewWithStore(theme.Title, theme.TableHeader, theme.TableRow, theme.TableRowSelected, theme.SectionHeader, StatusIcon, animIconFunc, dismissedStore),
		taskDetail:     taskdetail.New(theme.Title, theme.Border, StatusIcon),
		logView:        logview.New(theme.Title, 80, 20),
		kanban:         kanban.New(theme.Title, theme.Border, theme.TableRow, theme.TableRowSelected, StatusIcon, animIconFunc),
		dismissedStore: dismissedStore,
		viewMode:    ViewModeList,
		showPreview: false,
		ready:       false,
		repo:        repo,
		refreshInt:  time.Duration(refreshSeconds) * time.Second,
		toast:       toast.New(),
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
		m.help.SetSize(msg.Width, msg.Height)
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
		m.taskDetail.SetAllSessions(msg.tasks)
		m.kanban.SetSessions(msg.tasks)

		// Detect status changes and push toasts (skip first load)
		if m.prevSessions != nil {
			for _, s := range msg.tasks {
				if prev, ok := m.prevSessions[s.ID]; ok && prev != s.Status {
					m.toast.Push(StatusIcon(s.Status), s.Title, prev+" → "+s.Status)
				}
			}
		} else {
			// First load: pick the best default tab based on actual data
			m.ctx.StatusFilter = smartDefaultFilter(msg.counts)
			m.taskList.SetLoading(true)
			m.prevSessions = make(map[string]string, len(msg.tasks))
			for _, s := range msg.tasks {
				m.prevSessions[s.ID] = s.Status
			}
			return m, m.fetchTasks
		}
		// Update prevSessions for next comparison
		m.prevSessions = make(map[string]string, len(msg.tasks))
		for _, s := range msg.tasks {
			m.prevSessions[s.ID] = s.Status
		}
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
		m.toast.Tick()
		m.kanban.SetAnimFrame(m.animFrame)
		return m, m.animationTickCmd()

	case logPollTickMsg:
		if m.viewMode != ViewModeLog || !m.logView.IsLive() {
			return m, nil
		}
		session := m.taskList.SelectedTask()
		if session == nil {
			return m, nil
		}
		return m, tea.Batch(m.fetchLogPoll(session.ID, session.Repository, session.Source), m.logPollTick())

	case logPollResultMsg:
		if msg.err == nil {
			m.logView.AppendOrReplace(msg.log)
		}
		return m, nil

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
	case ViewModeKanban:
		mainView = m.kanban.View()
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

	result := headerView + mainView + footerView

	// Overlay help panel when visible
	if m.help.Visible() {
		result = m.help.View()
	}

	// Overlay toasts in the top-right corner
	if m.toast.HasToasts() {
		toastWidth := m.ctx.Width / 3
		if toastWidth < 30 {
			toastWidth = 30
		}
		m.toast.SetWidth(toastWidth)
		toastView := lipgloss.PlaceHorizontal(m.ctx.Width, lipgloss.Right, m.toast.View())
		result = toastView + "\n" + result
	}

	return result
}

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
	case ViewModeKanban:
		return m.handleKanbanKeys(msg)
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
			m.viewMode = ViewModeLog
			if isSessionRunning(session) {
				m.logView.SetLive(true)
				m.logView.SetFollowMode(true)
				return m, tea.Batch(m.fetchTaskLog(session.ID, session.Repository), m.logPollTick())
			}
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
		m.logView.SetLive(false)
		m.logView.SetFollowMode(false)
	case "j", "down":
		m.logView.SetFollowMode(false)
		m.logView.LineDown()
	case "k", "up":
		m.logView.SetFollowMode(false)
		m.logView.LineUp()
	case "d":
		m.logView.SetFollowMode(false)
		m.logView.HalfPageDown()
	case "u":
		m.logView.SetFollowMode(false)
		m.logView.HalfPageUp()
	case "g":
		m.logView.SetFollowMode(false)
		m.logView.GotoTop()
	case "G":
		m.logView.SetFollowMode(true)
		m.logView.GotoBottom()
	case "f":
		m.logView.SetFollowMode(!m.logView.FollowMode())
		if m.logView.FollowMode() {
			m.logView.GotoBottom()
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

type logPollTickMsg struct{}

type logPollResultMsg struct {
	log string
	err error
}

type errMsg struct {
	err error
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

func (m Model) logPollTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return logPollTickMsg{}
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
	if counts.Attention > 0 {
		return "attention"
	}
	if counts.Active > 0 {
		return "active"
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
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(hints)
	case ViewModeDetail:
		hints := []key.Binding{
			m.keys.NavigateBack,
			m.keys.ShowLogs,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
		m.footer.SetHints(hints)
	case ViewModeLog:
		logHints := []key.Binding{
			m.keys.NavigateBack,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "scroll")),
			m.keys.ToggleFollow,
			m.keys.ShowHelp,
			m.keys.ExitApp,
		}
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
