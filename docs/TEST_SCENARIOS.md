# Manual Test Scenarios

## Board and Navigation
1. Board renders Running/Done/Failed columns with correct counts.
2. `h`/`left` and `right` switch columns; `j/k` move within column.
3. Empty column shows clear placeholder and does not panic.

## Data Correctness
4. Sessions needing action (`needs-input`, failed, quiet-running) sort ahead of non-actionable rows in each column.
5. Repeated quiet duplicates are de-emphasized (badge + lower priority) without being hidden.
6. Failed/cancelled sessions appear in Failed column.
7. Completed sessions appear in Done column.

## Actions
8. `enter` opens detail for selected item.
9. `l` opens logs for selected item when available.
10. `o` opens PR for selected item with URL and number fallback paths.
11. Resume action (new): active local session resumes in Copilot CLI.

## Reliability
12. Startup without session data shows safe empty state.
13. Parsing malformed local session files does not crash app.
14. Refresh keeps selection stable when possible.
