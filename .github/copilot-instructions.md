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
    agentapi.go              # Data fetching — orchestrates CAPI + CLI fallback
    capi/                    # Direct Copilot API client (client.go, types.go, sessions.go)
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
      kanban/                # Kanban board view (In Progress/Idle/Done)
      mission/               # Mission control summary dashboard
      conversation/          # Styled chat bubble view
      tooltimeline/          # Tool execution timeline
      diffview/              # PR diff viewer
      toast/                 # Transient notification toasts
      help/                  # Keybinding help overlay
docs/
  DECISIONS.md               # Architecture decision log
```

## Data Source

- The data layer fetches agent sessions via two strategies (with automatic fallback):
  1. **Primary — Copilot API (CAPI)**: Direct HTTP calls to `api.githubcopilot.com` using the user's `gh auth` OAuth token (Bearer `gho_` prefix). Implemented in `internal/data/capi/`. Required headers: `Copilot-Integration-Id: copilot-4-cli`, `X-GitHub-Api-Version: 2026-01-09`. Returns structured JSON with all API fields available.
  2. **Fallback — CLI**: Shells out to `gh agent-task` commands (`list`, `view <id>`, `view <id> --log`, `view <id> --log --follow`) if CAPI auth fails.
- Token usage is parsed from local Copilot CLI logs in `~/.copilot/logs/` to enrich sessions with model and token count data
- Auth is handled automatically by `go-gh` picking up the user's `gh auth` token

## Coding Conventions

- Follow standard Go project layout (`cmd/`, `internal/`)
- Each TUI component lives in its own package under `internal/tui/components/`
- Components receive a shared `*context.ProgramContext` for access to config, dimensions, styles, and error state
- Key bindings are defined centrally in `internal/tui/keys.go`
- Styles and colors are defined in `internal/tui/theme.go` using Lip Gloss
- The Bubble Tea pattern: every component implements or contributes to `Init() tea.Cmd`, `Update(msg tea.Msg) (tea.Model, tea.Cmd)`, `View() string`
- Prefer direct CAPI calls (`internal/data/capi/`) for data fetching; use `exec.Command("gh", ...)` only in the CLI fallback path
- Error handling: return errors up, display them in the TUI rather than crashing

## Key Constraints

- The Copilot CLI plugin system was evaluated and ruled out — it provides no terminal UI control (only skills, MCP servers, hooks, and custom agents within the conversation model)
- The primary data path uses the Copilot API (`api.githubcopilot.com`) directly; the `gh agent-task` CLI is retained as a fallback
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
- For the data layer, mock CAPI HTTP responses or `exec.Command` output to test parsing without requiring `gh` to be installed
- Add regression tests for parser hardening, malformed input handling, and error propagation paths

## Testing Guidelines

- Tests must be deterministic — no flaky tests
- Avoid `time.Now()` or `time.Since()` in test assertions (use fixed timestamps)
- Avoid filesystem access in unit tests (use temp dirs or mock data)
- Pure function tests are preferred (input → output, no side effects)
- Use `testing.T.TempDir()` when filesystem access is unavoidable
- Mock `exec.Command` via the `execCommand` variable for CLI tests
- Don't test Bubble Tea Update/View wiring directly — test component logic instead
