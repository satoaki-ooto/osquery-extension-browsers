# Task Completion Checklist

## Required Steps After Implementing Changes

### 1. Code Quality Checks
```bash
# Format code (ALWAYS required)
go fmt ./...

# Run linter if available (install golangci-lint if not present)
make lint
# or: golangci-lint run
```

### 2. Testing
```bash
# Run all tests
make test
# or: go test -v ./...

# Run tests with coverage for comprehensive changes
make test-coverage
```

### 3. Build Verification
```bash
# Test build for current platform
make build

# For cross-platform changes, test all platforms
make build-all
```

### 4. Dependency Management
```bash
# Update dependencies if new imports were added
go mod tidy
```

### 5. Final Verification
```bash
# Run all checks together
make check  # This runs both lint and test
```

## Special Considerations

### For New Dependencies
- Always verify the dependency is appropriate and secure
- Run `go mod tidy` after adding imports
- Check if similar functionality already exists in the codebase

### For Platform-Specific Code
- Test on multiple platforms if possible
- Ensure all OS cases are handled in switch statements
- Verify file paths use proper separators

### For Browser-Related Changes
- Test with multiple browser variants
- Ensure profile detection works correctly
- Validate SQLite database access patterns

### For Interface Changes
- Ensure all implementations are updated
- Check for breaking changes in the common interfaces
- Verify backward compatibility

## Installation Requirements
Note: `golangci-lint` is not currently installed on the system. Install it for full linting capability:
```bash
# Install golangci-lint for comprehensive linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```