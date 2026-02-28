# gh-agent-viz

An interactive terminal UI for visualizing GitHub Copilot coding agent sessions. Built as a `gh` CLI extension.

![gh-agent-viz demo](docs/demo.gif)

> [!TIP]
> Run `gh agent-viz --demo` to explore the UI with sample data.

## Features

- 🎯 **Dashboard-first** — Multi-pane btop-style landing page with Active, Recent, Attention, Fleet, Repos, and Idle panels
- 📊 **Stats bar** — Always-visible summary of active, idle, done, and token usage
- 🔍 **Fuzzy search** — Press `/` to filter sessions by title, repo, branch, or status
- 🖱️ **Mouse support** — Scroll and click to focus panels
- 🎯 **Kanban board** — `K` to toggle status-column layout with compact cards
- 📸 **Snapshot debugging** — Press `S` to capture TUI state as JSON, or `--snapshot <path>` on launch
- 📌 **Session detail** — Comprehensive metadata, timeline, telemetry, and dependency graph
- 📝 **Log viewer** — Scrollable agent task logs with live tailing
- 💬 **Conversation view** — Styled chat bubbles for session dialogue
- 🔧 **Tool timeline** — Chronological trace of agent tool calls
- 🔍 **Diff view** — Colored PR diffs in the TUI
- 💻 **Local sessions** — Automatically ingests local Copilot CLI sessions from `~/.copilot/session-state/`
- 🎨 **Color themes** — catppuccin-mocha, dracula, tokyo-night, solarized-light
- 🔔 **Toast notifications** — Status change alerts and action confirmations
- 🔄 **Resume sessions** — Jump directly into active Copilot CLI sessions with one keystroke
- ⌨️ **Vim-style keys** — j/k navigation, familiar keybindings
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

#### Dashboard (home)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate within focused panel |
| `tab` / `shift+tab` | Cycle panel focus (Active → Recent → Attention → Repos → Idle) |
| `enter` | Drill into session detail, or filter by repo |
| `K` | Switch to kanban view |
| `/` | Fuzzy search sessions |
| `S` | Save snapshot to `/tmp/` |
| `r` | Refresh data |
| `?` | Toggle help overlay |
| `q` | Quit |

#### Kanban

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Cycle columns (wraps around) |
| `j` / `k` | Navigate cards |
| `enter` | View session details |
| `X` | Dismiss all completed sessions |
| `esc` | Back to dashboard |

#### List View

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate sessions |
| `enter` | View session details |
| `l` | View logs |
| `o` | Open PR in browser |
| `s` | Resume session |
| `x` | Dismiss session |
| `X` | Dismiss all completed |
| `p` | Toggle preview pane |
| `g` | Cycle group-by mode |
| `d` | View PR diff |
| `esc` | Back to dashboard |

#### Detail / Logs

| Key | Action |
|-----|--------|
| `l` | View logs (from detail) |
| `c` | Conversation view (local sessions) |
| `t` | Tool timeline (local sessions) |
| `d` | View PR diff |
| `f` | Toggle follow mode (in logs) |
| `j` / `k` | Scroll |
| `esc` | Back to dashboard |

### Resume Active Sessions

Press `s` on any active **local Copilot CLI session** (status: `running` or `queued`) to resume it directly in your terminal. This executes `gh copilot -- --resume <session-id>` and drops you into the Copilot CLI session.

**Note:** Only active local sessions can be resumed. Attempting to resume a remote agent-task row, or a completed/failed session, shows a clear error message.

## Configuration

Create a `.gh-agent-viz.yml` file in your home directory to customize settings:

```yaml
# List of repositories to watch
repos:
  - owner/repo1
  - owner/repo2

# Refresh interval in seconds (default: 30)
refreshInterval: 30

# Default view on launch: dashboard, table, kanban (default: dashboard)
defaultView: dashboard

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
- GitHub Copilot CLI with agent-task commands available (used as fallback data source)

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
- **Data Source**: Copilot API (`api.githubcopilot.com`) with `gh agent-task` CLI fallback

### Data Sources

gh-agent-viz pulls sessions from two sources:

1. **Remote Agent Tasks**: Primarily via direct HTTP to the Copilot API (`api.githubcopilot.com`), with `gh agent-task` CLI as fallback
2. **Local Copilot Sessions**: From `~/.copilot/session-state/*/workspace.yaml`

Both sources are displayed together in the unified session list. See [docs/LOCAL_SESSIONS.md](docs/LOCAL_SESSIONS.md) for details on local session ingestion.

### Project Structure

```
gh-agent-viz/
├── gh-agent-viz.go          # Entry point
├── cmd/                     # Cobra commands
├── internal/
│   ├── data/               # Data fetching (CAPI direct + CLI fallback)
│   │   ├── capi/           # Copilot API client
│   │   └── snapshot.go     # Machine-readable TUI state capture
│   ├── config/             # Configuration parsing
│   └── tui/                # Bubble Tea UI components
│       └── components/     # Dashboard, kanban, stats bar, header, footer, etc.
└── docs/                   # Documentation and architecture decisions
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
