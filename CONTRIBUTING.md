# Contributing to gh-agent-viz

Thanks for your interest in contributing!

## Getting started

```bash
git clone https://github.com/maxbeizer/gh-agent-viz.git
cd gh-agent-viz
make build
make test
make relink-local  # install locally for testing
```

## Development workflow

1. Create a branch from `main`
2. Make your changes
3. Run `go build ./...` and `go test ./...`
4. Run `gh agent-viz --demo` to visually verify your changes
5. Open a PR against `main`

## Code organization

- `cmd/` — CLI entry point (Cobra)
- `internal/tui/` — Bubble Tea TUI (split into ui.go, commands.go, keyhandlers.go, helpers.go)
- `internal/tui/components/` — UI components (each in its own package)
- `internal/data/` — Data fetching and parsing
- `internal/config/` — YAML config parser

## Conventions

- Each TUI component lives in its own package under `internal/tui/components/`
- Use shared helpers from `internal/data/session.go` (don't duplicate)
- Tests must be deterministic — no time.Now() in assertions, no network calls
- Key handlers go in `keyhandlers.go`, data fetching in `commands.go`
- Run `gh agent-viz --demo` to test UI changes without real data

## Commit messages

Follow conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
