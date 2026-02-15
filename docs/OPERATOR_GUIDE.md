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

- **Sessions at a Glance**: total/running/done/failed plus `needs action` count
- **Three status columns**: active work lanes
- **Session Summary panel**: plain-language context and recommended actions for the highlighted row

Columns:

- **Running**: active or queued sessions
- **Done**: completed sessions
- **Failed**: sessions that need attention

Each card includes explicit labels so triage is immediate:

- `Repository:` shows repo context (`not available` if missing)
- `Needs your action:` explains why action is needed (`waiting on your input`, `run failed`, `running but quiet`, or `no action needed`)
- `Last update:` shows freshness (`not recorded` when timestamp metadata is missing)

## 3) Core keys (daily use)

| Key | What it does |
|---|---|
| `h` / `←` and `→` | Move between columns |
| `j` / `k` | Move up/down in a column |
| `enter` | Open details pane |
| `l` | Open log view (only shown for remote agent-task rows) |
| `o` | Open PR in browser (only shown when selected row has a linked PR) |
| `s` | Resume active **local** Copilot session |
| `a` | Toggle **needs-action view** (sessions needing your action) |
| `tab` / `shift+tab` | Cycle filter: all ↔ needs action ↔ running ↔ done ↔ failed |
| `r` | Refresh now |
| `q` | Quit |

## 4) Typical workflow

1. Start in **needs-action view** (`a`) to triage what needs you now.
2. Open details (`enter`) for a session you care about.
3. Jump to logs (`l`) if something looks off.
4. Open PR (`o`) for completed remote work.
5. Resume local active work (`s`) when you want to continue in Copilot CLI.

Footer hints are contextual: if an action is unavailable for the selected row, it is hidden.

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
