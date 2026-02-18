# Copilot Instructions for gh-agent-viz

This document provides GitHub Copilot with full context about the gh-agent-viz project to assist effectively in any editor or chat context.

## Project Overview

- `gh-agent-viz` is a GitHub CLI (`gh`) extension that provides an interactive terminal UI (TUI) for visualizing GitHub Copilot coding agent sessions
- Users install it via `gh extension install maxbeizer/gh-agent-viz` and run it with `gh agent-viz`
- It is **not** a Copilot CLI plugin — it is a standalone `gh` extension with full terminal UI control

## Tech Stack

- **Language**: Go
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm architecture — Model/Update/View)
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Pre-built TUI components**: [Bubbles](https://github.com/charmbracelet/bubbles) (table, viewport, spinner, etc.)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **GitHub Integration**: [go-gh v2](https://github.com/cli/go-gh) — provides auth context from `gh auth`
- **Build/Release**: GoReleaser for cross-platform binaries

## Security-First Requirements (Critical)

Security is a primary product goal, not an afterthought.

- Never hardcode, print, or commit secrets, access tokens, cookies, or credentials.
- Treat all CLI output and local files as untrusted input; parse defensively and fail safely.
- Avoid shell command injection risks: use explicit argument arrays (`exec.Command`) instead of shell interpolation.
- Do not silently ignore security-relevant errors; surface actionable failures to users.
- Keep data minimization by default: collect only what is needed for UX and troubleshooting.
- Any new telemetry or analytics capability must document privacy implications and opt-in behavior.
- Prefer least-privilege behavior for file reads, command execution, and external calls.
- Do not introduce network exfiltration paths for local session data without explicit user intent and documentation.

## Architecture

Modeled on [`dlvhdr/gh-dash`](https://github.com/dlvhdr/gh-dash), which is the reference implementation for interactive Bubble Tea `gh` extensions (10k+ stars).

## Project Structure

```
gh-agent-viz.go              # Entry point, calls cmd.Execute()
cmd/
  root.go                    # Cobra root command, initializes Bubble Tea program
internal/
  config/
    config.go                # YAML config parser (.gh-agent-viz.yml)
  data/
    agentapi.go              # Data fetching — shells out to `gh agent-task` commands
  tui/
    ui.go                    # Top-level Bubble Tea Model (Init/Update/View)
    keys.go                  # Centralized key bindings
    context.go               # Shared ProgramContext passed to all components
    theme.go                 # Lip Gloss styles and colors
    components/
      tasklist/              # Main table view of agent sessions
      taskdetail/            # Detail pane for selected session
      logview/               # Scrollable log viewer
      header/                # Header bar with title and filter tabs
      footer/                # Footer with keybinding hints
docs/
  DECISIONS.md               # Architecture decision log
```

## Data Source

- The data layer shells out to `gh agent-task` CLI commands:
  - `gh agent-task list` — lists recent Copilot coding agent sessions
  - `gh agent-task view <id>` — shows detail for a specific session
  - `gh agent-task view <id> --log` — shows the event log
  - `gh agent-task view <id> --log --follow` — streams live logs
- **Important**: The `--json` flag support and exact output schema for these commands needs verification. The data structs in `internal/data/agentapi.go` are best-guess and may need adjustment based on actual CLI output.
- Auth is handled automatically by `go-gh` picking up the user's `gh auth` token

## Coding Conventions

- Follow standard Go project layout (`cmd/`, `internal/`)
- Each TUI component lives in its own package under `internal/tui/components/`
- Components receive a shared `*context.ProgramContext` for access to config, dimensions, styles, and error state
- Key bindings are defined centrally in `internal/tui/keys.go`
- Styles and colors are defined in `internal/tui/theme.go` using Lip Gloss
- The Bubble Tea pattern: every component implements or contributes to `Init() tea.Cmd`, `Update(msg tea.Msg) (tea.Model, tea.Cmd)`, `View() string`
- Use `exec.Command("gh", ...)` for shelling out to `gh agent-task` commands
- Error handling: return errors up, display them in the TUI rather than crashing

## Key Constraints

- The Copilot CLI plugin system was evaluated and ruled out — it provides no terminal UI control (only skills, MCP servers, hooks, and custom agents within the conversation model)
- There is no dedicated REST API for agent sessions yet — data comes from `gh agent-task` CLI commands
- If a REST API appears in the future, switch to using `go-gh` REST client (`gh.DefaultRESTClient()`) directly
- Changes that affect data handling, telemetry, or external integrations must update `docs/SECURITY.md`

## Shared Helpers (internal/data/session.go)

When working with session data across components, use these shared functions instead of reimplementing:
- `data.SessionNeedsAttention(s)` — true for needs-input/failed only
- `data.SessionIsActiveNotIdle(s)` — true for active sessions updated within 20min
- `data.StatusIsActive(status)` — true for running/queued/active/open/in-progress
- `data.IsDefaultBranch(branch)` — true for main/master/empty
- `data.FormatTokenCount(n)` — "2.7M", "11.7K", "437"

## Testing

- When writing tests, use Go's standard `testing` package
- For TUI components, test the Model's Update function with specific messages and verify state changes
- For the data layer, mock `exec.Command` output to test parsing without requiring `gh` to be installed
- Add regression tests for parser hardening, malformed input handling, and error propagation paths
