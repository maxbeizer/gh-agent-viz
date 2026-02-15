# Manual Test Scenarios

## Board and Navigation
1. Board renders Running/Done/Failed columns with correct counts.
2. `h`/`left` and `right` switch columns; `j/k` move within column.
3. Empty column shows clear placeholder and does not panic.

## Data Correctness
4. Recently updated sessions sort to top in each column.
5. Failed/cancelled sessions appear in Failed column.
6. Completed sessions appear in Done column.

## Actions
7. `enter` opens detail for selected item.
8. `l` opens logs for selected item when available.
9. `o` opens PR for selected item with URL and number fallback paths.
10. Resume action (new): active local session resumes in Copilot CLI.

## Reliability
11. Startup without session data shows safe empty state.
12. Parsing malformed local session files does not crash app.
13. Refresh keeps selection stable when possible.
