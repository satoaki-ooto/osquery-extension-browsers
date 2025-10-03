# Code Style Conventions

## Go Language Conventions
- **Go Version**: 1.22
- **Module Name**: `osquery-extension-browsers`

## Import Organization
1. Standard library packages
2. Third-party packages  
3. Local packages
- Use aliases only when necessary to avoid conflicts
- Remove unused imports

## Naming Conventions
- **Variables/Functions**: camelCase (e.g., `findProfiles`, `historyEntry`)
- **Exported Names**: PascalCase (e.g., `Browser`, `FindProfiles`)
- **Constants**: ALL_CAPS with underscores (following Go conventions)
- **Interfaces**: Descriptive names without "I" prefix (e.g., `Browser` not `IBrowser`)

## Code Formatting
- **Tool**: Use `go fmt` for all formatting
- **Line Length**: Maximum 100 characters
- **Indentation**: Tabs (not spaces) - Go standard
- **Braces**: Opening brace on same line (Go style)

## Error Handling
- Always handle errors explicitly - never ignore with `_`
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Use descriptive error messages
- Return errors as the last return value

## Testing Conventions
- Test files: `*_test.go` in same package
- Test functions: `TestFunctionName` or `TestStruct_MethodName`  
- Use table-driven tests for multiple test cases
- Include both unit tests and integration tests

## Documentation
- Document all exported functions, types, and variables
- Use Go doc comments (start with function/type name)
- Keep comments concise and focused on "why" not "what"

## Package Structure
- Keep packages focused and cohesive
- Use interfaces to define contracts between packages
- Place shared utilities in `common` packages