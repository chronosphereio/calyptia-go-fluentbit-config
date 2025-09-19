# Makefile for go-fluentbit-config

# Variables
BINARY_NAME=fluentbit-config
MAIN_PACKAGE=./cmd/fluentbit-config
MODULE_NAME=github.com/calyptia/go-fluentbit-config/v2

# Go parameters and flags
GOCMD=go
GOTEST=$(GOCMD) test -race -v -vet=all

# Default target
.PHONY: all
all: fmt build test

# Build the binary
.PHONY: build
build:
	$(GOCMD) build -race -v ./...

# Run tests
.PHONY: test
test:
	$(GOTEST) ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCMD) clean
	rm -f coverage.out coverage.html

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Run a specific test file
.PHONY: test-file
test-file:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make test-file FILE=path/to/test.go"; \
		exit 1; \
	fi
	$(GOTEST) $(FILE)

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the package (no binary)"
	@echo "  test          - Run all tests w/ vet"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-file     - Run a specific test file (use FILE=path/to/test.go)"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  help          - Show this help message"