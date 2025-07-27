# Agent Configuration for Osquery Browser History Extension

## Build/Lint/Test Commands

```bash
# Build the extension
go build -o osquery-browser-history

# Run all tests
go test ./...

# Run a specific test
go test -run TestFunctionName ./...

# Linting (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...
```

## Code Style Guidelines

### Imports
- Group imports in order: standard library, third-party packages, local packages
- Use aliases only when necessary to avoid conflicts
- Remove unused imports

### Formatting
- Use `go fmt` for all code formatting
- Limit line length to 100 characters
- Use tabs for indentation (not spaces)

### Types and Naming
- Use camelCase for variables and functions
- Use PascalCase for exported names
- Use descriptive names over abbreviations
- Prefer descriptive error messages

### Error Handling
- Always handle errors explicitly
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Don't ignore errors with `_` unless explicitly intended

### Testing
- Place tests in same package with `_test` suffix
- Use table-driven tests for multiple test cases
- Name test functions as `TestFunctionName` or `TestStruct_MethodName`

## Agent Rules
- Follow the golang-prototype-architect.md guidelines for new Go projects
- Use security-product-strategist.md for security-related features