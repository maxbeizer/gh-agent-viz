package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/config"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/conversation"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/footer"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/header"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/help"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/kanban"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/logview"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/mission"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/diffview"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/taskdetail"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/tasklist"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/toast"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/tooltimeline"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeDetail
	ViewModeLog
	ViewModeKanban
	ViewModeToolTimeline
	ViewModeMission
	ViewModeDiff
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
	diffView    diffview.Model
	kanban         kanban.Model
	toolTimeline   tooltimeline.Model
	conversationView conversation.Model
	mission        mission.Model
	dismissedStore *data.DismissedStore
	viewMode       ViewMode
	showConversation bool // true when conversation bubble view is active in log mode
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
		diffView:       diffview.New(80, 20),
		kanban:         kanban.New(theme.Title, theme.Border, theme.TableRow, theme.TableRowSelected, StatusIcon, animIconFunc),
		toolTimeline:   tooltimeline.New(80, 20),
		conversationView: conversation.New(80, 20),
		mission:        mission.New(theme.Title, theme.TableRow, theme.TableRowSelected, StatusIcon, animIconFunc),
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
		m.toolTimeline.SetSize(msg.Width-4, msg.Height-8)
		m.conversationView.SetSize(msg.Width-4, msg.Height-8)
		m.diffView.SetSize(msg.Width-4, msg.Height-8)
		m.help.SetSize(msg.Width, msg.Height)
		m.mission.SetSize(msg.Width, msg.Height-4)
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
		m.taskDetail.SetAllSessions(msg.allSessions)
		m.kanban.SetSessions(msg.allSessions)
		m.mission.SetSessions(msg.allSessions)

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

	case toolTimelineLoadedMsg:
		m.ctx.Error = nil
		m.toolTimeline.SetEvents(msg.events)
		return m, nil

	case diffLoadedMsg:
		m.ctx.Error = nil
		m.diffView.SetDiffs(msg.files)
		return m, nil

	case refreshTickMsg:
		return m, tea.Batch(m.fetchTasks, m.refreshCmd())

	case animationTickMsg:
		m.animFrame++
		m.taskList.SetAnimFrame(m.animFrame)
		m.toast.Tick()
		m.kanban.SetAnimFrame(m.animFrame)
		m.mission.SetAnimFrame(m.animFrame)
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

	case conversationLoadedMsg:
		m.conversationView.SetMessages(msg.messages)
		m.showConversation = true
		return m, nil

	case errMsg:
		// Show errors as toasts — they auto-dismiss and don't disrupt the view
		m.toast.Push("⚠️", "Error", msg.err.Error())
		return m, nil
	}

	// Update the log view if in log mode
	if m.viewMode == ViewModeLog {
		if m.showConversation {
			var cmd tea.Cmd
			m.conversationView, cmd = m.conversationView.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.logView, cmd = m.logView.Update(msg)
		return m, cmd
	}

	// Update the tool timeline if in timeline mode
	if m.viewMode == ViewModeToolTimeline {
		var cmd tea.Cmd
		m.toolTimeline, cmd = m.toolTimeline.Update(msg)
		return m, cmd
	}

	// Update the diff view if in diff mode
	if m.viewMode == ViewModeDiff {
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
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
		if m.showConversation {
			mainView = m.conversationView.View()
		} else {
			mainView = m.logView.View()
		}
	case ViewModeKanban:
		mainView = m.kanban.View()
	case ViewModeToolTimeline:
		mainView = m.toolTimeline.View()
	case ViewModeMission:
		mainView = m.mission.View()
	case ViewModeDiff:
		mainView = m.diffView.View()
	}

	if m.ctx.Debug {
		mainView = fmt.Sprintf("DEBUG ON • command logs: %s\n", data.DebugLogPath()) + mainView
	}

	// Show toasts just above the footer
	toastView := ""
	if m.toast.HasToasts() {
		toastView = "\n" + m.toast.View()
	}

	result := headerView + mainView + toastView + footerView

	// Overlay help panel when visible
	if m.help.Visible() {
		result = m.help.View()
	}

	return result
}

