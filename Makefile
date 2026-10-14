# ==============================================================================
# mxping - Makefile
# Fast, accurate email verification & mailbox pinging CLI
# ==============================================================================

# Application Metadata
BINARY_NAME     := mxping
MODULE          := github.com/yashbaddi/mxping
CMD_PATH        := ./cmd/mxping
VERSION_PKG     := $(MODULE)/internal/cli

# Build Variables & Dynamic Git Info
VERSION         ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
COMMIT_HASH     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME      := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Directories
BUILD_DIR       := bin
DIST_DIR        := dist

# Go Compiler Flags
LDFLAGS         := -s -w \
                   -X '$(VERSION_PKG).AppVersion=$(VERSION)'
GOFLAGS         := -trimpath

# Cross Compilation Target Platforms: OS/ARCH pairs
PLATFORMS       := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# Shell configuration
SHELL           := /bin/bash
.DEFAULT_GOAL   := help

# Terminal formatting
GREEN           := \033[0;32m
YELLOW          := \033[0;33m
CYAN            := \033[0;36m
RESET           := \033[0m

# ==============================================================================
# Development Targets
# ==============================================================================

## Display this help menu with descriptions
.PHONY: help
help:
	@echo ""
	@echo "mxping build automation commands:"
	@echo ""
	@awk '/^## / { desc = substr($$0, 4); next } desc && /^[a-zA-Z0-9_-]+:/ { split($$0, a, ":"); printf "  \033[0;32m%-18s\033[0m %s\n", a[1], desc; desc = "" }' $(MAKEFILE_LIST)
	@echo ""

## Build static binary for host OS and architecture (placed in bin/)
.PHONY: build
build:
	@echo -e "$(CYAN)--> Building $(BINARY_NAME) $(VERSION) for host...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo -e "$(GREEN)✓ Binary generated at $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

## Build and run mxping directly with optional ARGS (e.g. make run ARGS="user@example.com")
.PHONY: run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## Run all project unit tests
.PHONY: test
test:
	@echo -e "$(CYAN)--> Running unit test suite...$(RESET)"
	go test -v -race ./...

## Run tests and generate HTML code coverage report
.PHONY: test-coverage
test-coverage:
	@echo -e "$(CYAN)--> Generating test coverage report...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo -e "$(GREEN)✓ Coverage report generated at $(BUILD_DIR)/coverage.html$(RESET)"

## Format all Go source files with gofmt
.PHONY: fmt
fmt:
	@echo -e "$(CYAN)--> Formatting Go files...$(RESET)"
	gofmt -s -w .
	@echo -e "$(GREEN)✓ Formatting complete$(RESET)"

## Run go vet static analysis
.PHONY: vet
vet:
	@echo -e "$(CYAN)--> Running go vet...$(RESET)"
	go vet ./...
	@echo -e "$(GREEN)✓ Go vet passed$(RESET)"

## Tidy and verify go modules
.PHONY: tidy
tidy:
	@echo -e "$(CYAN)--> Running go mod tidy...$(RESET)"
	go mod tidy
	@echo -e "$(GREEN)✓ Modules tidied$(RESET)"

## Install binary into $GOPATH/bin
.PHONY: install
install:
	@echo -e "$(CYAN)--> Installing $(BINARY_NAME) to $(shell go env GOPATH)/bin...$(RESET)"
	CGO_ENABLED=0 go install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(CMD_PATH)
	@echo -e "$(GREEN)✓ Installed successfully$(RESET)"

# ==============================================================================
# Cross-Compilation & Release Targets
# ==============================================================================

## Cross-compile static binaries for Linux, macOS, and Windows (amd64 & arm64)
.PHONY: cross-compile
cross-compile: clean
	@echo -e "$(CYAN)--> Starting cross-compilation for version $(VERSION)...$(RESET)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output_name="$(BINARY_NAME)-$${os}-$${arch}"; \
		if [ "$$os" = "windows" ]; then \
			output_name="$${output_name}.exe"; \
		fi; \
		echo -e "  $(YELLOW)• Compiling for $${os}/$${arch}...$(RESET)"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
		go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$${output_name} $(CMD_PATH) || exit 1; \
	done
	@echo -e "$(GREEN)✓ Cross-compilation complete! Binaries saved in $(DIST_DIR)/$(RESET)"
	@$(MAKE) checksums

## Generate SHA-256 checksums for all release binaries in dist/
.PHONY: checksums
checksums:
	@echo -e "$(CYAN)--> Calculating SHA-256 checksums...$(RESET)"
	@cd $(DIST_DIR) && \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum $(BINARY_NAME)-* > checksums.txt; \
	else \
		shasum -a 256 $(BINARY_NAME)-* > checksums.txt; \
	fi
	@echo -e "$(GREEN)✓ Checksums generated at $(DIST_DIR)/checksums.txt$(RESET)"
	@cat $(DIST_DIR)/checksums.txt

# ==============================================================================
# Cleanup
# ==============================================================================

## Remove all build artifacts, binary outputs, and coverage reports
.PHONY: clean
clean:
	@echo -e "$(CYAN)--> Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(BINARY_NAME) coverage.* *.out
	@echo -e "$(GREEN)✓ Clean complete$(RESET)"
