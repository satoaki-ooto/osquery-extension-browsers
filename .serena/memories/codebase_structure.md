# Codebase Structure

## Directory Layout
```
├── cmd/
│   └── browser_extend_extension/
│       └── main.go                 # Extension entrypoint
├── internal/
│   └── browsers/
│       ├── common/                 # Shared interfaces and utilities
│       │   ├── interfaces.go       # Browser, Profile, HistoryEntry interfaces
│       │   ├── detector.go         # Browser detection logic
│       │   ├── process.go          # Process utilities
│       │   ├── retry.go            # Retry mechanisms
│       │   └── timestamp.go        # Timestamp handling
│       ├── chromium/               # Chromium-based browsers
│       │   ├── finder.go           # Browser finder implementation
│       │   ├── history.go          # History reading logic
│       │   ├── profile.go          # Profile management
│       │   └── variants.go         # Browser variants (Chrome, Edge, etc.)
│       └── firefox/                # Firefox-based browsers
│           ├── finder.go           # Browser finder implementation
│           ├── history.go          # History reading logic
│           ├── profile.go          # Profile management
│           └── variants.go         # Browser variants (Firefox, ESR, etc.)
├── .kiro/specs/                    # Feature specifications
├── .serena/                        # Serena MCP configuration
├── .specify/                       # Specify agent configuration
└── .vscode/                        # VS Code settings
```

## Key Files
- **main.go**: Contains `main()`, `browserHistoryTablePlugin()`, and `generateBrowserHistory()` functions
- **interfaces.go**: Defines core `Browser`, `Profile`, and `HistoryEntry` types
- **Test files**: Each package includes comprehensive tests with `_test.go` suffix

## Architecture Pattern
- Interface-based design with `Browser` interface implemented by Chromium and Firefox packages
- Common utilities in `internal/browsers/common/` shared across implementations
- Clear separation between browser-specific logic and shared functionality