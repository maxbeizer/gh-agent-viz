# Troubleshooting Guide

This guide helps diagnose and resolve common issues when using `gh-agent-viz`.

## Pre-flight Checks

Run these commands to verify your environment:

```bash
# Check GitHub CLI installation
gh --version

# Verify authentication
gh auth status

# Note: The `gh agent-task` command is used internally by gh-agent-viz
# It may not be available in all GitHub CLI versions or configurations
# You can test if it's available with:
gh agent-task list
```

## Authentication Problems

### GitHub CLI Not Found

**Problem:** Terminal shows "gh: command not found"

**Fix:** Install GitHub CLI from https://cli.github.com/

Verify with: `gh --version`

### Not Authenticated

**Problem:** Error message about authentication or unauthorized access

**Fix:** 

```bash
# Log in to GitHub
gh auth login

# Verify status
gh auth status

# If needed, refresh token with required scopes
gh auth refresh -s repo,read:org
```

### Agent Task Command Missing

**Problem:** `gh agent-task` command not recognized

**Note:** The `gh agent-task` command is what `gh-agent-viz` uses internally to fetch session data. This command is part of GitHub's Copilot infrastructure and may not be publicly available or enabled for all accounts yet.

**Fix:**

The agent-task commands may not be available in all GitHub CLI versions. Check with your GitHub administrator if this feature is enabled for your account.

```bash
# Try listing extensions
gh extension list

# Update all extensions
gh extension upgrade --all
```

**Important:** Even if `gh agent-task` is not directly available to you, `gh-agent-viz` may still work if it has been provided access through other means. Try running `gh agent-viz` to see if it works.

## Data Display Problems

### Empty Task List

**Problem:** No tasks appear in the table

**Possible causes and fixes:**

1. **No sessions exist** - Create a test session:
   ```bash
   # Verify with raw command
   gh agent-task list
   ```

2. **Wrong repository scope** - Check your config file or `--repo` flag:
   ```yaml
   # ~/.gh-agent-viz.yml
   repos:
     - owner/correct-repo-name  # Format: owner/repo
   ```

3. **Status filter hiding results** - Press `tab` to cycle through filters (All → Active → Completed → Failed)

### Stale Data

**Problem:** Task status doesn't reflect current state

**Fix:**

1. Press `r` to manually refresh
2. Check `refreshInterval` in `~/.gh-agent-viz.yml`
3. Restart the TUI: press `q` then relaunch

### Parse Errors

**Problem:** "failed to parse agent tasks" error

**Diagnosis:**

```bash
# Check raw output format
gh agent-task list --json

# If JSON fails, try plain text
gh agent-task list
```

**Fix:**

- The tool falls back to plain text parsing automatically
- If both fail, update your GitHub CLI: `gh extension upgrade --all`
- File a bug report with the output at https://github.com/maxbeizer/gh-agent-viz/issues

## Action Failures

### Open PR Action Not Working

**Problem:** Pressing `o` doesn't open browser

**Fixes:**

1. Verify PR URL exists: Press `enter` to view details, check for "PR URL" field
2. Test browser manually:
   ```bash
   # macOS
   open https://github.com
   
   # Linux
   xdg-open https://github.com
   ```
3. Copy URL manually if needed

### Log Viewer Errors

**Problem:** Pressing `l` shows "agent logs require a session ID"

**Explanation:** Some tasks don't have detailed log data. This happens when:
- Session was created as a PR without agent-task tracking
- Logs expired or were never created
- Session predates agent-task logging

**Workaround:**
- Press `o` to open PR in browser
- Check PR "Checks" tab for workflow logs
- Review commit history

### Slow or Hanging Refresh

**Problem:** Pressing `r` causes long pause or timeout

**Fixes:**

1. **Reduce scope** - Monitor fewer repositories:
   ```yaml
   repos:
     - owner/priority-repo-only
   ```

2. **Increase interval** - Reduce API call frequency:
   ```yaml
   refreshInterval: 90
   ```

3. **Check rate limits:**
   ```bash
   gh api rate_limit
   ```

4. **Use single-repo mode:**
   ```bash
   gh agent-viz --repo owner/one-repo
   ```

## Display Problems

### Corrupted or Garbled Output

**Problem:** Text overlaps, colors missing, broken layout

