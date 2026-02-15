# Operator Guide

Use this guide when you want to quickly supervise Copilot work from the terminal.

## 1) Launch

```bash
gh agent-viz
```

Optional:

```bash
gh agent-viz --repo owner/repo
gh agent-viz --debug
```

## 2) Read the board

The board includes:

- **ATC Overview**: total/active/done/failed/session-source counters
- **Three status columns**: active work lanes
- **Flight Deck**: selected-session context and recommended actions

Columns:

- **Running**: active or queued sessions
- **Done**: completed sessions
- **Failed**: sessions that need attention

## 3) Core keys (daily use)

| Key | What it does |
|---|---|
| `h` / `←` and `→` | Move between columns |
| `j` / `k` | Move up/down in a column |
| `enter` | Open details pane |
| `l` | Open log view (remote agent-task rows) |
| `o` | Open PR in browser (agent-task rows) |
| `s` | Resume active **local** Copilot session |
| `tab` / `shift+tab` | Cycle filter: all ↔ active ↔ completed ↔ failed |
| `r` | Refresh now |
| `q` | Quit |

## 4) Typical workflow

1. Start in **Running**.
2. Open details (`enter`) for a session you care about.
3. Jump to logs (`l`) if something looks off.
4. Open PR (`o`) for completed remote work.
5. Resume local active work (`s`) when you want to continue in Copilot CLI.

## 5) Recommended config

Create `~/.gh-agent-viz.yml`:

```yaml
repos:
  - owner/repo1
  - owner/repo2
refreshInterval: 30
defaultFilter: all
```

## 6) Debug mode

Run with:

```bash
gh agent-viz --debug
```

When commands fail, check:

`~/.gh-agent-viz-debug.log`

This is the fastest way to diagnose data-fetch and action failures.
