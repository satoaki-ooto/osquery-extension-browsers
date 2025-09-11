# Codebase Structure

## Directory Layout
```
osquery-extension-browsers/
├── cmd/
│   └── browser_extend_extension/
│       └── main.go                    # Entry point - osquery extension server
├── internal/
│   ├── browsers/
│   │   ├── common/
│   │   │   ├── interfaces.go          # Browser, Profile, HistoryEntry interfaces
│   │   │   └── detector.go            # Cross-browser detection utilities
│   │   ├── chromium/
│   │   │   ├── finder.go              # Chromium-based browser path detection
│   │   │   ├── history.go             # Chromium history database access
│   │   │   ├── profile.go             # Chromium profile management
│   │   │   └── variants.go            # Chromium browser variants (Chrome, Edge, etc.)
│   │   └── firefox/
│   │       ├── finder.go              # Firefox browser path detection
│   │       ├── history.go             # Firefox history database access
│   │       ├── profile.go             # Firefox profile management
│   │       └── variants.go            # Firefox browser variants
│   └── common/
│       ├── process.go                 # Process detection utilities
│       ├── retry.go                   # Retry logic for operations
│       └── timestamp.go               # Time handling utilities
├── .kiro/specs/                       # Feature specifications and requirements
├── .vscode/                           # VS Code configuration
├── go.mod                             # Go module definition
├── go.sum                             # Go dependency checksums
├── Makefile                           # Build and development commands
└── AGENTS.md                          # Agent configuration and guidelines
```

## Key Components

### Core Interfaces (`internal/browsers/common/interfaces.go`)
- **Browser**: Main interface for browser implementations
- **Profile**: Represents browser profile with metadata
- **HistoryEntry**: Represents single browser history entry

### Platform Support
- **Windows**: Uses APPDATA/LOCALAPPDATA environment variables
- **macOS (Darwin)**: Uses ~/Library/Application Support paths
- **Linux**: Uses ~/.config and ~/.mozilla paths

### Browser Support
#### Chromium-based Browsers
- Google Chrome
- Microsoft Edge
- Chromium
- Brave Browser
- Vivaldi

#### Firefox-based Browsers  
- Firefox
- Firefox ESR
- Firefox Developer Edition
- Firefox Nightly
- Zen Browser (Linux)

### Data Sources
- **Chromium**: History stored in SQLite databases in profile directories
- **Firefox**: History stored in places.sqlite, profiles defined in profiles.ini

## Future Enhancement Areas
- Firefox variant system (similar to Chromium architecture)
- Multi-user browser detection across system users
- Enhanced error handling and logging
- Comprehensive test coverage