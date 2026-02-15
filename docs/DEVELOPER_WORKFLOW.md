# Developer Workflow

Use the Makefile for the fastest local loop.

## Daily commands

```bash
make build
make test
make smoke
```

## Full validation (CI-like)

```bash
make ci
```

## Useful extras

```bash
make test-race
make coverage
make fmt
make lint
make clean
```

## Why this exists

- one command set for contributors
- fewer copy/paste mistakes
- same flow in local dev and automation
