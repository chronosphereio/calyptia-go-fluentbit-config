# Go parameters and flags
GOCMD=go
GOTEST=$(GOCMD) test -race -v -vet=all

# Default target
.PHONY: all
all: fmt build test

.PHONY: build
build:
	$(GOCMD) build -race -v ./...

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: test-coverage
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

.PHONY: clean
clean:
	$(GOCMD) clean
	rm -f coverage.out coverage.html

.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

.PHONY: test-file
test-file:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make test-file FILE=path/to/test.go"; \
		exit 1; \
	fi
	$(GOTEST) $(FILE)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the package (no binary)"
	@echo "  test          - Run all tests (w/ race and vet)"
	@echo "  test-coverage - Run all tests with the coverage report"
	@echo "  test-file     - Run a specific test file (use FILE=path/to/test.go)"
	@echo "  clean         - Clean build and test artifacts"
	@echo "  fmt           - Format the code"
	@echo "  help          - Show this help message"