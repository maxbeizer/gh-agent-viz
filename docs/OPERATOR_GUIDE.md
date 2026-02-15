# Operator Guide

This guide helps you effectively supervise multiple GitHub Copilot coding agent workstreams using `gh-agent-viz`.

## Overview

`gh-agent-viz` is designed for developers and operators who need to monitor and manage multiple Copilot coding agent sessions across one or more repositories. It provides a command-center-style interface for:

- **Real-time monitoring** of agent session status across repos
- **Quick navigation** to task details and logs
- **Status filtering** to focus on active, completed, or failed sessions
- **Direct actions** like opening PRs or refreshing data

## Getting Started

### Single Repository Monitoring

Launch the TUI for a specific repository:

```bash
gh agent-viz --repo owner/repo-name
```

This displays all agent sessions for the specified repository.

### Multi-Repository Monitoring

To monitor multiple repositories, create a configuration file at `~/.gh-agent-viz.yml`:

```yaml
repos:
  - owner/frontend-app
  - owner/backend-api
  - owner/infrastructure
  
refreshInterval: 30
defaultFilter: all
```

Then launch without the `--repo` flag:

```bash
gh agent-viz
```

The TUI will show sessions from all configured repositories.

## Interface Overview

The TUI is organized into three main sections:

```
┌─────────────────────────────────────┐
│ Header: Title & Status Filter       │  <- Shows current filter (all/active/completed/failed)
├─────────────────────────────────────┤
│                                     │
│ Main View:                          │
│  - Task List (default)              │  <- Table of agent sessions
│  - Task Detail (on enter)           │  <- Full metadata for selected session
│  - Log Viewer (on 'l')              │  <- Scrollable agent logs
│                                     │
└─────────────────────────────────────┘
│ Footer: Keybinding Hints            │  <- Context-aware shortcuts
└─────────────────────────────────────┘
```

## Core Workflows

### Workflow 1: Monitoring Active Sessions

1. Launch `gh agent-viz` (shows all sessions by default)
2. Press `tab` to cycle to "Active" filter
3. Use `j`/`k` or arrow keys to navigate the list
4. Press `enter` on a session to view full details
5. Press `esc` to return to the list

**Tip:** Sessions marked as "running" (🔄) or "queued" (⏳) are currently active.

### Workflow 2: Investigating Failures

1. Press `tab` repeatedly until the filter shows "Failed"
2. Navigate to a failed session
3. Press `l` to view logs
4. Scroll through logs using:
   - `j`/`k` - line by line
   - `d`/`u` - half page
   - `g`/`G` - top/bottom
5. Press `esc` to go back

**Tip:** Look for error messages, tool failures, or timeout indicators in the logs.

### Workflow 3: Reviewing Completed Work

1. Press `tab` to filter to "Completed" sessions
2. Navigate to a session
3. Press `o` to open the associated PR in your browser
4. Review the changes made by the agent

**Tip:** Completed sessions (✓) represent finished work that may need human review.

### Workflow 4: Refreshing Data

Agent session data is refreshed automatically based on your `refreshInterval` (default: 30 seconds). To manually refresh:

1. Press `r` from the task list view
2. The list updates with the latest session data

**Tip:** Manual refresh is useful when you know a session status has changed.

## Keyboard Shortcuts Reference

### Task List View

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Move down | Navigate to next session |
| `k` / `↑` | Move up | Navigate to previous session |
| `enter` | View details | Show full metadata for selected session |
| `l` | View logs | Open log viewer for selected session |
| `o` | Open PR | Open associated PR in browser |
| `r` | Refresh | Manually refresh task list |
| `tab` | Toggle filter | Cycle through status filters |
| `q` / `ctrl+c` | Quit | Exit the application |

### Task Detail View

| Key | Action | Description |
|-----|--------|-------------|
| `esc` | Back | Return to task list |
| `l` | View logs | Switch to log viewer |
| `o` | Open PR | Open associated PR in browser |
| `q` / `ctrl+c` | Quit | Exit the application |

### Log Viewer

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Scroll down | Move down one line |
| `k` / `↑` | Scroll up | Move up one line |
| `d` | Page down | Scroll down half page |
| `u` | Page up | Scroll up half page |
| `g` | Top | Jump to beginning of logs |
| `G` | Bottom | Jump to end of logs |
| `esc` | Back | Return to task list |
| `q` / `ctrl+c` | Quit | Exit the application |

## Status Indicators

Sessions are marked with color-coded status icons:

- **🔄 Running** (blue): Agent is actively working
- **⏳ Queued** (yellow): Session is waiting to start
- **✓ Completed** (green): Work finished successfully
- **✗ Failed** (red): Session encountered an error

## Configuration Best Practices

### Multi-Team Monitoring

For teams managing multiple projects, organize repos by priority:

```yaml
repos:
  # Critical production repos first
  - company/api-gateway
  - company/payment-service
  
  # Feature repos second
  - company/new-dashboard
  - company/mobile-app
  
refreshInterval: 20  # Faster refresh for production monitoring
defaultFilter: active  # Focus on in-progress work
```

### Individual Developer Setup

For personal use, list your active work repos:

```yaml
repos:
  - myuser/current-project
  
refreshInterval: 60  # Slower refresh to reduce API calls
defaultFilter: all  # See everything at a glance
```

### CI/CD Integration

When using agent sessions in CI/CD pipelines, monitor failures:

```yaml
repos:
  - org/repo1
  - org/repo2
  
defaultFilter: failed  # Immediately see what needs attention
```

## Advanced Usage

### Filtering by Session Age

While there's no built-in age filter, you can use the status workflow:

1. Filter to "Completed" to see recently finished sessions
2. Filter to "Active" to see current work
3. Old sessions naturally move out of "Active" status

### Monitoring Specific Branches

The task detail view (press `enter`) shows the branch for each session. To focus on a specific branch:

1. Navigate through sessions in list view
2. Note the "Repository" column (includes repo name)
3. Press `enter` to see full branch information

**Note:** Branch-level filtering is not currently supported but is planned for future releases.

### Working with PR Links

When a session has an associated PR:

1. The "PR" column shows the PR number
2. Press `o` to open the PR in your default browser
3. Review code changes, add comments, or approve

## Performance Tips

### Large Repository Lists

If monitoring many repos slows down the interface:

1. Reduce the `refreshInterval` to 60+ seconds
2. Scope to specific repos using `--repo` flag instead of config file
3. Use status filters to reduce visible rows

### Network Latency

If refreshes are slow due to network issues:

1. Increase `refreshInterval` to reduce API calls
2. Use `r` for manual refresh only when needed
3. Consider running `gh agent-viz` closer to your work sessions (e.g., on a jump box)

## Integration with GitHub CLI

`gh-agent-viz` leverages your existing `gh` authentication:

- No separate login required
- Respects your `gh` configuration
- Uses the same credentials as other `gh` commands

To verify your authentication:

```bash
gh auth status
```

## Best Practices Summary

1. **Use status filters** - Don't scroll through all sessions; filter to what matters
2. **Configure refresh interval** - Balance freshness with performance
3. **Organize repos by priority** - Put critical repos first in your config
4. **Use keyboard shortcuts** - Vim-style keys make navigation fast
5. **Check logs on failures** - Always investigate failed sessions promptly
6. **Open PRs for review** - Use `o` to quickly jump to GitHub for code review

## Next Steps

- See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for help with common issues
- Check [DECISIONS.md](DECISIONS.md) for architectural background
- Review the main [README.md](../README.md) for installation and basic usage
