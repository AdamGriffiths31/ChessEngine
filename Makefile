# Gchess Makefile
# Provides clean builds and development tools

# Variables
GO := go
GOFLAGS := -ldflags="-s -w"
GOBUILDFLAGS := $(GOFLAGS) -trimpath
# -timeout 30m: internal/movegen (deep perft) exceeds go test's default 10m
# per-binary timeout under -race + coverage instrumentation on this
# hardware. Task 2.7 gates deep perft behind testing.Short(); revisit then.
GOTESTFLAGS := -race -coverprofile=coverage.out -timeout 30m
MODULE := github.com/AdamGriffiths31/ChessEngine
# Short commit SHA used to name saved benchmark result files (bench-save).
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)

# Binary names
MAIN_BINARY := chess-engine
UCI_BINARY := tools/bin/uci
BENCH_BINARY := tools/bin/bench

# Directories
BUILD_DIR := bin
TOOLS_DIR := tools
COVERAGE_DIR := coverage
DIST_DIR := dist

# Go needs a writable temp dir and build cache. Some shells (e.g. MSYS2
# make on Windows) strip TMP/TEMP/LocalAppData from the recipe environment,
# which makes `go` fall back to an unwritable system dir (or fail outright
# looking for %LocalAppData%). Point both at project-local directories
# instead so builds work regardless of the inherited environment.
# The *_DIR variables stay relative because make target/prerequisite names
# cannot contain spaces (CURDIR may); the exported env vars must be
# absolute because go subprocesses run with the package directory -- not
# the repo root -- as cwd (e.g. coverage builds of packages with no test
# files, and test binaries resolving t.TempDir), where a relative .gotmp
# does not exist.
GOTMPDIR_DIR := .gotmp
export GOTMPDIR := $(abspath $(GOTMPDIR_DIR))
GOCACHE_DIR := .gocache
export GOCACHE := $(abspath $(GOCACHE_DIR))

# OS and architecture detection
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Default target
.PHONY: all
all: clean build lint

# Help target
.PHONY: help
help:
	@echo "Gchess Build System"
	@echo "==================="
	@echo ""
	@echo "Build Targets:"
	@echo "  all           - Clean and build all binaries"
	@echo "  build         - Build all binaries"
	@echo "  main          - Build main chess engine binary"
	@echo "  uci           - Build UCI protocol binary"
	@echo "  benchbin      - Build bench multi-tool binary (sts/profile)"
	@echo ""
	@echo "Development Targets:"
	@echo "  test          - Run all tests with coverage"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-short    - Run short tests only"
	@echo "  bench         - Run Go benchmarks"
	@echo "  bench-save    - Save movegen/eval/search benchmarks to tools/results/bench_<sha>.txt"
	@echo "  bench-compare - Compare two bench-save results: OLD=<file> NEW=<file> (needs benchstat)"
	@echo "  lint          - Run linters (golangci-lint)"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  mod           - Tidy go modules"
	@echo ""
	@echo "Analysis Targets:"
	@echo "  coverage      - Generate coverage report"
	@echo "  coverage-html - Generate HTML coverage report"
	@echo "  perft         - Run perft tests"
	@echo ""
	@echo "Quick Commands:"
	@echo "  chess         - Clean, build, lint, and run chess engine"
	@echo "  dev           - Full development build with tests"
	@echo "  quick         - Quick format and build"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean         - Remove build artifacts"
	@echo "  clean-all     - Clean everything including logs and results"
	@echo "  install       - Install binaries to GOPATH/bin"
	@echo "  deps          - Download dependencies"
	@echo "  check-deps    - Check if required tools are installed"
	@echo ""
	@echo "Variables:"
	@echo "  GOOS=$(GOOS)"
	@echo "  GOARCH=$(GOARCH)"

# Build targets
.PHONY: build
build: deps fmt vet lint main uci benchbin

.PHONY: main
main: $(BUILD_DIR) $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Building main chess engine..."
	$(GO) build $(GOBUILDFLAGS) -o $(BUILD_DIR)/$(MAIN_BINARY) ./cmd/gchess

.PHONY: uci
uci: $(TOOLS_DIR)/bin $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Building UCI binary..."
	$(GO) build $(GOBUILDFLAGS) -o $(UCI_BINARY) ./cmd/uci

.PHONY: benchbin
benchbin: $(TOOLS_DIR)/bin $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Building bench binary..."
	$(GO) build $(GOBUILDFLAGS) -o $(BENCH_BINARY) ./cmd/bench

