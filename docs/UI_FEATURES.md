# UI Features

Comprehensive guide to the visual features in `gh-agent-viz`.

## Kanban Board View

Press `K` to toggle the kanban board layout. The board organizes sessions into four status columns:

| Column | Statuses |
|--------|----------|
| **Running** | Active and queued sessions |
| **Needs Input** | Sessions waiting for human action |
| **Completed** | Successfully finished sessions |
| **Failed** | Sessions that errored out |

### Navigation

- `h` / `l` — move between columns
- `j` / `k` — move within a column
- `enter` — open session details

### When to use it

The kanban view is most useful when you are monitoring many concurrent agent sessions across repositories. It gives you a fast visual overview of where sessions are in their lifecycle without scrolling through a flat list.

Press `K` again to return to the default list view.

## Toast Notifications

Toast notifications appear as temporary overlays in the top-right corner of the terminal when session statuses change between refreshes.

### Behavior

- **Trigger**: A session changes status (e.g., running → completed, running → needs input) between automatic or manual refreshes.
- **Auto-dismiss**: Each toast disappears after **5 seconds**.
- **Stacking**: Up to **3 toasts** are visible at once. Newer toasts push older ones down.
- **Initial load**: No toasts fire on the first data load — they only appear for status *changes*.

Toasts help you notice important transitions (especially failures or input-needed events) without having to actively watch the session list.

## Session Timeline Bar

The session timeline bar is a compact Unicode visualization of a session's lifecycle. It appears in the detail view (both full-screen and split-pane modes).

### How to read it

The bar spans from the session's **created** timestamp to **now**:

| Character | Meaning |
|-----------|---------|
| `░` | Idle — session exists but is not actively running |
| `▓` | Completed active — work was happening during this interval |
| `█` | Running active — session is currently executing |

Example:

```
░░░░▓▓▓▓▓▓████
^              ^
created        now
```

This session was idle briefly, then ran actively, and is currently still running.

### Time range

- Start: session `created_at` timestamp
- End: current time (`now`)
- The bar width scales to fit the available terminal width.

## Dependency Graph

The dependency graph visualizes relationships between sessions. It appears in the detail view when related sessions are detected.

### How relationships are detected

Sessions are considered related when they share:

- **Same repository** — sessions targeting the same `owner/repo`
- **Branch prefixes** — branches sharing a common prefix (e.g., `feature/auth-login` and `feature/auth-signup`)

### Notation

Relationships are rendered using box-drawing characters:

```
┌─ Add login endpoint (running)
├─ Add signup flow (completed)
└─ Fix auth tests (needs input)
```

Parent-child or sibling relationships are inferred from branch naming and repository context. The graph is informational — it does not imply execution order or blocking dependencies.

## Color Themes

gh-agent-viz supports multiple color themes to match your terminal aesthetic.

### Available themes

| Theme | Description |
|-------|-------------|
| `default` | Adaptive theme that auto-detects light/dark terminal background |
| `catppuccin-mocha` | Warm pastel theme from the Catppuccin palette |
| `dracula` | Dark theme with vibrant accents |
| `tokyo-night` | Cool-toned dark theme inspired by Tokyo Night |
| `solarized-light` | Light theme using the Solarized Light palette |

### Configuration

Set the theme in `~/.gh-agent-viz.yml`:

```yaml
theme: catppuccin-mocha
```

### Adaptive default

When no theme is specified (or `theme: default`), gh-agent-viz queries your terminal's background color and selects appropriate contrast levels automatically. This works in most modern terminals (iTerm2, Ghostty, Kitty, Windows Terminal, etc.).

## Live Log Tailing

Live log tailing streams agent session logs in real time, similar to `tail -f`.

### How to activate

1. Highlight a **running** session in the session list.
2. Press `l` to open the log viewer.
3. Press `f` to enable follow mode.

### Follow mode

When follow mode is active:

- A **LIVE 🔴** indicator appears in the log viewer header.
- New log lines are fetched every **2 seconds** via `gh agent-task view <id> --log --follow`.
- The viewport automatically scrolls to the bottom as new lines arrive.
- Press `f` again to disable follow mode and browse the log freely.

### Log viewer keys (while viewing logs)

| Key | Action |
|-----|--------|
| `f` | Toggle follow mode |
| `j` / `↓` | Scroll down one line |
| `k` / `↑` | Scroll up one line |
| `d` | Page down |
| `u` | Page up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `esc` | Return to session list |

