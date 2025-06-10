.PHONY: build run test clean install deps lint

# Binary name
BINARY_NAME=mcp-terminal-server
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run the server
run:
	@echo "Running $(BINARY_NAME)..."
	$(GORUN) ./cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	@echo "Coverage report:"
	$(GOCMD) tool cover -func=coverage.out

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./test/integration/... -timeout 30s

# Run all tests including integration
test-all: test test-integration

# Run specific package tests
test-terminal:
	@echo "Running terminal package tests..."
	$(GOTEST) -v ./internal/terminal

test-session:
	@echo "Running session package tests..."
	$(GOTEST) -v ./internal/session

# Build test applications
test-apps:
	@echo "Building test applications..."
	@cd test/apps && $(MAKE) all

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GOCMD) clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run linter
lint:
	@echo "Running linter..."
	$(GOVET) ./...

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Development mode with auto-reload (requires air)
dev:
	@echo "Running in development mode..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

# All: clean, deps, build, test
all: clean deps build test