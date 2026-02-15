# Local Copilot CLI Session Ingestion

This document describes the local Copilot CLI session ingestion feature added to gh-agent-viz.

## Overview

gh-agent-viz now supports ingesting local Copilot CLI sessions from `~/.copilot/session-state/*/workspace.yaml` in addition to remote agent-task sessions via the `gh agent-task` CLI.

## Architecture

### Unified Session Model

A new `Session` type unifies both agent-task and local Copilot sessions:

```go
type Session struct {
    ID         string
    Status     string
    Title      string
    Repository string
    Branch     string
    PRURL      string
    PRNumber   int
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Source     SessionSource  // "agent-task" or "local-copilot"
}
```

### Session Sources

Two session sources are supported via the `SessionSource` enum:
- `SourceAgentTask`: GitHub agent-task sessions from `gh agent-task` CLI
- `SourceLocalCopilot`: Local Copilot CLI sessions from `~/.copilot/session-state/`

### Tolerant Parsing

The local session parser implements tolerant parsing with fallback behavior:

1. **Primary Parse**: Attempt to parse YAML using standard unmarshaling
2. **Fallback Parse**: If YAML is malformed, extract key fields line-by-line
3. **Silent Continue**: Skip sessions that cannot be parsed at all

This ensures the TUI never crashes due to malformed session files.

## Status Mapping

Local session status is derived using `DeriveLocalSessionStatus()`:

### Explicit Status Mapping

| Input Statuses | Normalized Output |
|---------------|------------------|
| completed, finished, done, merged, closed | completed |
| running, in progress, active, open | running |
| failed, error, cancelled, canceled | failed |
| queued, pending, waiting | queued |

### Time-Based Derivation

When no explicit status is provided or status is unknown:
- **Last activity > 24 hours ago**: Status = `completed`
- **Last activity recent**: Status = `running`
- **No timestamp**: Status = `unknown`

## API

### FetchLocalSessions()

```go
func FetchLocalSessions() ([]Session, error)
```

Fetches all local Copilot CLI sessions from `~/.copilot/session-state/`. Returns an empty list (not an error) if the directory doesn't exist.

### FetchAllSessions(repo string)

```go
func FetchAllSessions(repo string) ([]Session, error)
```

Fetches both agent-task and local sessions. Filters by repository if specified. This is the recommended API for fetching all available sessions.

### Session Conversion

```go
func FromAgentTask(task AgentTask) Session
func (s Session) ToAgentTask() AgentTask
```

Conversion functions maintain backward compatibility with existing code that expects `AgentTask` objects.

## Testing

All functionality is covered by tests:

- **Status Mapping**: Tests for all explicit status values and time-based derivation
- **Valid YAML**: Tests successful parsing of well-formed workspace.yaml files
- **Malformed YAML**: Tests fallback parsing for broken YAML
- **Empty Cases**: Tests handling of missing directories, empty sessions, missing fields
- **Integration**: Manual integration test verifies end-to-end functionality

Run tests with:
```bash
go test ./internal/data/...
```

## File Format

Expected `workspace.yaml` structure (all fields optional except `session_id`):

```yaml
session_id: "abc123-session"
start_time: "2026-02-15T03:10:00Z"
last_activity: "2026-02-15T03:30:00Z"
message_count: 15
status: "completed"
repository: "owner/repo"
branch: "main"
title: "Session title"
conversation_history:
  - role: user
    content: "First message"
```

## UI Changes

- Task list now shows both agent-task and local sessions
- Detail view displays session source (agent-task or local-copilot)
- PR-related fields (PR URL, PR number) only shown for agent-task sessions
- Local sessions display message when trying to open in browser ("local sessions don't have associated pull requests")

## Backward Compatibility

All existing agent-task functionality remains unchanged:
- `FetchAgentTasks()` still works as before
- `FetchAgentTaskDetail()` still works as before
- `FetchAgentTaskLog()` still works as before
- `AgentTask` type still exists for backward compatibility

## Known Limitations

1. **Detail View**: Currently only agent-task sessions support the detail view. Local sessions show basic info from the list view only.
2. **Logs**: Local sessions don't support log viewing (no equivalent to `gh agent-task view --log`)
3. **Browser Open**: Cannot open local sessions in browser (no associated PR)

## Future Enhancements

Possible future improvements:
- Enhanced detail view for local sessions showing conversation history
- Support for viewing local session events/logs
- Real-time watching of local session files for live updates
- Better title extraction from conversation history
- Repository auto-detection from git context if not in YAML
