# CLI ↔ Analytics Contract (gh-agent-viz ↔ copilot-atc)

## Goal
Define a secure, incremental contract between the terminal operator hub and the private analytics platform.

## Scope and boundaries

- `gh-agent-viz` remains the real-time control surface (state + actions).
- `copilot-atc` owns long-horizon analytics (usage, cost, optimization).
- Default mode is local-only; remote sync is explicitly opt-in.

## Data model (v1)

### Session event envelope

```json
{
  "schemaVersion": "1.0",
  "eventId": "uuid",
  "capturedAt": "2026-02-15T00:00:00Z",
  "source": "gh-agent-viz",
  "session": {
    "id": "session-id",
    "sourceType": "local-copilot|agent-task",
    "status": "running|queued|completed|failed|unknown",
    "repository": "owner/repo",
    "branch": "feature/foo",
    "title": "short title",
    "prNumber": 123,
    "updatedAt": "2026-02-15T00:00:00Z"
  },
  "metrics": {
    "durationSeconds": 0,
    "tokenInput": 0,
    "tokenOutput": 0,
    "tokenTotal": 0,
    "estimatedCostUsd": 0
  },
  "metadata": {
    "operatorAction": "view|open_pr|view_logs|resume_session",
    "debugEnabled": false
  }
}
```

### Required fields (v1)

- `schemaVersion`, `eventId`, `capturedAt`
- `session.id`, `session.sourceType`, `session.status`
- `session.repository` (when known)
- `metadata.operatorAction`

### Optional fields (v1)

- token/cost fields (populate when data source supports them)
- branch/title/pr metadata

## Transport contract

### Local-first (default)

- Write NDJSON snapshots/events to a local file path (for example: `~/.gh-agent-viz/events.ndjson`).
- No network egress in default mode.

### Remote sync (opt-in)

- Config flag: `analytics.remote.enabled: true`
- HTTPS endpoint with mTLS or token auth.
- Exponential backoff, bounded retries, and explicit surfaced errors.

## Privacy and security rules

- No prompt content or raw conversation text in analytics payloads.
- Do not include secrets/tokens/environment variable values.
- Hash or redact user-identifying fields where not operationally required.
- Persist only minimum required event fields.
- Keep debug logs and analytics event streams separate.

## Versioning and compatibility

- Use semantic schema versions (`1.x` backward-compatible).
- Reject unknown major versions (`2.x`) with explicit errors.
- Reserve additive fields for minor versions.

## Rollout plan

1. Land contract doc and sample payloads (this doc).
2. Add local exporter interface in `gh-agent-viz` behind a feature flag.
3. Validate schema with integration smoke tests.
4. Add optional remote sink in private `copilot-atc` repo.
5. Add optimization recommendations once usage quality is stable.
