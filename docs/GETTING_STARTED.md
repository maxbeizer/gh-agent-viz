# Getting Started

If the board feels confusing at first, start here.

## 1) Launch with focus

```bash
gh agent-viz --repo owner/repo
```

Starting with one repo makes the board much easier to read.

## 2) What you are looking at

The main screen has a **Sessions at a Glance strip**, a **3-column board**, and a **Session Summary** panel for the highlighted row.

Columns:

- **Running** = active or queued sessions
- **Done** = completed sessions
- **Failed** = sessions that need attention

Each row is labeled for fast scanning:

`status icon + title (+ badge)`
`Repository: ...`
`Needs your action: ... • Last update: ...`

Attention reasons are explicit:

- `waiting on your input`
- `run failed`
- `running but quiet` (running/queued but stale)
- `no action needed`

Example:

`🟢 Add retry logic`
`Repository: maxbeizer/gh-agent-viz`
`Needs your action: no action needed • Last update: 5m ago`

## 3) Why you may see “Untitled Session” or “not available / not recorded”

This usually means older/local session metadata is incomplete.

- `Untitled Session` = session didn’t store a usable summary/title
- `not available` = repository/branch metadata was unavailable
- `not recorded` = no reliable timestamp signal was found

To reduce noise:

1. Use `--repo owner/repo`
2. Press `a` to jump straight to sessions that need your attention
3. Press `r` to refresh

## 4) Core controls (minimum set)

- `h` / `→`: switch columns
- `j` / `k`: move selection
- `enter`: open details
- `l`: open logs (remote rows)
- `o`: open PR (remote agent rows)
- `s`: resume active local session
- `tab` / `shift+tab`: change filter forward/backward
- `a`: toggle needs-action view
- `q`: quit

## 5) First useful workflow

1. Filter to needs-action view (`a`)
2. Open a row (`enter`)
3. Check logs (`l`) if needed
4. Open PR (`o`) or resume (`s`) depending on row source

## 6) Views & Visualizations

gh-agent-viz includes several visual features beyond the default list view:

- **Kanban board** (`K`) — status-column layout for monitoring many sessions at once
- **Toast notifications** — automatic alerts when session statuses change
- **Session timeline bar** — Unicode lifecycle visualization in detail view
- **Dependency graph** — related session visualization in detail view
- **Color themes** — configurable presets (catppuccin-mocha, dracula, tokyo-night)
- **Live log tailing** — real-time log streaming with follow mode (`f` in log viewer)

See [UI_FEATURES.md](UI_FEATURES.md) for the full guide.
