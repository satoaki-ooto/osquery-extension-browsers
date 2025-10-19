# Suggested Commands for Osquery Browser Extension

## Build Commands
```bash
# Build for current platform
go build -o browser-extend-extension cmd/browser_extend_extension/main.go

# Build using Makefile (recommended)
make build

# Build for specific platforms
make build-linux-amd64
make build-linux-arm64
make build-darwin-amd64
make build-darwin-arm64
make build-windows-amd64

# Build for all currently supported platforms
make build-current

# Build for all platforms
make build-all
```

## Development Commands
```bash
# Install/update dependencies
go mod tidy
# OR
make deps

# Format code (required before commits)
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run
# OR
make lint

# Run linter with auto-fix
golangci-lint run --fix
# OR
make lint-fix
```

## Testing Commands
```bash
# Run all tests
go test ./...
# OR
make test

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestFunctionName ./...

# Run tests in specific package
go test ./internal/browsers/firefox/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
# OR
make test-coverage

# Run integration tests
go test -v ./internal/browsers/integration_test.go

# Run benchmark tests
go test -bench=. ./internal/browsers/benchmark_test.go
```

## Quality Assurance
```bash
# Run all checks (lint + test)
make check

# Clean build artifacts
make clean

# Clean cache and build artifacts
make cache-clean
make clean
```

## Usage Commands
```bash
# Run the extension (requires running osquery instance)
./browser-extend-extension --socket /path/to/osquery.socket --timeout 3 --interval 3

# Query browser history in osquery
osqueryi> SELECT * FROM browser_history LIMIT 10;
osqueryi> SELECT browser, profile, url, title, visit_count FROM browser_history WHERE url LIKE '%github%';
```

## System Commands (macOS/Darwin)
```bash
# File operations
ls -la                    # List files with details
find . -name "*.go"       # Find Go files
grep -r "pattern" .       # Search in files

# Git operations
git status                # Check repository status
git diff                  # Show changes
git add .                 # Stage changes
git commit -m "message"   # Commit changes
git log                   # View commit history

# Module operations
go mod tidy               # Clean up dependencies
go mod verify             # Verify dependencies
go mod download           # Download dependencies
```

## Debugging Commands
```bash
# Verbose build output
go build -v -o browser-extend-extension cmd/browser_extend_extension/main.go

# Run with race detector
go test -race ./...

# Check for common mistakes
go vet ./...

# Show package dependencies
go list -m all
```