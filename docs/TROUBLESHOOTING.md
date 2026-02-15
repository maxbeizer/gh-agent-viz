# Troubleshooting

Quick fixes for the most common `gh-agent-viz` issues.

## Fast pre-checks

```bash
gh --version
gh auth status
gh agent-viz --help
```

## No sessions visible

Try:

1. Press `r` to refresh.
2. Press `a` to check attention mode, then `tab` to cycle other filters in case rows are hidden.
3. Run with repo scope:
   ```bash
   gh agent-viz --repo owner/repo
   ```
4. Check raw source command:
   ```bash
   gh agent-task list
   ```

## “failed to fetch agent tasks”

Common causes:

- `gh` not authenticated
- insufficient access to repo/session data
- temporary GitHub CLI/Copilot backend issue

Actions:

```bash
gh auth status
gh auth refresh -s repo,read:org
```

Then retry `gh agent-viz`.

## `o` (open PR) fails

`o` works only when the selected row has PR metadata (typically remote agent-task rows).

Use `enter` first to confirm PR fields exist.

## `s` (resume) fails

`s` only works for **local**, **active** sessions (`running` or `queued`).

If the session is remote, completed, or failed, resume is intentionally blocked.

## Debug mode for hard issues

```bash
gh agent-viz --debug
```

Then inspect:

`~/.gh-agent-viz-debug.log`

The log includes command, status, and output for data-layer calls.

## UI looks broken

- Use a larger terminal (at least 80x24)
- Ensure color-capable terminal:
  ```bash
  echo $TERM
  ```
  Prefer `xterm-256color` or equivalent.
