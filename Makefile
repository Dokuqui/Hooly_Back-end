# Define variables
APP_NAME := hooly
SRC := ./cmd/main.go
BUILD_DIR := ./bin
BIN := $(BUILD_DIR)/$(APP_NAME)

# Default target
.DEFAULT_GOAL := run

# Commands
.PHONY: run build lint test deps clean help

# Run the application
run: lint test deps build
	@echo "Running $(APP_NAME)..."
	@$(BIN)

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BIN) $(SRC)
	@echo "Build complete. Binary is located at $(BIN)."

# Lint the code
format:
	@echo "Running lint checks..."
	@go fmt ./... || { echo "Formating failed. Fix issues and try again."; exit 1; }

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v || { echo "Tests failed. Fix issues and try again."; exit 1; }

# Check and install dependencies
deps:
	@echo "Checking dependencies..."
	@go mod tidy
	@go mod verify

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

# Show help
help:
	@echo "Usage:"
	@echo "  make run     - Run the application (includes lint, test, deps, build)"
	@echo "  make build   - Build the application"
	@echo "  make lint    - Run lint checks"
	@echo "  make test    - Run tests"
	@echo "  make deps    - Check dependencies"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make help    - Show this help message"
