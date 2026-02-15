.PHONY: help build run test test-race coverage smoke ci lint fmt tidy clean

BINARY ?= bin/gh-agent-viz
GO ?= go

help:
	@echo "gh-agent-viz developer commands"
	@echo ""
	@echo "  make build       Build ./$(BINARY)"
	@echo "  make run         Build and run locally"
	@echo "  make test        Run unit tests"
	@echo "  make test-race   Run tests with race + coverage.out"
	@echo "  make coverage    Print coverage summary (requires coverage.out)"
	@echo "  make smoke       Run integration smoke script"
	@echo "  make ci          Run build + vet + test-race"
	@echo "  make lint        Run golangci-lint if installed"
	@echo "  make fmt         Format all Go packages"
	@echo "  make tidy        Run go mod tidy"
	@echo "  make clean       Remove build artifacts"

build:
	@mkdir -p $(dir $(BINARY))
	$(GO) build -o $(BINARY) ./gh-agent-viz.go

run: build
	./$(BINARY)

test:
	$(GO) test ./...

test-race:
	$(GO) test -v -race -coverprofile=coverage.out ./...

coverage:
	$(GO) tool cover -func=coverage.out

smoke:
	./test/integration/smoke_test.sh

ci:
	$(GO) build ./... && $(GO) vet ./... && $(GO) test -v -race -coverprofile=coverage.out ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed; skipping"; \
	fi

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf bin coverage.out
