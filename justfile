# Justfile for kyletube

set dotenv-load := true

mod devcontainer

# Default target - show available commands
default:
    @just --list

# Run tests
test *args:
    go test {{args}} ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Build the application
build:
    go build -o bin/kyletube ./cmd/kyletube

# Run the application
run *args:
    go run ./cmd/kyletube {{args}}

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run

# Run all checks (tests, fmt, vet)
check:
    go fmt ./...
    go vet ./...
    go test ./...

# Install dependencies
deps:
    go mod download
    go mod tidy

# Clean build artifacts
clean:
    rm -rf bin/
    go clean