# Cross-compilation targets
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOBUILDFLAGS) -o $(BUILD_DIR)/$(MAIN_BINARY)-linux-amd64 ./cmd/gchess
	GOOS=linux GOARCH=amd64 $(GO) build $(GOBUILDFLAGS) -o $(TOOLS_DIR)/bin/uci-linux-amd64 ./cmd/uci

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOBUILDFLAGS) -o $(BUILD_DIR)/$(MAIN_BINARY)-windows-amd64.exe ./cmd/gchess
	GOOS=windows GOARCH=amd64 $(GO) build $(GOBUILDFLAGS) -o $(TOOLS_DIR)/bin/uci-windows-amd64.exe ./cmd/uci

.PHONY: build-macos
build-macos:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOBUILDFLAGS) -o $(BUILD_DIR)/$(MAIN_BINARY)-darwin-amd64 ./cmd/gchess
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOBUILDFLAGS) -o $(TOOLS_DIR)/bin/uci-darwin-amd64 ./cmd/uci

.PHONY: build-all-platforms
build-all-platforms: build-linux build-windows build-macos

# Development targets
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

.PHONY: mod
mod:
	@echo "Tidying modules..."
	$(GO) mod tidy

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

.PHONY: lint
lint: check-golangci-lint $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Running linters..."
	golangci-lint run ./...

# Testing targets
.PHONY: test
test: $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Running tests with coverage..."
	$(GO) test $(GOTESTFLAGS) ./...

.PHONY: test-verbose
test-verbose: $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Running tests with verbose output..."
	$(GO) test -v $(GOTESTFLAGS) ./...

.PHONY: test-short
test-short: $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Running short tests..."
	$(GO) test -short ./...

.PHONY: bench
bench: $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

.PHONY: bench-save
bench-save: $(TOOLS_DIR)/results $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Running benchmark suite (movegen/eval/search, count=10) for commit $(GIT_SHA)..."
	$(GO) test -bench . -benchmem -count 10 ./internal/movegen ./internal/eval ./internal/search > $(TOOLS_DIR)/results/bench_$(GIT_SHA).txt
	@echo "Saved to $(TOOLS_DIR)/results/bench_$(GIT_SHA).txt"

.PHONY: bench-compare
bench-compare:
	@if [ -z "$(OLD)" ] || [ -z "$(NEW)" ]; then \
		echo "Usage: make bench-compare OLD=<old-results.txt> NEW=<new-results.txt>"; \
		echo "  (result files are produced by 'make bench-save', e.g. tools/results/bench_<sha>.txt)"; \
		exit 1; \
	fi
	@command -v benchstat >/dev/null 2>&1 || (echo "benchstat not found. Install with: go install golang.org/x/perf/cmd/benchstat@latest" && exit 1)
	benchstat $(OLD) $(NEW)

.PHONY: perft
perft: $(GOTMPDIR_DIR) $(GOCACHE_DIR)
	@echo "Running perft tests..."
	$(GO) test -run TestPerft -v ./internal/movegen

# Coverage targets
.PHONY: coverage
coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html: test coverage-dir
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated at $(COVERAGE_DIR)/coverage.html"

