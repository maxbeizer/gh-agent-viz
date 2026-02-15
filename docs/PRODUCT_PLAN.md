# gh-agent-viz Product Plan

## Outcome
Ship a compelling session command center for Copilot workflows with clear state visibility, rapid navigation, and supervisory controls.

## Stream A: Unified Session Data Layer
### Scope
- Add local Copilot CLI session ingestion from `~/.copilot/session-state`.
- Normalize data shape between local sessions and `gh agent-task` sessions.
- Add source type and canonical status mapping.

### Acceptance
- Board can render mixed sources without breaking existing agent-task behavior.
- Unknown/partial data handled safely.

## Stream B: UX and Navigation
### Scope
- Strengthen Kanban board interactions (column counts, selection persistence, empty states).
- Add "resume session" action for in-progress Copilot CLI sessions.
- Improve header/footer with clear mode + action hints.

### Acceptance
- User can move quickly from board -> detail/log/resume.
- Actions degrade gracefully when unsupported for selected row.

## Stream C: Testing and Quality
### Scope
- Add parser tests for local session metadata/event status inference.
- Add integration smoke scenarios and scripted checks.
- Create integration-test agent instructions for repeatable validation.

### Acceptance
- CI covers core transformations and navigation logic.
- Integration smoke tests validate top user paths.

## Stream D: Documentation and Adoption
### Scope
- README updates for new session sources and controls.
- "Operator guide" for supervising many Copilot workstreams.
- troubleshooting section for missing/invalid local session data.

### Acceptance
- New users can discover key workflows in <5 minutes.

## Execution Plan
1. Build unified session abstractions and ingestion.
2. Wire board/actions to source-aware behavior.
3. Land tests + integration runner.
4. Publish docs and polish.

## Risk Management
- **Data drift risk**: local session schema may vary by Copilot version.
  - Mitigation: tolerant parsing + explicit fallbacks.
- **Action safety risk**: invalid resume targets.
  - Mitigation: strict ID validation + clear user errors.
- **UX overload risk**: too many actions at once.
  - Mitigation: progressive disclosure and context-aware key hints.
