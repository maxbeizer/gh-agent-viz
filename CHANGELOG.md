# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com),
and this project adheres to [Semantic Versioning](https://semver.org).

## [v0.10.1] - 2026-03-30

### Fixed

- **Reduced CPU usage** — animation tick loop now stops entirely when no sessions are active and no toasts are visible, instead of ticking indefinitely. Fast animation interval slowed from 100ms to 200ms.
- **Split view caching** — detail pane in split view no longer recomputes dependency graph on every render; caches by selected session ID and invalidates on data refresh.

### Added

- `--profile <file>` flag — captures a CPU profile (pprof format) for the duration of the session. Analyze with `go tool pprof -http=:8080 <file>`.

## [v0.10.0] - 2026-03-27

### Added

- **Active Sessions view** — press `A` from any view to see a lazygit-style split panel focused on what your agents are doing right now
  - Left panel: compact 2-line session list with status breakdown (e.g. "3 running · 1 failed · 1 waiting")
  - Right panel: detail pane with metadata (repo, branch, PR, elapsed, model, tokens), current activity, and recent log tail from `events.jsonl`
  - Responsive layout: side-by-side on wide terminals (≥100 cols), stacked on narrow
  - Quick actions: `enter` details, `o` open PR, `l` logs, `c` copy session ID, `x` dismiss
  - Only shows actively working sessions (not idle) plus needs-input and failed
  - Configurable as default view via `defaultView: active` in `.gh-agent-viz.yml`
- **Powerline footer** across all views — gh-inbox-inspired status bar with colored view badge (⚡ Active, 🚀 Mission, 📋 List, etc.) and compact keybinding hints
- **Clipboard support** — copy session ID to clipboard (`c` key in active view)
- `A` hotkey added to help overlay

### Changed

- **Default theme is now Catppuccin Mocha** — rich dark palette with lavender titles, surface-layered panels, and pastel accents. Previous theme available via `theme: default` in config.
- **Upgraded to Charm v2 ecosystem**: Bubble Tea v2.0.2, Lip Gloss v2.0.2, Bubbles v2.0.0 (requires Go 1.25+)
  - Declarative `View()` API (returns `tea.View` with AltScreen/MouseMode fields)
  - Enhanced keyboard handling (`tea.KeyPressMsg`)
  - Split mouse event types (`MouseWheelMsg`, `MouseClickMsg`)
  - New cursed renderer for faster, smoother rendering

### Fixed

- Footer pinned to the bottom of the terminal in all views (no more floating footer)
- Active view respects terminal height — proper scrolling with position indicator
- Session count in active view matches stats bar (excludes idle sessions)

## [v0.9.3] - 2026-03-19

### Performance

- Reduced idle CPU usage: switched from WithMouseAllMotion to WithMouseCellMotion, adaptive animation tick (500ms idle vs 100ms active)

## [v0.9.2] - 2026-03-19

### Added

- Version displayed on the loading/boot screen below the logo

## [v0.9.1] - 2026-03-19

### Fixed

- Version upgrade nudge no longer shows when already on the latest version (v prefix normalization)

## [v0.9.0] - 2026-03-19

### Added

- Version display in header banner next to tagline
- Automatic upgrade check on startup — shows amber "⬆ vX.Y.Z available" nudge when a newer release exists

### Fixed

- Mission control and kanban now always have current session data when switching views

## [v0.8.2] - 2026-03-19

### Fixed

- Mission control and kanban now always have current session data when switching views — previously showed stale/empty data if loaded on a different view

## [v0.8.1] - 2026-03-18

### Fixed

- Active panel count now matches the fleet summary bar — needs-input sessions are no longer double-counted as "active" (they appear in Attention instead)

## [v0.8.0] - 2026-03-17

### Fixed

- Mission Control dashboard no longer clips the top when many sessions are present
- Budget-based height allocation ensures panels never exceed terminal height
- Idle panel entries no longer wrap to 2 lines in the right column
- Fleet panel is treated as fixed-size and no longer shows erroneous "▼ N more"

### Added

- Scrollable panels in Mission Control — j/k scrolls within the focused panel with `▲ N above` / `▼ N more` indicators
- Tab-based navigation for narrow terminals (width < 100) — one section at a time with a highlighted tab bar showing counts
- Scroll offset follows cursor so the selected item is always visible

## [v0.7.0] - 2026-03-01

### Added

- Live Git Activity view — press `G` to see real-time uncommitted changes in an agent's working directory
- Auto-polling every 3 seconds with colored diff output and file stats
- Parse `cwd`/`git_root` from workspace.yaml into `Session.WorkDir`

## [v0.6.0] - 2026-03-01

### Added

- Animated loading screen with ⚡ branding, spinner, and fun randomized taglines shown during initial data fetch
- Replaces empty containers with a polished startup experience

## [v0.5.0] - 2026-02-28

### Added

- Direct Copilot API client (`internal/data/capi/`) — calls `api.githubcopilot.com` directly instead of shelling out to `gh agent-task`
- Structured JSON responses with access to all API fields (timestamps, model, premium_requests, error details)
- Graceful fallback to CLI subprocess when CAPI auth is unavailable
- `go-gh` dependency for auth token retrieval

### Changed

- `FetchAgentTasks`, `FetchAgentTaskDetail`, and `FetchAgentTaskLog` now try CAPI first
- Updated documentation across 5 files to reflect new data-fetching architecture

## [v0.3.0] - 2026-02-20

### Added

- Progressive loading — show local sessions immediately

## [v0.2.0] - 2026-02-20

### Fixed

- Correct GoReleaser config so `gh extension install` works

## [v0.1.0] - 2026-02-19

### Added

- TUI extension for visualizing Copilot agent sessions
- Local Copilot CLI session ingestion with unified Session model
- Resume-session action for active CLI sessions
- Session usage telemetry and org metrics support
- Focused-pane UX redesign with filter tab bar
- Split-pane layout with detail preview sidebar
- Animated braille spinners for running/queued sessions
- ASCII art header branding banner
- Rich markdown log rendering with glamour
- Configurable color theme system with presets
- Live log tailing with follow mode for running sessions
- Sparkline activity indicators to session list
- Toast notification system for status changes
- Kanban view mode with status columns
- Session dependency graph in detail view
- Session timeline bar visualization in detail view
- Help overlay and simplified footer hints
- Explicit attention reasons in badges and detail view
- Smart default tab — land on attention, running, or all
- Conversation view with styled chat bubbles
- Tool execution timeline view
- Mission control dashboard view
- Session diff view for reviewing agent code changes
- `!` and `@` shortcuts to view repo and file issues
- Link local sessions to PRs via branch name lookup
- `--demo` mode with fake data for screenshots and GIF recording
- Copilot token usage integration from CLI log files
- `d` (diff) and `t` (tools) keys to list view
- Copilot-setup-steps for coding agent environment
- Integration smoke tests for pre-merge confidence
- Unit tests and CI workflow
- Operator guide and troubleshooting documentation
- Debug mode diagnostics and test-agent coverage
- GitHub Pages docs pipeline and analytics contract
- Developer workflow and onboarding docs
- Copilot instructions for gh-agent-viz
- Stable regression tests for data, config, logview, taskdetail

### Changed

- Tasklist UX overhaul — compact cards, dedup info, auto-grouping
- Toggle-able session grouping by repo/status/source
- Visual overhaul — status-tinted rows, warmer palette, hierarchy
- Overhaul mission control into summary dashboard
- Subtler running animation, kanban columns → in progress/idle/done
- Reorder tabs — RUNNING first, ATTENTION last
- Recency-first sort and dismiss sessions with x
- Rework attention model — only needs-input and failed need attention
- Replace braille spinners with pulsing color indicators
- Refactor: split ui.go into commands, keyhandlers, and helpers
- Refactor: consolidate duplicated helpers into data package
- Surface session duration and telemetry in UI
- Polish task list UX for high-signal triage
- Improve board readability and action framing
- Make action hints contextual to selected session
- Improve session card and selected panel clarity
- Improve vertical resize and attention-first session triage
- Improve responsive resize and input-needed detection
- Prioritize actionable sessions in task list
- Require Go 1.24.2+ and fail fast on older toolchains
- Refocus documentation on product usage
- Visual polish — colored gutters, dividers, and depth

### Fixed

- Resolve lint errors and add GoReleaser for distribution
- UX polish — toast feedback for unavailable actions, better empty states
- Pre-launch audit fixes — permissions, help, docs, version
- Replace 🧑 with ✋ for inclusivity
- Token parser handles multi-line JSON, colored status icons
- PR tag rendered with color (not faint), add solarized-light theme
- Diff esc returns to list, PR indicator as inline text
- Diff loading state and PR branch indicator in list
- Move toasts above footer instead of top of app
- PR lookup includes merged PRs, skips default branches
- Stop animating idle sessions in kanban and mission control
- Start on RUNNING tab to prevent attention flash on startup
- Pagination, header, and cursor gutter in mission control
- Tagline positioning and background fixes
- Enter expands collapsed groups, c key works from list/detail
- Exclude dismissed sessions from tab counts
- Make attention badges actionable instead of speculative
- Truncate footer key hints to fit terminal width
- Age out stale sessions from ACTION tab after 4 hours
- Improve duplicate session UX with count indicators
- Replace vague "⚠ check progress" badge with contextual idle duration
- Make "a" key jump to ACTION tab
- Enable log viewing for local Copilot CLI sessions
- Default preview pane to closed on startup and tab switch
- Show loading indicator on initial startup instead of empty state
- Persist dismissed sessions across restarts
- Make 'a' key cycle filters and reorder tabs with ALL last
- Use tea.ExecProcess for resume session to get interactive TTY
- Default to action tab, add filter transition feedback, whimsical empty states
- Make resume session (s key) reliable across all views
- Fix local session parsing and board labels
- Open PR action for full URL targets
- Make relink-local rebuild binary before reinstalling
- Make docs workflow non-blocking before Pages enablement

### Removed

- Dead quiet duplicate de-emphasis code
- Sparkline activity indicators
- Redundant Agent Sessions title from header
- Remove background bleed from tagline

[v0.3.0]: https://github.com/maxbeizer/gh-agent-viz/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/maxbeizer/gh-agent-viz/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/maxbeizer/gh-agent-viz/releases/tag/v0.1.0
