---
title: gh-agent-viz Docs
---

# gh-agent-viz

Your Copilot agent **mission control** in the terminal.

## Start here

1. **Install**
   ```bash
   gh extension install maxbeizer/gh-agent-viz
   ```
2. **Launch**
   ```bash
   gh agent-viz
   ```
3. **Explore with demo data**
   ```bash
   gh agent-viz --demo
   ```
4. **Key shortcuts** — press `?` for the full keybinding reference

## Guides

- [Getting Started](GETTING_STARTED.md) — what the board means and how to read it
- [UI Features](UI_FEATURES.md) — all views, themes, and visualizations
- [Operator Guide](OPERATOR_GUIDE.md) — daily workflows
- [Troubleshooting](TROUBLESHOOTING.md) — common issues and fixes
- [Debug Mode](DEBUG_MODE.md) — capturing diagnostics
- [Local Sessions](LOCAL_SESSIONS.md) — how local Copilot sessions are detected
- [Security](SECURITY.md) — safety expectations and controls

## Views

| Key | View | What it shows |
|-----|------|--------------|
| _(default)_ | **Session list** | All sessions with status, badges, PR indicators |
| `enter` | **Detail** | Full session info, timeline, token usage, attention reason |
| `K` | **Kanban** | Sessions in columns: In Progress / Idle / Done |
| `M` | **Mission Control** | Fleet summary dashboard with per-repo breakdown |
| `l` | **Logs** | Raw markdown session log with live tailing |
| `c` | **Conversation** | Styled chat bubbles for session dialogue |
| `t` | **Tool Timeline** | Chronological trace of agent tool executions |
| `d` | **Diff** | Colored PR diff (green/red) |
| `?` | **Help** | Full keybinding reference |

## Developer docs

- [Developer Workflow](DEVELOPER_WORKFLOW.md) — build, test, and smoke flow
- [Decisions](DECISIONS.md) — architecture decision log
