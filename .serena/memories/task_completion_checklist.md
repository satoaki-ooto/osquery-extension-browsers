# Task Completion Checklist

## Before Submitting Changes
### Required Steps (in order):
1. **Format Code**: Run `go fmt ./...` - ensures consistent formatting
2. **Run Tests**: Execute `go test ./...` - verify all tests pass
3. **Run Linter**: Execute `golangci-lint run` - check code quality and style
4. **Build Verification**: Run `make build` - ensure code compiles successfully

### Optional Quality Checks:
- **Test Coverage**: Run `make test-coverage` for coverage analysis
- **Cross-platform Build**: Run `make build-all` for multi-platform verification  
- **Clean Build**: Run `make clean && make build` for fresh build verification

## Code Review Guidelines
- Verify error handling is explicit and descriptive
- Check that interfaces are properly implemented
- Ensure tests are added for new functionality
- Confirm documentation is updated for exported functions
- Validate that naming follows Go conventions

## Git Workflow
1. Create feature branch from main
2. Make changes following code style conventions
3. Run complete checklist above
4. Commit with descriptive message
5. Create pull request with clear description

## Security Considerations
- Review browser data access patterns
- Ensure proper file permission handling
- Validate input sanitization for SQL queries
- Check for potential information disclosure

## Platform-specific Testing
- Test on target operating systems (Windows, macOS, Linux)
- Verify browser detection works across different versions
- Validate profile enumeration on multi-user systems

## Integration Testing
- Test with actual osquery instance
- Verify socket communication
- Validate SQL query results match expected browser history data