# gh-agent-viz -> gh-atc Strategy

## Executive decision
Use a **dual-track product strategy**:

1. **This repository (`gh-agent-viz`)** becomes the terminal-first operational hub (working title: **gh-atc**) for live Copilot session supervision.
2. **Private companion repository (`maxbeizer/copilot-atc`)** focuses on analytics, spend intelligence, and optimization recommendations.

## Why this split works

### What belongs in the CLI TUI
- real-time state visibility (running/done/failed)
- rapid operator actions (open PR, open logs, resume session)
- lightweight local metrics snapshots
- workflow-first triage and supervision

### What belongs in the private analytics platform
- long-horizon usage analytics and trends
- billing/token spend aggregation and projections
- optimization recommendations based on historical patterns
- potentially richer visualizations and policy/reporting flows

## Tauri question: should we use it?

Tauri can make sense for the analytics product (desktop dashboards, richer charts), but it is not required for the operator hub.

Recommended order:
1. Keep terminal-first execution in `gh-agent-viz` for speed and adoption.
2. Define an analytics data contract.
3. Prototype analytics UI separately (web or Tauri) once data quality and model usefulness are validated.

## Naming recommendation

- Product language: **Copilot ATC (Air Traffic Control)**.
- Near-term implementation:
  - retain current repository name for continuity
  - introduce `gh-atc` as the preferred product/command alias in docs and roadmap
  - add migration messaging if command/repo rename happens later

## Immediate next steps
1. Finalize naming transition plan (issue #16).
2. Define CLI <-> analytics contract (issue #17).
3. Continue shipping core ATC features in this repo while private analytics evolves independently.
