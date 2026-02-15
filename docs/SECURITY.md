# Security Guidelines

This project is a terminal-side control surface for Copilot workflows. Security is critical because it touches local session artifacts, GitHub-authenticated CLI operations, and potentially future analytics pipelines.

## Security Principles

1. **Least privilege by default**
   - Only read files and execute commands required for the requested feature.
2. **No secret exposure**
   - Never log or commit tokens, credentials, cookies, or sensitive local metadata.
3. **Untrusted input handling**
   - Treat `gh` command output and local session files as untrusted; parse defensively.
4. **Explicit error visibility**
   - Do not hide security-relevant failures behind silent fallbacks.
5. **Data minimization**
   - Collect/store only the minimum data needed for UI and diagnostics.

## Threat Surface

- Parsing local session metadata (`~/.copilot/session-state`)
- Running shell commands (`gh agent-task`, `gh pr`, `gh copilot`)
- Rendering user-controlled text in TUI
- Future telemetry/analytics integrations

## Required Engineering Practices

- Use `exec.Command` with argument arrays; avoid command string interpolation.
- Validate IDs/URLs before using them for actions such as resume/open.
- Handle malformed files and command failures without panics.
- Keep actionable errors user-facing and avoid leaking sensitive payloads.
- Add tests for malformed/hostile input and parser edge cases.

## Data & Privacy Guardrails

- Local session data stays local by default.
- Any remote analytics upload must be opt-in and documented.
- New data fields collected for analytics must be justified and reviewed.

## Security Review Triggers

Require explicit review when changes involve:
- new external network destinations
- new persistent storage of session/user data
- credential/token handling paths
- command execution paths or file-access expansion
