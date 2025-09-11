# Osquery Browser History Extension

## Project Purpose
This project is an osquery extension that provides browser history data collection capabilities. It creates a custom table called `browser_history` that can be queried through osquery to retrieve browsing history from multiple browsers and profiles across the system.

## Key Features
- **Multi-browser support**: Supports both Chromium-based browsers (Chrome, Edge, Chromium, Brave, Vivaldi) and Firefox-based browsers (Firefox, Firefox ESR, Firefox Developer Edition, Firefox Nightly)
- **Multi-platform**: Works on Windows, macOS (Darwin), and Linux
- **Profile-aware**: Detects and processes multiple browser profiles per browser
- **Cross-platform architecture**: Uses Go's cross-compilation capabilities for multiple OS/architecture combinations

## Tech Stack
- **Language**: Go 1.24.3
- **Main Dependencies**:
  - `github.com/osquery/osquery-go` - Core osquery extension framework
  - `github.com/shirou/gopsutil/v3` - System information gathering
  - `github.com/mattn/go-sqlite3` - SQLite database interaction for browser history
  - `github.com/go-ini/ini` - INI file parsing for Firefox profiles

## Architecture
- **Entry point**: `cmd/browser_extend_extension/main.go` - Creates osquery extension server
- **Browser abstraction**: `internal/browsers/common/interfaces.go` - Common interfaces for all browsers
- **Browser implementations**: Separate packages for Chromium and Firefox with platform-specific path detection
- **Utilities**: Common utilities for process detection, retry logic, and timestamp handling

## Current Development Status
- Core functionality implemented for both Chromium and Firefox browsers
- Planning phase includes Firefox variant system enhancement and multi-user browser detection
- No tests currently present in the codebase