**Fixes:**

1. **Increase terminal size** - Minimum 80x24 characters recommended
2. **Check terminal type:**
   ```bash
   echo $TERM
   # Should be xterm-256color or similar
   
   # Set if needed:
   export TERM=xterm-256color
   ```
3. **Use a modern terminal:**
   - macOS: iTerm2
   - Linux: GNOME Terminal, Alacritty
   - Windows: Windows Terminal

### Missing Colors

**Problem:** Status icons and table appear monochrome

**Fix:**

```bash
# Enable 256-color support
export TERM=xterm-256color

# Test colors work
printf "\x1b[38;5;196mTest\x1b[0m\n"
```

If colors still don't work, your terminal may not support them. Try a different terminal emulator.

## Configuration Problems

### Config File Ignored

**Problem:** Settings in `~/.gh-agent-viz.yml` don't take effect

**Checks:**

1. **Verify location:**
   ```bash
   ls -la ~/.gh-agent-viz.yml
   ```

2. **Validate YAML syntax:**
   - Use spaces, not tabs for indentation
   - Proper format: `key: value`
   - No trailing spaces

3. **Example valid config:**
   ```yaml
   repos:
     - owner/repo1
     - owner/repo2
   refreshInterval: 30
   defaultFilter: all
   ```

4. **Test minimal config:**
   ```yaml
   refreshInterval: 45
   ```
   If this works, add settings incrementally to find issues.

### Repository Names Not Working

**Problem:** Repos listed in config show no tasks

**Common mistakes:**

```yaml
# WRONG - includes URL
repos:
  - https://github.com/owner/repo

# WRONG - missing owner
repos:
  - repo-name

# CORRECT
repos:
  - owner/repo-name
```

**Verify access:**

```bash
# Test each repo individually
gh agent-viz --repo owner/repo1
gh repo view owner/repo1
```

## Build and Installation Issues

### Go Not Found

**Problem:** Building from source fails with "go: command not found"

**Fix:**

Install Go 1.21 or later from https://go.dev/dl/

Verify: `go version`

### Build Errors

**Problem:** `go build` shows errors or missing dependencies

**Fix:**

```bash
cd /path/to/gh-agent-viz

# Clean and update dependencies
go clean -modcache
go mod tidy
go mod download

# Rebuild
go build -o gh-agent-viz ./gh-agent-viz.go
```

## Performance Issues

### Slow or Laggy Interface

**Problem:** Keypresses delayed, choppy scrolling, slow refresh

**Solutions:**

1. **Filter aggressively** - Use `tab` to show only Active or Failed sessions
2. **Increase refresh interval:**
   ```yaml
   refreshInterval: 120  # 2 minutes
   ```
3. **Monitor fewer repos:**
   ```yaml
   repos:
     - owner/critical-repo-only
   ```
4. **Try faster terminal** - Alacritty, Kitty, or WezTerm

## Debugging Tips

### Inspect Raw Data

When something doesn't work, check the underlying commands:

```bash
# List sessions
gh agent-task list

# View session detail
gh agent-task view <SESSION_ID>

# View logs
gh agent-task view <SESSION_ID> --log

# Export for inspection
gh agent-task list --json > /tmp/sessions.json
cat /tmp/sessions.json
```

### System Information

Collect this info when filing bug reports:

```bash
# OS and terminal
uname -a
echo $TERM

# Tool versions
gh --version
go version  # If building from source

# Auth status
gh auth status
```

## Getting Help

### Filing Bug Reports

Include in your report:

1. **What you expected** vs **what happened**
2. **Steps to reproduce**
3. **System info** (commands above)
4. **Error messages** (full text or screenshot)
5. **Sample output** from `gh agent-task list --json`

Submit at: https://github.com/maxbeizer/gh-agent-viz/issues

### Community Resources

- **Issues:** https://github.com/maxbeizer/gh-agent-viz/issues
- **Discussions:** https://github.com/maxbeizer/gh-agent-viz/discussions

## Related Guides

- [README.md](../README.md) - Installation and basic usage
- [OPERATOR_GUIDE.md](OPERATOR_GUIDE.md) - Supervising multiple workstreams
- [DECISIONS.md](DECISIONS.md) - Architecture and design rationale
