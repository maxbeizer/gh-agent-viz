# Getting Started

If the board feels confusing at first, start here.

## 1) Launch with focus

```bash
gh agent-viz --repo owner/repo
```

Starting with one repo makes the board much easier to read.

## 2) What you are looking at

The main screen has an **ATC overview strip**, a **3-column board**, and a **Flight Deck** panel for the selected row.

Columns:

- **Running** = active or queued sessions
- **Done** = completed sessions
- **Failed** = sessions that need attention

Each row is:

`status icon + title`
`repository • source • last updated`
`↳ contextual hint (for selected row)`

When a session likely needs your reply, you'll see a `🧑 input needed` badge.

Example:

`🟢 Add retry logic`
`maxbeizer/gh-agent-viz • local • 5m ago`

## 3) Why you may see “Untitled Session” or “unknown”

This usually means older/local session metadata is incomplete.

- `Untitled Session` = session didn’t store a usable summary/title
- `unknown` = no reliable timestamp/status signal was found

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
