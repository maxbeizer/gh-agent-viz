package tui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/config"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/conversation"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/footer"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/gitactivity"
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
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/statsbar"
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
	ViewModeGitActivity
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
	gitActivity    gitactivity.Model
	dismissedStore *data.DismissedStore
	statsBar       statsbar.Model
	viewMode       ViewMode
	showConversation bool // true when conversation bubble view is active in log mode
	showPreview  bool
	ready        bool
	repo         string
	refreshInt   time.Duration
	animFrame    int
	toast        toast.Model
	prevSessions map[string]string // session ID → previous status
	demo         bool
	allSessions  []data.Session // accumulated across load phases
	initialLoadDone bool       // true after all initial load phases complete
	lastFingerprint string     // hash of session data; used to skip no-op refreshes
	searchActive bool          // true when search input is active
	searchQuery  string        // current search filter text
	snapshotPath string        // if set, write snapshot on initial load and quit
	loadSpinner  spinner.Model // animated spinner shown during initial load
	loadTagline  string        // randomized tagline for the loading screen
}

// NewModel creates a new TUI model
func NewModel(repo string, debug bool, demo bool, snapshotPath string) Model {
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

	// Loading screen spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"})
	tagline := loadingTaglines[rand.Intn(len(loadingTaglines))]

	// Determine default view mode
	defaultView := ViewModeMission // dashboard-first by default
	switch ctx.Config.DefaultView {
	case "table":
		defaultView = ViewModeList
	case "kanban":
		defaultView = ViewModeKanban
	case "dashboard", "mission", "":
		defaultView = ViewModeMission
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
		gitActivity:    gitactivity.New(80, 20),
		dismissedStore: dismissedStore,
		statsBar:       statsbar.New(),
		viewMode:    defaultView,
		showPreview: false,
		ready:       false,
		repo:        repo,
		refreshInt:  time.Duration(refreshSeconds) * time.Second,
		toast:        toast.New(),
		prevSessions: make(map[string]string, 64),
		demo:         demo,
		snapshotPath: snapshotPath,
		loadSpinner: sp,
		loadTagline: tagline,
	}
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.loadSpinner.Tick,
		m.fetchLocalSessions,  // Phase 1: fast, shows content immediately
		m.fetchAgentTasks,     // Phase 2: runs concurrently, returns when API responds
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
		m.help.SetSize(msg.Width, msg.Height)
		m.statsBar.SetWidth(msg.Width)
		m.updateSplitLayout()
		m.ready = true
		// Debounce heavy component resizes (logview, diffview, etc.)
		// so rapid resize events don't each trigger a full re-render.
		return m, tea.Tick(
			50*time.Millisecond,
			func(time.Time) tea.Msg { return resizeDebouncedMsg{} },
		)

	case spinner.TickMsg:
		if !m.initialLoadDone {
			var cmd tea.Cmd
			m.loadSpinner, cmd = m.loadSpinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case resizeDebouncedMsg:
		m.logView.SetSize(m.ctx.Width-4, m.ctx.Height-8)
		m.toolTimeline.SetSize(m.ctx.Width-4, m.ctx.Height-8)
		m.conversationView.SetSize(m.ctx.Width-4, m.ctx.Height-8)
		m.diffView.SetSize(m.ctx.Width-4, m.ctx.Height-8)
		m.gitActivity.SetSize(m.ctx.Width-4, m.ctx.Height-8)
		m.mission.SetSize(m.ctx.Width, m.ctx.Height-4)
		m.kanban.SetSize(m.ctx.Width, m.ctx.Height-4)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case localSessionsLoadedMsg:
		// Phase 1: show local sessions immediately
		m.mergeSessions(msg.sessions)
		// Kick off token usage loading after first render
		return m, m.fetchTokenUsage

	case agentTasksLoadedMsg:
		// Phase 2: merge agent tasks into existing sessions
		if msg.sessions != nil {
			m.mergeSessions(msg.sessions)
		}
		return m, nil

	case tokenUsageLoadedMsg:
		// Phase 3: enrich sessions with token data
		if msg.usage != nil {
			m.enrichTokenUsage(msg.usage)
		}
		// Initial load complete — snapshot prevSessions and start refresh timer
		if !m.initialLoadDone {
			m.initialLoadDone = true
			clear(m.prevSessions)
			for _, s := range m.allSessions {
				m.prevSessions[s.ID] = s.Status
			}
			if m.snapshotPath != "" {
				m.writeSnapshot()
				return m, tea.Quit
			}
		}
		return m, m.refreshCmd()

	case tasksLoadedMsg:
		m.ctx.Error = nil
		m.ctx.Counts = msg.counts
		m.header.SetCounts(header.FilterCounts{
			All:       msg.counts.All,
			Attention: msg.counts.Attention,
			Warning:   msg.counts.Warning,
			Active:    msg.counts.Active,
			Completed: msg.counts.Completed,
			Failed:    msg.counts.Failed,
		})
		m.taskList.SetLoading(false)
		m.taskList.SetTasks(msg.tasks)
		// Only push to the active secondary view
		switch m.viewMode {
		case ViewModeKanban:
			m.kanban.SetSessions(msg.allSessions)
		case ViewModeMission:
			m.mission.SetSessions(msg.allSessions)
		default:
			m.taskDetail.SetAllSessions(msg.allSessions)
		}

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
			clear(m.prevSessions)
			for _, s := range msg.tasks {
				m.prevSessions[s.ID] = s.Status
			}
			return m, m.fetchTasks
		}
		// Update prevSessions for next comparison
		clear(m.prevSessions)
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

	case gitDiffLoadedMsg:
		m.ctx.Error = nil
		m.gitActivity.SetDiffResult(msg.result)
		return m, nil

	case gitDiffPollTickMsg:
		if m.viewMode != ViewModeGitActivity {
			return m, nil
		}
		session := m.taskList.SelectedTask()
		if session == nil || session.WorkDir == "" {
			return m, nil
		}
		return m, tea.Batch(m.fetchGitDiff(session.WorkDir), m.gitDiffPollTick())

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

	// Update the git activity view if active
	if m.viewMode == ViewModeGitActivity {
		var cmd tea.Cmd
		m.gitActivity, cmd = m.gitActivity.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return "Loading sessions..."
	}

	// Show loading screen until initial data load completes
	if !m.initialLoadDone {
		return m.viewLoading()
	}

	// Update footer hints based on current context
	m.updateFooterHints()

	// All views: show banner + stats bar, no tab bar
	chrome := m.header.ViewBannerOnly() + m.statsBar.View() + "\n"
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
	case ViewModeGitActivity:
		mainView = m.gitActivity.View()
	}

	if m.ctx.Debug {
		mainView = fmt.Sprintf("DEBUG ON • command logs: %s\n", data.DebugLogPath()) + mainView
	}

	// Show toasts just above the footer
	toastView := ""
	if m.toast.HasToasts() {
		toastView = "\n" + m.toast.View()
	}

	// Search bar indicator
	searchView := ""
	if m.searchActive || m.searchQuery != "" {
		searchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"}).
			Bold(true)
		queryDisplay := m.searchQuery
		if m.searchActive {
			queryDisplay += "▍" // cursor
		}
		searchView = searchStyle.Render(fmt.Sprintf("  🔍 Filter: %s", queryDisplay)) + "\n"
	}

	result := chrome + mainView + toastView + searchView + footerView

	// Overlay help panel when visible
	if m.help.Visible() {
		result = m.help.View()
	}

	return result
}