## Filter Tabs

The header shows filter tabs that organize sessions by status.

### Tab behavior

| Tab | Shows | When it matters |
|-----|-------|-----------------|
| **ATTENTION** | `needs-input` and `failed` sessions only | Something is waiting on **you** |
| **RUNNING** | Active sessions (running, queued) | Sessions currently working |
| **DONE** | Completed sessions | Finished work |
| **FAILED** | Failed sessions | Errors to investigate |
| **ALL** | Everything | Full overview |

### Smart default tab

On startup, the UI picks the most useful tab automatically:
1. **ATTENTION** — if there are sessions waiting on you
2. **RUNNING** — if there are active sessions
3. **ALL** — fallback when nothing is active

### What "needs attention" means

Only two statuses trigger the ATTENTION tab:
- **`needs-input`** — the agent has explicitly asked a question and is blocked waiting for your answer
- **`failed`** — the agent hit an error and stopped

Idle running sessions (e.g., an agent that finished responding and is waiting for your next message) show up under RUNNING with a `💤 idle` badge. This is a known limitation — the data source doesn't distinguish "agent actively working" from "agent waiting for the user to continue." See [#121](https://github.com/maxbeizer/gh-agent-viz/issues/121) for discussion.

Follow mode is only available for sessions with status `running`. For completed or failed sessions, the log viewer shows the full static log.

## Conversation View

Press `c` to open the conversation view from the session list, detail view, or log view. This renders the session's dialogue as styled chat bubbles.

### Layout

- **User messages** are left-aligned.
- **Agent messages** are right-aligned.
- **Tool executions** are shown inline between messages.

### Requirements

Conversation view only works for **local-copilot sessions** that have event logs on disk (`~/.copilot/session-state/`). Remote agent-task sessions do not expose conversation-level data.

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `d` | Page down |
| `u` | Page up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `esc` | Return to previous view |

## Tool Timeline

Press `t` to open the tool timeline from the session list or detail view. This shows a chronological trace of every tool execution in the session.

### Icons

| Icon | Tool type |
|------|-----------|
| 🔧 | bash |
| ✏️ | edit |
| 📄 | view |
| 🔍 | search |
| 📤 | git |
| 🧪 | test |
| ⚙️ | other |

### Requirements

Tool timeline is only available for **local-copilot sessions** with event logs.

## Diff View

Press `d` to open the PR diff from the session list or detail view. The diff is rendered with syntax-aware coloring directly in the TUI.

### Color coding

- **Green** — added lines
- **Red** — deleted lines
- **Cyan** — hunk headers (`@@` lines)

### PR discovery

For local sessions, diff view discovers the associated PR by looking up the session's branch name. This works even for **merged PRs**. While the diff is loading, the UI shows `🔄 Loading diff...`.

## Mission Control

Press `M` to toggle the mission control dashboard. This provides a high-level fleet overview across all monitored repositories.

### Sections

- **Fleet summary** — aggregate session counts (active, idle, done, failed) with a proportional bar chart.
- **Per-repo breakdown** — active/idle/done/failed counts for each repository.
- **Needs your attention** — surfaces sessions with `needs-input` or `failed` status for quick triage.

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move to next repo |
| `k` / `↑` | Move to previous repo |
| `M` | Return to session list |

## Help Overlay

Press `?` to toggle a full keybinding reference overlay. The overlay works in all view modes.

### Sections

The help overlay organizes shortcuts by category:

- **Navigation** — movement and selection
- **Actions** — open PR, resume session, dismiss, refresh
- **Views** — logs, conversation, diff, kanban, mission control
- **Groups** — group-by mode, expand/collapse
- **Log View** — follow mode, scrolling
- **Meta** — repo link, file issue, quit

## Meta Shortcuts

| Key | Action |
|-----|--------|
| `!` | Open the gh-agent-viz repository in your browser |
| `@` | File a new issue against gh-agent-viz in your browser |

## PR Integration for Local Sessions

Local Copilot CLI sessions on feature branches can automatically discover their associated pull request.

### How it works

- Press `o` to open the PR in your browser. The PR is discovered by looking up the session's branch name via the GitHub API.
- Sessions on feature branches show a `PR` tag in the meta line of the session card.
- **Main/master branches are skipped** — PR discovery only runs for feature branches.
- Works for both **open and merged** pull requests.
