.PHONY: help check-go-version build run install-local relink-local test test-race coverage smoke ci lint fmt tidy clean

BINARY ?= bin/gh-agent-viz
GO ?= go
GO_MIN_MAJOR ?= 1
GO_MIN_MINOR ?= 24
GO_MIN_PATCH ?= 2

check-go-version:
	@current="$$( $(GO) env GOVERSION 2>/dev/null | sed 's/^go//' )"; \
	current_major="$${current%%.*}"; \
	current_minor="$${current#*.}"; \
	current_minor="$${current_minor%%.*}"; \
	current_patch="$${current#*.*.}"; \
	if [ "$$current_patch" = "$$current" ]; then current_patch=0; fi; \
	current_patch="$${current_patch%%.*}"; \
	if [ -z "$$current_major" ] || [ -z "$$current_minor" ] || [ -z "$$current_patch" ]; then \
		echo "Unable to detect Go version via '$(GO) env GOVERSION'."; \
		exit 1; \
	fi; \
	if [ "$$current_major" -lt "$(GO_MIN_MAJOR)" ] || { [ "$$current_major" -eq "$(GO_MIN_MAJOR)" ] && { [ "$$current_minor" -lt "$(GO_MIN_MINOR)" ] || { [ "$$current_minor" -eq "$(GO_MIN_MINOR)" ] && [ "$$current_patch" -lt "$(GO_MIN_PATCH)" ]; }; }; }; then \
		echo "Go $$current detected. gh-agent-viz requires Go $(GO_MIN_MAJOR).$(GO_MIN_MINOR).$(GO_MIN_PATCH)+."; \
		echo "Upgrade Go and retry."; \
		exit 1; \
	fi

help:
	@echo "gh-agent-viz developer commands"
	@echo ""
	@echo "  make build       Build ./$(BINARY)"
	@echo "  make run         Build and run locally"
	@echo "  make install-local  Install extension from current checkout"
	@echo "  make relink-local   Reinstall local extension link"
	@echo "  make test        Run unit tests"
	@echo "  make test-race   Run tests with race + coverage.out"
	@echo "  make coverage    Print coverage summary (requires coverage.out)"
	@echo "  make smoke       Run integration smoke script"
	@echo "  make ci          Run build + vet + test-race"
	@echo "  make lint        Run golangci-lint if installed"
	@echo "  make fmt         Format all Go packages"
	@echo "  make tidy        Run go mod tidy"
	@echo "  make clean       Remove build artifacts"

build: check-go-version
	@mkdir -p $(dir $(BINARY))
	$(GO) build -o $(BINARY) ./gh-agent-viz.go

run: build
	./$(BINARY)

install-local:
	gh extension install .

relink-local: build
	@if gh extension list | grep -qE '^gh agent-viz[[:space:]]'; then \
		gh extension remove agent-viz; \
	fi
	$(GO) build -o gh-agent-viz ./gh-agent-viz.go
	gh extension install .

test: check-go-version
	$(GO) test ./...

test-race: check-go-version
	$(GO) test -v -race -coverprofile=coverage.out ./...

coverage:
	$(GO) tool cover -func=coverage.out

smoke:
	./test/integration/smoke_test.sh

ci: check-go-version
	$(GO) build ./... && $(GO) vet ./... && $(GO) test -v -race -coverprofile=coverage.out ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed; skipping"; \
	fi

fmt: check-go-version
	$(GO) fmt ./...

tidy: check-go-version
	$(GO) mod tidy

clean:
	rm -rf bin coverage.out