// loadingTaglines are randomly selected for the loading screen.
var loadingTaglines = []string{
	"Waking up the agents…",
	"Scanning the multiverse for sessions…",
	"Tuning into the agent frequency…",
	"Herding digital cats…",
	"Reticulating splines…",
	"Consulting the oracle…",
	"Brewing a fresh batch of data…",
	"Spinning up the flux capacitor…",
}

// viewLoading renders a centered animated loading screen.
func (m Model) viewLoading() string {
	logo := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"}).
		Render("⚡ Agent Viz")

	spinnerLine := m.loadSpinner.View() + " " + lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "242", Dark: "245"}).
		Italic(true).
		Render(m.loadTagline)

	block := lipgloss.JoinVertical(lipgloss.Center, "", logo, "", spinnerLine, "")

	return lipgloss.Place(
		m.ctx.Width, m.ctx.Height,
		lipgloss.Center, lipgloss.Center,
		block,
	)
}

// viewModeName returns a human-readable name for the current view mode.
func (m Model) viewModeName() string {
	switch m.viewMode {
	case ViewModeList:
		return "list"
	case ViewModeDetail:
		return "detail"
	case ViewModeLog:
		return "log"
	case ViewModeKanban:
		return "kanban"
	case ViewModeToolTimeline:
		return "timeline"
	case ViewModeMission:
		return "dashboard"
	case ViewModeDiff:
		return "diff"
	case ViewModeGitActivity:
		return "git-activity"
	default:
		return "unknown"
	}
}

// writeSnapshot captures the current TUI state to snapshotPath as JSON.
func (m Model) writeSnapshot() {
	sessions := make([]data.SnapshotSession, len(m.allSessions))
	for i, s := range m.allSessions {
		sessions[i] = data.SnapshotSession{
			ID:             s.ID,
			Status:         s.Status,
			Title:          s.Title,
			Repository:     s.Repository,
			AttentionLevel: data.SessionAttentionLevel(s).String(),
		}
	}
	snap := &data.Snapshot{
		ViewMode: m.viewModeName(),
		TerminalSize: data.SnapshotSize{
			Width:  m.ctx.Width,
			Height: m.ctx.Height,
		},
		RenderedOutput: m.View(),
		SessionCount:   len(m.allSessions),
		FilterCounts: data.SnapshotCounts{
			All:       m.ctx.Counts.All,
			Attention: m.ctx.Counts.Attention,
			Warning:   m.ctx.Counts.Warning,
			Active:    m.ctx.Counts.Active,
			Completed: m.ctx.Counts.Completed,
			Failed:    m.ctx.Counts.Failed,
		},
		Sessions:     sessions,
		FocusedPanel: m.viewModeName(),
	}
	_ = data.WriteSnapshot(m.snapshotPath, snap)
}

