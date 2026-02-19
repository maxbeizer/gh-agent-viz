# gh-agent-viz

An interactive terminal UI for visualizing GitHub Copilot coding agent sessions. Built as a `gh` CLI extension.

> **Tip:** Run `gh agent-viz --demo` to explore the UI with sample data.

## Features

- 📊 **Interactive TUI** - Browse agent sessions with keyboard navigation
- 📌 **Sessions at a Glance + Session Summary panel** - Fast status counts plus plain-language details/actions for the highlighted row
- 🔍 **Task Details** - View comprehensive task metadata (status, repo, branch, PR links)
- 📝 **Log Viewer** - Scrollable, searchable agent task logs
- 💻 **Local Sessions** - Automatically ingests local Copilot CLI sessions from `~/.copilot/session-state/`
- 🎨 **Status Indicators** - Color-coded status icons (running, queued, completed, failed)
- 🧑 **Input Needed Detection** - Highlights sessions that appear blocked waiting for human input
- 🚦 **Action Reasons** - Every card includes an explicit `Needs your action:` reason (`waiting on your input`, `run failed`, `running but quiet`, or `no action needed`)
- 🧭 **Action-First Ordering** - Sessions needing input/failure/quiet checks surface first, while older quiet duplicates are de-emphasized with a `↺ quiet duplicate` badge
- ⚡ **Quick Actions** - Contextual hints only show actions available for the highlighted session
- 🔄 **Resume Sessions** - Jump directly into active Copilot CLI sessions with one keystroke
- ⌨️ **Vim-style Keys** - j/k navigation, familiar keybindings
- 🛡️ **Tolerant Parsing** - Gracefully handles malformed session files without crashing
- 🎯 **Kanban board view** — `K` to toggle status-column layout
- 🔔 **Toast notifications** — status change alerts overlay
- 📊 **Session timeline** — Unicode lifecycle bar in detail view
- 🔗 **Dependency graph** — related session visualization
- 🎨 **Color themes** — catppuccin-mocha, dracula, tokyo-night presets
- 📡 **Live log tailing** — real-time log streaming with follow mode
- 💬 **Conversation view** — styled chat bubbles for session dialogue
- 🔧 **Tool timeline** — chronological trace of agent actions
- 📊 **Mission control** — fleet summary dashboard
- 🔍 **Diff view** — colored PR diffs in the TUI
- ❓ **Help overlay** — `?` shows all keybindings

See [docs/UI_FEATURES.md](docs/UI_FEATURES.md) for a comprehensive guide to all visual features.

## Installation

### Install via GitHub CLI

```bash
gh extension install maxbeizer/gh-agent-viz
```

### Build from Source

Requires Go 1.24.2 or newer.

```bash
git clone https://github.com/maxbeizer/gh-agent-viz.git
cd gh-agent-viz
go build -o bin/gh-agent-viz ./gh-agent-viz.go
```

### Install from Local Checkout (Development)

```bash
gh extension install .
```

This local install path uses the repository's executable wrapper and runs with your installed Go toolchain.

To reload while developing, just run `gh agent-viz` again after code changes. Local installs cannot be upgraded with `gh extension upgrade`.

If you need to (re)link the local checkout, run:

```bash
make relink-local
```

## Usage

If you want a fast walkthrough of what the board is showing, read [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md).

### Launch the TUI

```bash
gh agent-viz
```

### Scope to a Specific Repository

```bash
gh agent-viz --repo owner/repo
```

### Enable Debug Mode

```bash
gh agent-viz --debug
```

Debug mode writes command diagnostics to `~/.gh-agent-viz-debug.log` to speed up troubleshooting.
When enabled, the UI also shows a persistent debug banner with the log path.

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` / `j` / `k` | Navigate sessions |
| `enter` | View session details |
| `esc` | Go back |
| `l` | View logs |
| `o` | Open PR in browser |
| `s` | Resume session |
| `x` | Dismiss session |
| `r` | Refresh |
| `p` | Toggle preview pane |
| `K` | Toggle kanban view |
| `tab` / `shift+tab` | Cycle filter tabs |
| `a` | Jump to attention tab |
| `g` | Cycle group-by mode |
| `space` | Expand/collapse group |
| `c` | Conversation view |
| `t` | Tool timeline |
| `d` | Diff view (list/detail); page down (in logs) |
| `M` | Toggle mission control |
| `?` | Toggle help overlay |
| `!` | Open gh-agent-viz repo in browser |
| `@` | File a new issue in browser |
| `f` | Toggle follow mode (in logs) |
| `u` | Page up (in logs) |
| `g` / `G` | Top/bottom (in logs) |
| `q` | Quit |

Action hints in the footer are contextual: unavailable actions are hidden for the currently highlighted session.

### Resume Active Sessions

Press `s` on any active **local Copilot CLI session** (status: `running` or `queued`) to resume it directly in your terminal. This executes `gh copilot -- --resume <session-id>` and drops you into the Copilot CLI session.

**Note:** Only active local sessions can be resumed. Attempting to resume a remote agent-task row, or a completed/failed session, shows a clear error message.

### At-a-glance card semantics

Each session card is intentionally labeled for quick triage:

- `Repository:` shows linked repo context (`not available` when missing)
- `Needs your action:` explains why it needs action now (or confirms `no action needed`)
- Older repeated quiet sessions are intentionally de-emphasized (`↺ quiet duplicate`) to reduce list noise without hiding rows
- `Last update:` shows recency using friendly wording like `not recorded` when metadata is missing
- The **Session Summary** panel mirrors the same plain-language fields for the highlighted row

### Log Viewer Navigation

When viewing logs:

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down one line |
| `k` / `↑` | Scroll up one line |
| `d` | Scroll down half page |
| `u` | Scroll up half page |
| `g` | Go to top |
| `G` | Go to bottom |
| `f` | Toggle follow mode (live tailing) |
| `esc` | Return to task list |

## Configuration

Create a `.gh-agent-viz.yml` file in your home directory to customize settings:

```yaml
# List of repositories to watch
repos:
  - owner/repo1
  - owner/repo2

