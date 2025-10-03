# Osquery Browser History Extension - Project Overview

## Purpose
An osquery extension that exposes a virtual table to query browser history across multiple browsers and platforms. This allows system administrators and security analysts to query browser history data using SQL through osquery.

## Tech Stack
- **Language**: Go 1.22
- **Main Dependencies**:
  - `github.com/osquery/osquery-go` - Official osquery Go SDK
  - `github.com/mattn/go-sqlite3` - SQLite database driver
  - `github.com/shirou/gopsutil/v3` - System process utilities
  - `github.com/go-ini/ini` - INI file parsing for Firefox profiles

## Key Features
- **Multi-browser Support**: Chromium family (Chrome, Edge, Chromium, Brave, Vivaldi, Comet) and Firefox family (Firefox, ESR, Developer Edition, Nightly, Zen on Linux)
- **Multi-platform**: Windows, macOS (Darwin), Linux
- **Multi-profile Detection**: Automatically discovers and enumerates browser profiles
- **Robust Architecture**: Includes process detection, retry logic, and timestamp handling utilities

## Main Components
- **Extension Entrypoint**: `cmd/browser_extend_extension/main.go`
- **Common Interfaces**: `internal/browsers/common/` - shared interfaces, detector, process utilities
- **Browser Implementations**: 
  - `internal/browsers/chromium/` - Chromium-based browser support
  - `internal/browsers/firefox/` - Firefox-based browser support
- **Specs**: `.kiro/specs/` - specifications for multi-user browser detection

## Usage
The extension connects to osquery via socket and provides a `browser_history` virtual table that can be queried using standard SQL.