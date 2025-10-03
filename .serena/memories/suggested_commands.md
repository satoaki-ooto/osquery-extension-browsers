# Suggested Commands for Osquery Browser Extension

## Build Commands
```bash
# Build for current platform
go build -o osquery-browser-history cmd/browser_extend_extension/main.go

# Build using Makefile
make build

# Build for specific platform
make build-linux-amd64
make build-darwin-arm64  
make build-windows-amd64

# Build for all supported platforms
make build-all
```

## Development Commands
```bash
# Install/update dependencies
go mod tidy

# Format code (required before commits)
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Run linter with auto-fix
golangci-lint run --fix
# OR using Makefile
make lint
make lint-fix
```

## Testing Commands
```bash
# Run all tests
go test ./...
make test

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestFunctionName ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
# OR using Makefile
make test-coverage
```

## Quality Assurance
```bash
# Run all checks (lint + test)
make check

# Clean build artifacts  
make clean
```

## Usage Commands
```bash
# Run the extension (requires running osquery instance)
./osquery-browser-history --socket /path/to/osquery.socket --timeout 3 --interval 3

# Query browser history in osquery
osqueryi> SELECT * FROM browser_history LIMIT 10;
```

## System Commands (macOS/Darwin)
```bash
# File operations
ls -la          # List files with details
find . -name "*.go"  # Find Go files
grep -r "pattern" .  # Search in files

# Git operations
git status      # Check repository status
git diff        # Show changes
git add .       # Stage changes
git commit -m "message"  # Commit changes
```