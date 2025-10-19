# Code Style Conventions

## Go Language Conventions
- **Go Version**: 1.22
- **Module Name**: `osquery-extension-browsers`
- **Binary Name**: `browser-extend-extension`

## Import Organization
1. Standard library packages (e.g., `fmt`, `os`, `path/filepath`)
2. Third-party packages (e.g., `github.com/osquery/osquery-go`)
3. Local packages (e.g., `osquery-extension-browsers/internal/browsers/common`)
- Use aliases only when necessary to avoid conflicts
- Remove unused imports
- Group imports with blank lines between categories

## Naming Conventions
- **Variables/Functions**: camelCase (e.g., `findProfiles`, `historyEntry`, `browserPath`)
- **Exported Names**: PascalCase (e.g., `Browser`, `FindProfiles`, `HistoryEntry`)
- **Constants**: ALL_CAPS with underscores or PascalCase for exported constants
- **Interfaces**: Descriptive names without "I" prefix (e.g., `Browser` not `IBrowser`)
- **Package names**: Short, lowercase, single-word names (e.g., `chromium`, `firefox`, `common`)

## Code Formatting
- **Tool**: Use `go fmt` for all formatting (run before every commit)
- **Line Length**: Maximum 100 characters (soft limit, can be exceeded for readability)
- **Indentation**: Tabs (not spaces) - Go standard
- **Braces**: Opening brace on same line (K&R style)
- **Blank lines**: Use sparingly for logical separation

## Error Handling
- Always handle errors explicitly - never ignore with `_` unless justified
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Use descriptive error messages that include operation context
- Return errors as the last return value
- Prefer early returns for error conditions

## Testing Conventions
- **Test files**: `*_test.go` in same package as code under test
- **Test functions**: `TestFunctionName` or `TestStruct_MethodName`
- **Use table-driven tests** for multiple test cases with similar structure
- **Test types**: Include unit tests, integration tests, and benchmarks
- **Coverage**: Aim for high coverage on critical paths
- **Mocking**: Use interfaces to enable testability

## Documentation
- Document all exported functions, types, and variables
- Use Go doc comments (start with function/type name)
- Keep comments concise and focused on "why" not "what"
- Format: `// FunctionName does something...`
- Update documentation when changing behavior

## Package Structure
- Keep packages focused and cohesive
- Use interfaces to define contracts between packages
- Place shared utilities in `common` packages
- Separate platform-specific code clearly
- Minimize package dependencies

## Security Best Practices
- Sanitize file paths before accessing browser databases
- Handle concurrent access to browser files safely
- Validate input from external sources
- Use appropriate file permissions
- Never log sensitive browser history data in production

## Performance Considerations
- Use worker pools for concurrent processing
- Close database connections and file handles properly
- Implement retry logic for transient failures
- Use defer for cleanup operations
- Profile code for performance bottlenecks