# Utility targets
.PHONY: install
install: build
	@echo "Installing binaries..."
	$(GO) install ./cmd/gchess
	$(GO) install ./cmd/uci
	$(GO) install ./cmd/bench

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(TOOLS_DIR)/bin/chess_engine $(TOOLS_DIR)/bin/uci $(TOOLS_DIR)/bin/bench
	rm -rf $(TOOLS_DIR)/bin/*-linux-amd64* $(TOOLS_DIR)/bin/*-windows-amd64* $(TOOLS_DIR)/bin/*-darwin-amd64*
	rm -f coverage.out
	rm -rf $(COVERAGE_DIR)
	rm -rf $(DIST_DIR)
	rm -rf $(GOTMPDIR_DIR)
	rm -rf $(GOCACHE_DIR)

.PHONY: clean-all
clean-all: clean
	@echo "Cleaning everything..."
	rm -rf logs/
	rm -rf $(TOOLS_DIR)/results/*.pgn $(TOOLS_DIR)/results/*.log
	rm -rf ui/logs/
	rm -rf profiles/
	find . -name "*.test" -delete
	find . -name "*.prof" -delete

.PHONY: clean-logs
clean-logs:
	@echo "Cleaning log files..."
	rm -rf logs/
	rm -rf ui/logs/
	rm -rf $(TOOLS_DIR)/results/*.log

.PHONY: clean-results
clean-results:
	@echo "Cleaning benchmark results..."
	rm -rf $(TOOLS_DIR)/results/*.pgn

# Dependency checking
.PHONY: check-deps
check-deps:
	@echo "Checking required dependencies..."
	@command -v go >/dev/null 2>&1 || (echo "Go is not installed" && exit 1)
	@command -v golangci-lint >/dev/null 2>&1 || echo "Warning: golangci-lint not found. Install with: GOTOOLCHAIN=go1.26.4 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8"
	@command -v zip >/dev/null 2>&1 || echo "Warning: zip not found (only needed for 'make dist'). On Git Bash/MSYS2: pacman -S zip"
	@test -f $(TOOLS_DIR)/engines/cutechess-cli || echo "Warning: cutechess-cli not found at $(TOOLS_DIR)/engines/cutechess-cli"
	@test -f $(TOOLS_DIR)/engines/stockfish || echo "Warning: stockfish not found at $(TOOLS_DIR)/engines/stockfish"
	@echo "Dependency check complete."

.PHONY: check-golangci-lint
check-golangci-lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not found. Installing..." && GOTOOLCHAIN=go1.26.4 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8)

# Directory creation
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(TOOLS_DIR)/bin:
	mkdir -p $(TOOLS_DIR)/bin

$(TOOLS_DIR)/results:
	mkdir -p $(TOOLS_DIR)/results

$(GOTMPDIR_DIR):
	mkdir -p $(GOTMPDIR_DIR)

$(GOCACHE_DIR):
	mkdir -p $(GOCACHE_DIR)

coverage-dir:
	mkdir -p $(COVERAGE_DIR)

dist-dir:
	mkdir -p $(DIST_DIR)

# Distribution targets
.PHONY: dist
dist: dist-dir build-all-platforms
	@command -v zip >/dev/null 2>&1 || (echo "zip is required for 'make dist' (packages the Windows artifact). On Git Bash/MSYS2: pacman -S zip" && exit 1)
	@echo "Creating distribution packages..."
	# Linux
	tar -czf $(DIST_DIR)/chess-engine-linux-amd64.tar.gz -C $(BUILD_DIR) $(MAIN_BINARY)-linux-amd64 -C ../$(TOOLS_DIR)/bin uci-linux-amd64
	# Windows
	zip -j $(DIST_DIR)/chess-engine-windows-amd64.zip $(BUILD_DIR)/$(MAIN_BINARY)-windows-amd64.exe $(TOOLS_DIR)/bin/uci-windows-amd64.exe
	# macOS
	tar -czf $(DIST_DIR)/chess-engine-darwin-amd64.tar.gz -C $(BUILD_DIR) $(MAIN_BINARY)-darwin-amd64 -C ../$(TOOLS_DIR)/bin uci-darwin-amd64
	@echo "Distribution packages created in $(DIST_DIR)/"

# Quick development commands
.PHONY: chess
chess: clean all run

.PHONY: dev
dev: clean fmt vet lint test build

.PHONY: quick
quick: fmt build

.PHONY: run
run: main
	./$(BUILD_DIR)/$(MAIN_BINARY)

.PHONY: run-uci
run-uci: uci
	./$(UCI_BINARY)

# Debugging targets
.PHONY: debug
debug:
	@echo "Building with debug symbols..."
	$(GO) build -gcflags="-N -l" -o $(BUILD_DIR)/$(MAIN_BINARY)-debug ./cmd/gchess

.PHONY: race
race:
	@echo "Building with race detector..."
	$(GO) build -race -o $(BUILD_DIR)/$(MAIN_BINARY)-race ./cmd/gchess

# Profiling helpers
.PHONY: profile-cpu
profile-cpu: benchbin
	./$(BENCH_BINARY) profile -cpuprofile=cpu.prof

.PHONY: profile-mem
profile-mem: benchbin
	./$(BENCH_BINARY) profile -memprofile=mem.prof

# Git helpers
.PHONY: pre-commit
pre-commit: fmt vet lint test

# Show build info
.PHONY: version
version:
	@echo "Gchess Build Information"
	@echo "========================"
	@echo "Go version: $(shell go version)"
	@echo "OS/Arch: $(GOOS)/$(GOARCH)"
	@echo "Module: $(MODULE)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build time: $(shell date)"

# Ensure tools directory exists
$(TOOLS_DIR):
	mkdir -p $(TOOLS_DIR)

# Force rebuild of everything
.PHONY: rebuild
rebuild: clean-all build