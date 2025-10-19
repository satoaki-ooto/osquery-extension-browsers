# Osquery Browser History Extension

An osquery extension that exposes a virtual table to query browser history across Chromium- and Firefox-based browsers on Windows, macOS, and Linux.

## Features
- Multi-browser support
  - Chromium family: Chrome, Edge, Chromium, Brave, Vivaldi, Comet
  - Firefox family: Firefox, ESR, Developer Edition, Nightly (Zen on Linux)
- Multi-platform: Windows, macOS (Darwin), Linux
- Multi-profile detection and enumeration
- Utilities: robust process detection, retry logic, timestamp handling

## Project Layout
- cmd/browser_extend_extension/main.go — extension entrypoint
- internal/browsers/common — interfaces, detector, process, retry, timestamp
- internal/browsers/chromium — finder, history, profile, variants
- internal/browsers/firefox — finder, history, profile, variants
- .kiro/specs — specs for multi-user browser detection

## Build
```bash
# Using Go directly
go build -o osquery-browser-history cmd/browser_extend_extension/main.go

# Or with Makefile
make build
```

## Test
```bash
go test ./...
# Or
make test
```

## Lint & Format
```bash
go fmt ./...
# Requires golangci-lint
golangci-lint run
# Or
make lint
```

## Usage with osquery
The extension must connect to a running osqueryd/osqueryi via a socket.
```bash
./osquery-browser-history --socket /path/to/osquery.socket --timeout 3 --interval 3
```
Then, within osquery:
```sql
SELECT * FROM browser_history LIMIT 10;
```

## Supported Data Sources
- Chromium: SQLite History databases per profile
- Firefox: places.sqlite with profiles defined via profiles.ini

## Development Notes
- Go 1.24.x
- Cross-platform path detection and per-profile enumeration
- See FIREFOX_HISTORY_CHANGES.md for Firefox schema notes

## Contributing
- Run tests and linter before submitting changes
- Follow code style in AGENTS.md and .serena/memories/*

## License
MIT (unless otherwise specified in repository)
