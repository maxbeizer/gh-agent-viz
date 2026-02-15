# Debug Mode

Use debug mode when something fails and you need exact command output.

## Enable

```bash
gh agent-viz --debug
```

## What it does

- Logs data-layer `gh` command execution details
- Logs UI action commands (resume session, open PR)
- Captures command, status, and output
- Shows an in-app `DEBUG ON` banner with the active log path

Log file:

`~/.gh-agent-viz-debug.log`

## Typical debug workflow

1. Run with `--debug`
2. Reproduce the issue (refresh, open logs, open PR, etc.)
3. Inspect the log:
   ```bash
   tail -n 120 ~/.gh-agent-viz-debug.log
   ```
4. Share relevant lines in an issue

## What to look for

- `unknown flag` errors from `gh` commands
- auth/permission failures
- repository scoping mistakes
- malformed command output

## Safety note

Debug logs may include repository names and command outputs. Review before sharing publicly.
