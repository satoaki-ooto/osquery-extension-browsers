# Osquery Browser History Extension - Project Overview

## Purpose
An osquery extension that exposes a virtual table to query browser history across multiple browsers and platforms. This allows system administrators and security analysts to query browser history data using SQL through osquery.

## Tech Stack
- **Language**: Go 1.22
- **Main Dependencies**:
  - `github.com/osquery/osquery-go` v0.0.0-20250131154556-629f995b6947 - Official osquery Go SDK
  - `github.com/mattn/go-sqlite3` v1.14.29 - SQLite database driver for reading browser databases
  - `github.com/shirou/gopsutil/v3` v3.24.5 - System process utilities for browser detection
  - `github.com/go-ini/ini` v1.67.0 - INI file parsing for Firefox profiles.ini

## Key Features
- **Multi-browser Support**: 
  - Chromium family: Chrome, Edge, Chromium, Brave, Vivaldi, Comet
  - Firefox family: Firefox, ESR, Developer Edition, Nightly (Zen on Linux)
- **Multi-platform**: Windows, macOS (Darwin), Linux
- **Multi-profile Detection**: Automatically discovers and enumerates browser profiles per user
- **Robust Architecture**: Includes process detection, retry logic, timestamp handling utilities, and worker pool for concurrent processing

## Main Components
- **Extension Entrypoint**: `cmd/browser_extend_extension/main.go`
- **Common Interfaces**: `internal/browsers/common/` - shared interfaces, detector, process utilities, retry logic, timestamp handling
- **Browser Implementations**: 
  - `internal/browsers/chromium/` - Chromium-based browser support (finder, history, profile, variants)
  - `internal/browsers/firefox/` - Firefox-based browser support (finder, history, profile, variants)
- **Testing**: Comprehensive unit tests, integration tests, benchmark tests, and worker pool tests

## Usage
The extension connects to osquery via socket and provides a `browser_history` virtual table that can be queried using standard SQL:
```bash
./browser-extend-extension --socket /path/to/osquery.socket --timeout 3 --interval 3
```

Then query within osquery:
```sql
SELECT * FROM browser_history LIMIT 10;
```

## Documentation
- `README.md` - Main project documentation
- `FIREFOX_HISTORY_CHANGES.md` - Firefox schema notes
- `EXTENSION_STARTUP_ISSUE_ANALYSIS.md` - Troubleshooting guide
- `AGENTS.md` / `CLAUDE.md` - Agent configuration and code style guidelines
- `CRUSH.md` - Additional project context
- `.serena/memories/` - Onboarding documentation
- `.specify/` - Specify agent templates and constitution