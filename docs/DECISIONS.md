# Architecture Decisions

This document captures the key decisions made when building `gh-agent-viz`.

## Why a `gh` Extension and Not a Copilot CLI Plugin

### Copilot CLI Plugin System

The GitHub Copilot CLI (`github/copilot-cli`) has a plugin system that can be accessed via `/plugin install`. Plugins can provide several capabilities:

- **Custom agents** via `.agent.md` files
- **Skills** via `SKILL.md` files  
- **MCP servers** via `.mcp.json` configuration
- **Hooks** for session lifecycle events (e.g., `preToolUse`, `agentStop`)

The official plugin repository is `github/copilot-plugins`, with examples like:
- `workiq` - Combines MCP server and skill functionality
- `spark` - Provides skills only

### The Limitation

**Plugins operate within the agent conversation model.** They give the agent new knowledge and tools, but provide **zero control over the terminal UI**. There is no plugin API for:
- Rendering custom views
- Drawing tables or interactive elements
- Creating TUI components
- Handling keyboard input for navigation

A plugin could answer "what are my agent sessions?" conversationally, but it **cannot render an interactive dashboard**.

### The Solution

For interactive visualization, a **`gh` CLI extension** is the correct approach because:
- Full control over the terminal and UI rendering
- Can use frameworks like Bubble Tea for rich TUI experiences
- Follows established patterns from successful extensions like `gh-dash`
- Gets free authentication via the `gh` auth system
- Easy distribution through `gh extension install`

## Architecture Decisions

### Language: Go

**Decision:** Use Go as the primary language.

**Rationale:**
- Matches the `gh` CLI ecosystem
- Proven track record with `gh-dash` (10k+ stars)
- Access to the excellent Charmbracelet Bubble Tea framework
- Strong standard library and tooling
- Easy cross-platform compilation

### TUI Framework: Bubble Tea + Lip Gloss + Bubbles

**Decision:** Use the Charmbracelet stack for the TUI.

**Rationale:**
- **Bubble Tea**: Elm architecture for terminal UIs - clean, predictable state management
- **Lip Gloss**: Terminal styling library for colors, borders, and layouts
- **Bubbles**: Pre-built components (tables, viewports, spinners)
- Same stack used by `gh-dash`, proven to work well for `gh` extensions
- Active maintenance and excellent documentation

### CLI Framework: Cobra

**Decision:** Use Cobra for command-line argument parsing.

**Rationale:**
- Standard in the Go ecosystem
- Used by `gh` itself and `gh-dash`
- Powerful flag and subcommand support
- Auto-generated help documentation

### Data Source: Direct Copilot API with CLI Fallback

**Decision:** Fetch data primarily via direct HTTP calls to the Copilot API (`api.githubcopilot.com`), falling back to `gh agent-task` CLI commands if CAPI auth fails.

**Primary — Copilot API (CAPI):**
- Direct HTTP to `api.githubcopilot.com` using the user's `gh auth` OAuth token (Bearer `gho_` prefix)
- Required headers: `Copilot-Integration-Id: copilot-4-cli`, `X-GitHub-Api-Version: 2026-01-09`
- Implemented in `internal/data/capi/` (client.go, types.go, sessions.go)

**Fallback — CLI:**
- `gh agent-task list` - Lists recent agent sessions with status, repo, and timestamps
- `gh agent-task view <id> --log` - Shows event log for a session
- `gh agent-task view <id> --log --follow` - Streams live logs

