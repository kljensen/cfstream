# Justfile for cfstream

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
    mkdir -p bin
    go build -o bin/cfstream .

# Install to specified directory (e.g., just install ~/bin)
install INSTALL_DIR: build
    cp bin/cfstream {{INSTALL_DIR}}/cfstream
    @echo "✓ Installed to {{INSTALL_DIR}}"

# Run the application
run *args:
    go run ./cmd/cfstream {{args}}

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
