# Code Style and Conventions

## Go Conventions
Following standard Go conventions as outlined in AGENTS.md:

### Import Organization
1. **Standard library packages** (e.g., `fmt`, `os`, `path/filepath`)
2. **Third-party packages** (e.g., `github.com/osquery/osquery-go`)
3. **Local packages** (e.g., `osquery-extension-browsers/internal/...`)

### Naming Conventions
- **Variables and functions**: camelCase (e.g., `browserHistoryTable`, `generateBrowserHistory`)
- **Exported names**: PascalCase (e.g., `FindProfiles`, `HistoryEntry`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Interfaces**: Often end with -er suffix or descriptive names (e.g., `Browser`)

### Code Organization
- **Package structure**: Domain-driven with `internal/browsers/{browser-type}/` structure
- **Interfaces**: Defined in `internal/browsers/common/interfaces.go` for cross-browser abstraction
- **Error handling**: Always explicit, with context wrapping using `fmt.Errorf("context: %w", err)`

### Formatting Rules
- **Line length**: 100 characters maximum
- **Indentation**: Tabs (not spaces)
- **Formatting**: Always use `go fmt ./...`
- **Imports**: Remove unused imports, use aliases only when necessary

### Comments and Documentation
- **NO COMMENTS**: The project explicitly avoids adding comments unless requested
- **Function names**: Should be descriptive enough to be self-documenting
- **Variable names**: Prefer descriptive names over abbreviations

### Testing Conventions
- **Test files**: `*_test.go` suffix in same package
- **Test functions**: `TestFunctionName` or `TestStruct_MethodName`
- **Table-driven tests**: Preferred for multiple test cases
- **Current status**: No tests present in codebase yet

### Platform-Specific Code
- Uses `runtime.GOOS` for platform detection
- Platform-specific paths organized in separate case statements
- Supports Windows, Darwin (macOS), and Linux