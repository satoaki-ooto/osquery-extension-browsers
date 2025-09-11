# Development Commands

## Build Commands
```bash
# Build for current platform
make build
# or manually: go build -o browser-extend-extension cmd/browser_extend_extension/main.go

# Build for all supported platforms
make build-all

# Build for specific platforms
make build-linux        # Linux (AMD64 + ARM64)
make build-darwin       # macOS (AMD64 + ARM64)  
make build-windows      # Windows (AMD64)

# Clean build artifacts
make clean
```

## Development Commands
```bash
# Install/update dependencies
make deps
# or: go mod tidy

# Format code
go fmt ./...

# Run all tests
make test
# or: go test -v ./...

# Run tests with coverage
make test-coverage

# Run specific test
go test -run TestFunctionName ./...
```

## Quality Assurance Commands
```bash
# Run linter (requires golangci-lint installation)
make lint
# or: golangci-lint run

# Run linter with auto-fix
make lint-fix
# or: golangci-lint run --fix

# Run all checks (lint + test)
make check
```

## Running the Extension
```bash
# The extension requires osquery to be running and needs socket connection
./browser-extend-extension --socket /path/to/osquery.socket --timeout 3 --interval 3
```

## System Tools (Darwin)
- `go` - /usr/local/go/bin/go (version 1.24.3)
- `git` - /opt/homebrew/bin/git  
- `make` - /usr/bin/make
- `golangci-lint` - Not found (needs installation for linting)

## Note
The project uses standard Go toolchain commands. All Make targets are thin wrappers around Go commands for convenience.