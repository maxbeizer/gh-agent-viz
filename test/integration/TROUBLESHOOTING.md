# Integration Smoke Tests - Troubleshooting Guide

This document provides guidance for triaging and resolving integration smoke test failures.

## Quick Links

- Test script: `test/integration/smoke_test.sh`
- Agent profile: `.github/agents/integration-test.agent.md`
- CI workflow: `.github/workflows/ci.yml`

## Running Tests Locally

```bash
# Run integration smoke tests
./test/integration/smoke_test.sh

# Run with verbose output
bash -x ./test/integration/smoke_test.sh

# Run specific test by editing the script to comment out other tests
vim test/integration/smoke_test.sh
```

## Common Failure Scenarios

### Build Failures

**Symptom**: "Project build" test fails

**Possible Causes**:
- Syntax errors in Go code
- Missing dependencies
- Import cycle
- Incompatible Go version

**Debugging**:
```bash
# Try building manually to see full error
go build -o gh-agent-viz ./gh-agent-viz.go

# Check for missing dependencies
go mod tidy
go mod verify

# Run go vet for static analysis
go vet ./...
```

### Help Text Failures

**Symptom**: "Help text contains description" or "Help text documents X flag" fails

**Possible Causes**:
- Help text was modified in cmd/root.go
- Flags were added/removed without updating tests

**Debugging**:
```bash
# Check actual help output
./gh-agent-viz --help

# Compare with expected text in smoke_test.sh
grep -A5 "Help text contains" test/integration/smoke_test.sh
```

**Fix**: Update test expectations in `test/integration/smoke_test.sh` to match actual help output, or fix the help text if it's incorrect.

### Navigation Failures

**Symptom**: "Invalid flag produces error message" fails

**Possible Causes**:
- Cobra error handling changed
- Error message format changed

**Debugging**:
```bash
# Test invalid flag handling
./gh-agent-viz --invalid-flag 2>&1

# Should output "unknown flag" error
```

### Binary Startup Failures

**Symptom**: "Binary starts without crashing" fails

**Possible Causes**:
- Panic in initialization code
- Required dependencies missing at runtime
- TUI framework initialization issues

**Debugging**:
```bash
# Try running the binary directly (it will fail to connect to gh, but shouldn't crash)
timeout 2 ./gh-agent-viz 2>&1

# Check for panics or unexpected errors
```

### Module Verification Failures

**Symptom**: "Go module dependencies verified" fails

**Possible Causes**:
- go.sum is out of sync with go.mod
- Corrupted module cache

**Debugging**:
```bash
# Update and verify modules
go mod tidy
go mod verify

# Clean and rebuild cache
go clean -modcache
go mod download
```

## CI-Specific Issues

### Timeout Issues

If tests timeout in CI:
1. Check if the timeout in the test script needs adjustment (currently 2 seconds for binary startup)
2. Verify CI runner has sufficient resources
3. Look for deadlocks or blocking operations in initialization code

### Environment Differences

Local vs CI differences to be aware of:
- CI runs in fresh Ubuntu environment
- No GitHub CLI authentication in CI (by design - tests don't need it)
- CI may have stricter timeouts
- Different terminal capabilities (CI is non-interactive)

## Updating Tests

When making intentional changes that affect test expectations:

1. Update the test assertions in `test/integration/smoke_test.sh`
2. Run tests locally to verify they pass
3. Update `.github/agents/integration-test.agent.md` if test scope changes
4. Update this troubleshooting guide if new failure modes are discovered

## Test Philosophy

These smoke tests are designed to:
- ✅ Catch build breakages
- ✅ Validate CLI interface contracts
- ✅ Ensure binary can start without crashing
- ✅ Verify module dependencies

They are NOT designed to:
- ❌ Test full TUI functionality (requires interactive terminal)
- ❌ Test GitHub API integration (requires authentication)
- ❌ Test detailed business logic (covered by unit tests)

Keep tests fast (< 30 seconds total) and focused on critical paths.
