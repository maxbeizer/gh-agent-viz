# gh-agent-viz

An interactive terminal UI for visualizing GitHub Copilot coding agent sessions. Built as a `gh` CLI extension.

## Features

- 📊 **Interactive TUI** - Browse agent sessions with keyboard navigation
- 🔍 **Task Details** - View comprehensive task metadata (status, repo, branch, PR links)
- 📝 **Log Viewer** - Scrollable, searchable agent task logs
- 🎨 **Status Indicators** - Color-coded status icons (running, queued, completed, failed)
- ⚡ **Quick Actions** - Open PRs in browser, refresh data, filter by status
- ⌨️ **Vim-style Keys** - j/k navigation, familiar keybindings

## Installation

### Install via GitHub CLI

```bash
gh extension install maxbeizer/gh-agent-viz
```

### Build from Source

```bash
git clone https://github.com/maxbeizer/gh-agent-viz.git
cd gh-agent-viz
go build -o gh-agent-viz ./gh-agent-viz.go
```

## Usage

### Launch the TUI

```bash
gh agent-viz
```

### Scope to a Specific Repository

```bash
gh agent-viz --repo owner/repo
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `h` / `←` | Move to previous column |
| `→` | Move to next column |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `enter` | View task details |
| `l` | View task logs |
| `o` | Open PR in browser |
| `r` | Refresh task list |
| `tab` | Cycle status filter (all → active → completed → failed) |
| `esc` | Go back to task list |
| `q` | Quit |

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

# Default status filter: all, active, completed, failed (default: all)
defaultFilter: all
```

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

### Technology Stack

- **Language**: Go
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Data Source**: `gh agent-task` commands with `--json` output

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

- Go 1.21 or later
- GitHub CLI authenticated

### Build

```bash
go build -o gh-agent-viz ./gh-agent-viz.go
```

### Run

```bash
./gh-agent-viz
```

### Dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/spf13/cobra
go get github.com/cli/go-gh/v2
go get gopkg.in/yaml.v3
```

## Reference

This project follows patterns from [gh-dash](https://github.com/dlvhdr/gh-dash), the gold standard for interactive Bubble Tea `gh` extensions.

## License

MIT
