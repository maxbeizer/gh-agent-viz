# Getting Started

If the board feels confusing at first, start here.

## 1) Launch with focus

```bash
gh agent-viz --repo owner/repo
```

Starting with one repo makes the board much easier to read.

## 2) What you are looking at

The main screen has an **ATC overview strip**, a **3-column board**, and a **Selected Session** panel for the highlighted row.

Columns:

- **Running** = active or queued sessions
- **Done** = completed sessions
- **Failed** = sessions that need attention

Each row is labeled for fast scanning:

`status icon + title (+ badge)`
`Repository: ...`
`Attention: ... • Last update: ...`

Attention reasons are explicit:

- `needs your input`
- `failed`
- `active but quiet` (running/queued but stale)
- `no action needed`

Example:

`🟢 Add retry logic`
`Repository: maxbeizer/gh-agent-viz`
`Attention: no action needed • Last update: 5m ago`

## 3) Why you may see “Untitled Session” or “not linked / not recorded”

This usually means older/local session metadata is incomplete.

- `Untitled Session` = session didn’t store a usable summary/title
- `not linked` = repository/branch metadata was unavailable
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
- `a`: toggle attention mode
- `q`: quit

## 5) First useful workflow

1. Filter to `attention` (`a`)
2. Open a row (`enter`)
3. Check logs (`l`) if needed
4. Open PR (`o`) or resume (`s`) depending on row source
