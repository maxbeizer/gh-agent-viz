# Integration Test Agent Profile

This agent profile defines the execution context for integration smoke tests in gh-agent-viz.

## Role

Integration Test Agent - validates key user journeys and critical paths before merging.

## Capabilities

- Build validation
- CLI help and usage validation
- Navigation path testing (keyboard shortcuts, TUI interactions)
- Action path testing (view details, view logs, refresh, quit)
- Error handling validation

## Execution Context

### Environment
- CI environment or local development machine
- Go 1.21+ installed
- GitHub CLI (`gh`) not required for smoke tests (mocked)

### Test Scope

1. **Build Path**
   - Project builds without errors
   - Binary is executable
   - No missing dependencies

2. **Help Path**
   - `--help` flag displays usage information
   - Help text includes all documented flags
   - Help text is properly formatted

3. **Navigation Path**
   - Binary launches without crashing
   - Validates command structure
   - Validates flag parsing

4. **Action Path**
   - Version/about information accessible
   - Configuration loading (when config file exists)
   - Graceful handling when no gh CLI available

## Success Criteria

All smoke tests must pass for the build to be considered healthy:
- Exit code 0 for successful tests
- Clear, actionable error messages on failure
- Fast execution (< 30 seconds total)
- No flaky tests

## Failure Triage

When tests fail:
1. Check build logs for compilation errors
2. Review test output for specific assertion failures
3. Verify environment prerequisites (Go version, dependencies)
4. Check for breaking changes in dependencies

## Usage

### Local execution
```bash
./test/integration/smoke_test.sh
```

### CI execution
Automatically run in GitHub Actions CI pipeline on all PRs and main branch commits.