**Rationale:**
- Direct CAPI provides structured JSON with all API fields available, faster responses, and no text parsing
- CLI fallback ensures the tool still works if the user's token lacks CAPI scopes
- Authentication is free via `go-gh` library (picks up user's `gh auth` token)

### Distribution: `gh extension install`

**Decision:** Distribute as a `gh` extension.

**Installation:**
```bash
gh extension install maxbeizer/gh-agent-viz
```

**Usage:**
```bash
gh agent-viz
gh agent-viz --repo owner/repo
```

**Rationale:**
- Native integration with GitHub CLI
- Familiar installation pattern for `gh` users
- Automatic updates via `gh extension upgrade`
- No separate authentication setup needed

### Reference Architecture: `gh-dash`

**Decision:** Model architecture on `dlvhdr/gh-dash`.

**Rationale:**
- Gold standard for Bubble Tea `gh` extensions (10k+ stars)
- MIT licensed, can study implementation patterns
- Proven component organization:
  - Separate packages for each UI component
  - Shared context struct
  - Centralized key bindings
  - Theme/styling in dedicated package

## Key Go Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework (Elm architecture) |
| `github.com/charmbracelet/lipgloss` | Terminal styling and layout |
| `github.com/charmbracelet/bubbles` | Pre-built UI components |
| `github.com/spf13/cobra` | CLI argument parsing |
| `github.com/cli/go-gh/v2` | GitHub auth context and API client |
| `gopkg.in/yaml.v3` | Configuration file parsing |

## What Was Ruled Out

### Copilot CLI Plugin

**Why not:** No API for custom UI rendering. Plugins can only provide conversational responses within the agent's chat interface.

**When it makes sense:** If we wanted to add a conversational interface to query agent sessions ("show me my last 5 agent runs"), a plugin could complement the TUI.

### Standalone Binary Outside `gh` Ecosystem

**Why not:** Would require separate authentication setup and wouldn't integrate with the `gh` CLI workflow.

**When it makes sense:** If we needed to support users who don't use `gh` CLI at all.

### Direct REST API Calls

**Why not (standalone):** The Copilot API is now the primary data source (see above), but a CLI fallback is retained for auth-edge cases.

**When it makes sense:** Future enhancement when GitHub provides a proper API endpoint.

## Future Considerations

### When a REST API Becomes Available

The primary data path now uses the Copilot API (`api.githubcopilot.com`) directly. If GitHub introduces a separate dedicated REST API endpoint for agent sessions, we should evaluate migrating to it for broader compatibility.

### Live Log Streaming

The `--follow` flag on `gh agent-task view <id> --log` enables live log streaming. This could be implemented using:
- A goroutine pumping lines into a Bubble Tea channel
- Real-time UI updates as logs arrive
- Useful for monitoring active agent sessions

### Companion Copilot CLI Plugin

We could create a plugin that provides:
- MCP tool wrapper around the same data layer
- Conversational queries alongside the TUI
- Example: "What agent sessions are currently running?"
- Would use the same underlying CAPI client or `gh agent-task` commands

### Configuration Enhancements

Future config file options could include:
- Custom status colors and emoji icons
- Keybinding customization
- Default sort order (by updated, created, status)
- Auto-refresh behavior
- Repository watchlist with priority indicators

## Project Structure

```
gh-agent-viz/
├── gh-agent-viz.go          # Entry point
├── cmd/
│   └── root.go              # Cobra root command
├── internal/
│   ├── data/
│   │   ├── agentapi.go      # Data fetching layer (orchestrates CAPI + CLI fallback)
│   │   └── capi/            # Direct Copilot API client
│   ├── config/
│   │   └── config.go        # Configuration parsing
│   └── tui/
│       ├── ui.go            # Main Bubble Tea model
│       ├── context.go       # Shared program context
│       ├── theme.go         # Lip Gloss styles
│       ├── keys.go          # Key bindings
│       └── components/
│           ├── header/
│           ├── footer/
│           ├── tasklist/
│           ├── taskdetail/
│           └── logview/
├── docs/
│   └── DECISIONS.md         # This file
├── .goreleaser.yaml         # Cross-platform builds
├── .gitignore               # Go ignores
└── README.md                # Install and usage docs
```

## Key Patterns from gh-dash

### Data Fetching via Copilot API (Primary) with CLI Fallback

```go
// Primary: direct CAPI call (internal/data/capi/)
client := capi.NewClient(token)
sessions, err := client.ListSessions(ctx)

// Fallback: shell out to gh agent-task
cmd := exec.Command("gh", "agent-task", "list", "--json")
output, err := cmd.Output()
```

### Bubble Tea Model Pattern

```go
type Model struct {
    // State fields
}

func (m Model) Init() tea.Cmd { /* ... */ }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* ... */ }
func (m Model) View() string { /* ... */ }
```

### Component Organization

Each UI component lives in its own package with:
- Model struct for state
- Constructor function
- Update/View methods
- Shared context passed from parent

### Centralized Key Bindings

Define all key bindings in one place for:
- Consistency across the application
- Easy customization
- Help text generation

### Theme-Based Styling

Lip Gloss styles defined in a theme package:
- Status colors (running, completed, failed)
- Table styling (header, rows, selection)
- Borders and layout
- Consistent visual design
