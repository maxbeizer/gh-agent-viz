# gh-agent-viz Product Brief

## Product Thesis
`gh-agent-viz` should evolve from a "task viewer" into a **Copilot Session Command Center**: the fastest way for an engineer to understand what Copilot is doing, what needs attention, and where to jump in next.

## Problem
Today, users can run many Copilot-driven workflows (Copilot CLI sessions, coding-agent tasks, PR-linked sessions), but observability and control are fragmented:
- hard to see active vs completed work at a glance
- hard to detect stalled sessions or failures quickly
- hard to jump directly into the exact in-progress context

## Target Users
1. **Solo power users** running many concurrent Copilot sessions.
2. **Tech leads/managers** supervising multiple Copilot-generated workstreams.
3. **Maintainers** triaging and merging Copilot-created PRs safely.

## Jobs To Be Done
- "Show me what is running right now and what is done."
- "Tell me what changed in the last hour and what is blocked."
- "Let me jump straight into the exact live session or PR context."

## Product Principles
1. **State-first UX**: Running work is always visually dominant.
2. **One-keystroke control**: every frequent operation is keyboard-first.
3. **Trust via evidence**: each status should be backed by timestamp, logs, and linked artifact (session/PR).
4. **Progressive depth**: board -> detail -> logs -> resume session.

## Core Experience (V1 target)
- Kanban-like board focused on **Running / Done / Failed**.
- Unified session model that can include:
  - Copilot coding-agent sessions (`gh agent-task`)
  - local Copilot CLI sessions (`~/.copilot/session-state`)
- fast actions:
  - open PR
  - open logs
  - resume active Copilot CLI session
- stable sorting by recency and health cues.

## Differentiators
- First terminal-native control surface for both remote and local Copilot session state.
- Purpose-built for "manager of agents" workflows (parallel PR streams + verification loops).
- Actionable, not just observable: each row should support immediate next action.

## Success Metrics
- Time-to-find-running-session < 10 seconds.
- Time-to-resume-active-session < 15 seconds.
- Daily active usage and repeat usage growth.
- Reduced stale/forgotten Copilot sessions.

## Non-goals (near-term)
- Replacing IDE chat experiences.
- Full orchestration platform with scheduling/policy engines.
- Multi-tenant cloud backend.
