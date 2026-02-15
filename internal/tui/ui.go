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
func NewModel(repo string) Model {
	ctx := NewProgramContext()
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
		mainView = fmt.Sprintf("Error: %v\n\n%s", m.ctx.Error, mainView)
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
	tasks []data.AgentTask
}

type taskDetailLoadedMsg struct {
	task *data.AgentTask
}

type taskLogLoadedMsg struct {
	log string
}

type refreshTickMsg struct{}

type errMsg struct {
	err error
}

// fetchTasks fetches the list of agent tasks
func (m Model) fetchTasks() tea.Msg {
	tasks, err := data.FetchAgentTasks(m.repo)
	if err != nil {
		return errMsg{err}
	}

	// Filter tasks based on status filter
	if m.ctx.StatusFilter != "all" {
		filtered := []data.AgentTask{}
		for _, task := range tasks {
			if m.ctx.StatusFilter == "active" && (task.Status == "running" || task.Status == "queued") {
				filtered = append(filtered, task)
			} else if task.Status == m.ctx.StatusFilter {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	return tasksLoadedMsg{tasks}
}

// fetchTaskDetail fetches detailed information for a task
func (m Model) fetchTaskDetail(id string, repo string) tea.Cmd {
	return func() tea.Msg {
		task, err := data.FetchAgentTaskDetail(id, repo)
		if err != nil {
			return errMsg{err}
		}
		return taskDetailLoadedMsg{task}
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

func (m Model) openTaskPR(task *data.AgentTask) tea.Cmd {
	return func() tea.Msg {
		if task == nil {
			return errMsg{fmt.Errorf("no task selected")}
		}

		switch {
		case task.PRURL != "":
			output, err := exec.Command("gh", "pr", "view", task.PRURL, "--web").CombinedOutput()
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
		case task.PRNumber > 0 && task.Repository != "":
			output, err := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", task.PRNumber), "-R", task.Repository, "--web").CombinedOutput()
			if err != nil {
				return errMsg{fmt.Errorf("failed to open PR: %s", strings.TrimSpace(string(output)))}
			}
		default:
			return errMsg{fmt.Errorf("selected task has no pull request to open")}
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