# Refresh interval in seconds (default: 30)
refreshInterval: 30

# Default status filter: all, attention, active, completed, failed (default: all)
defaultFilter: all

# Color theme: default, catppuccin-mocha, dracula, tokyo-night, solarized-light
theme: catppuccin-mocha
```

## Documentation

- **[Quick Docs Home](docs/index.md)** - Start here for product usage
- **[Getting Started](docs/GETTING_STARTED.md)** - Understand what you are seeing on screen
- **[Operator Guide](docs/OPERATOR_GUIDE.md)** - Daily workflows and keybindings
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Fast fixes for common issues
- **[Debug Mode](docs/DEBUG_MODE.md)** - Capture actionable diagnostics
- **[Developer Workflow](docs/DEVELOPER_WORKFLOW.md)** - Makefile commands for build/test/smoke
- **[UI Features](docs/UI_FEATURES.md)** - Kanban, toasts, timeline, dependency graph, themes, live tailing, conversation view, tool timeline, diff view, mission control, help overlay
- **[Architecture Decisions](docs/DECISIONS.md)** - Technical design rationale and patterns

## Requirements

- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
- GitHub Copilot CLI with agent-task commands available

## Architecture

This is a `gh` CLI extension (not a Copilot CLI plugin) because:

- **Copilot CLI plugins** operate within the agent conversation model - they provide skills, MCP servers, and custom agents, but offer **no control over terminal UI**
- **`gh` extensions** have full control over the terminal, enabling interactive TUI experiences

See [docs/DECISIONS.md](docs/DECISIONS.md) for detailed architecture decisions.

## Security

Security is a core requirement for this project. See [docs/SECURITY.md](docs/SECURITY.md) for security principles, threat surface, and required engineering practices.

## Documentation Site

- GitHub Pages is deployed from `docs/` via `.github/workflows/docs-pages.yml`.
- Site URL: https://maxbeizer.github.io/gh-agent-viz/
- If initial deploy cannot enable Pages automatically, enable GitHub Pages once in repository settings and rerun the Docs workflow.
- Keep docs in sync with shipped behavior when merging changes.

### Technology Stack

- **Language**: Go
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Data Source**: `gh agent-task` commands with `--json` output

### Data Sources

gh-agent-viz pulls sessions from two sources:

1. **Remote Agent Tasks**: Via `gh agent-task` CLI commands
2. **Local Copilot Sessions**: From `~/.copilot/session-state/*/workspace.yaml`

Both sources are displayed together in the unified session list. See [docs/LOCAL_SESSIONS.md](docs/LOCAL_SESSIONS.md) for details on local session ingestion.

### Project Structure

```
gh-agent-viz/
├── gh-agent-viz.go          # Entry point
├── cmd/                     # Cobra commands
├── internal/
│   ├── data/               # Data fetching (gh agent-task)
│   ├── config/             # Configuration parsing
│   └── tui/                # Bubble Tea UI components
│       └── components/     # Header, footer, task list, detail, logs
└── docs/                   # Architecture decisions
```

## Development

### Prerequisites

- Go 1.24.2 or later
- GitHub CLI authenticated

### One-command workflow (recommended)

```bash
make build
make test
make smoke
```

### Full local validation (CI-like)

```bash
make ci
```

### Additional developer commands

```bash
make test-race
make coverage
make fmt
make lint
make clean
```

See `make help` and [docs/DEVELOPER_WORKFLOW.md](docs/DEVELOPER_WORKFLOW.md).

## Reference

This project follows patterns from [gh-dash](https://github.com/dlvhdr/gh-dash), the gold standard for interactive Bubble Tea `gh` extensions.

## License

MIT
