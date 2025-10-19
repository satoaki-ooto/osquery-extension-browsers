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
│       │   ├── detector_test.go    # Detector tests
│       │   ├── process.go          # Process utilities
│       │   ├── retry.go            # Retry mechanisms
│       │   └── timestamp.go        # Timestamp handling
│       ├── chromium/               # Chromium-based browsers
│       │   ├── finder.go           # Browser finder implementation
│       │   ├── finder_test.go      # Finder tests
│       │   ├── history.go          # History reading logic
│       │   ├── profile.go          # Profile management
│       │   └── variants.go         # Browser variants (Chrome, Edge, Brave, Vivaldi, Comet, etc.)
│       ├── firefox/                # Firefox-based browsers
│       │   ├── finder.go           # Browser finder implementation
│       │   ├── finder_test.go      # Finder tests
│       │   ├── history.go          # History reading logic
│       │   ├── history_test.go     # History tests
│       │   ├── profile.go          # Profile management
│       │   └── variants.go         # Browser variants (Firefox, ESR, Developer, Nightly, Zen)
│       ├── benchmark_test.go       # Performance benchmarks
│       ├── integration_test.go     # Integration tests
│       └── worker_pool_test.go     # Worker pool tests
├── .serena/                        # Serena MCP configuration
│   ├── memories/                   # Project onboarding memories
│   ├── .gitignore
│   └── project.yml
├── .specify/                       # Specify agent configuration
│   ├── memory/
│   │   └── constitution.md
│   ├── scripts/bash/               # Helper scripts
│   └── templates/                  # Agent and spec templates
├── .kilocode/                      # Kilocode configuration
│   └── mcp.json
├── .roo/                           # Roo configuration
│   └── mcp.json
├── .vscode/                        # VS Code settings
├── .gitignore
├── go.mod                          # Go module dependencies
├── go.sum                          # Go dependency checksums
├── Makefile                        # Build automation
├── README.md                       # Main documentation
├── AGENTS.md                       # Agent configuration
├── CLAUDE.md                       # Agent configuration (duplicate)
├── CRUSH.md                        # Project context
├── crush.json                      # Crush configuration
├── opencode.json                   # OpenCode configuration
├── FIREFOX_HISTORY_CHANGES.md      # Firefox-specific notes
└── EXTENSION_STARTUP_ISSUE_ANALYSIS.md  # Troubleshooting guide
```

## Key Files
- **main.go**: Contains `main()`, `browserHistoryTablePlugin()`, and `generateBrowserHistory()` functions
- **interfaces.go**: Defines core `Browser`, `Profile`, and `HistoryEntry` types
- **Makefile**: Build targets for multiple platforms (Linux, macOS, Windows) and architectures (AMD64, ARM64)
- **Test files**: Each package includes comprehensive tests with `_test.go` suffix

## Architecture Pattern
- **Interface-based design**: `Browser` interface implemented by Chromium and Firefox packages
- **Common utilities**: `internal/browsers/common/` shared across implementations
- **Clear separation**: Browser-specific logic separated from shared functionality
- **Concurrent processing**: Worker pool pattern for parallel browser history processing
- **Cross-platform support**: Platform-specific path detection and profile enumeration