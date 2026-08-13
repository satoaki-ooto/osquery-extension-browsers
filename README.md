# Osquery Browser History Extension

An osquery extension that exposes browser profiles, native visit observations, and
current page state across Chromium- and Firefox-based browsers on Windows, macOS,
and Linux.

## Features
- Multi-browser support
  - Chromium family: Chrome, Edge, Chromium, Brave, Vivaldi, Comet
  - Firefox family: Firefox, ESR, Developer Edition, Nightly (Zen on Linux)
- Multi-platform: Windows, macOS (Darwin), Linux
- Multi-profile detection and enumeration
- Deterministic profile and native-visit observation IDs
- Read-only live SQLite access with WAL visibility, a bounded busy timeout, and a
  private snapshot fallback for browser-held locks
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
Then, within osquery, query the three-table visit view:
```sql
SELECT
  v.obs_id,
  v.profile_id,
  v.native_visit_id,
  v.visit_time,
  v.visit_time_us,
  h.url,
  h.title,
  h.title_present,
  p.browser_variant,
  p.os_user_name,
  p.profile_directory,
  p.profile_display_name,
  p.profile_display_name_present,
  p.profile_account,
  p.profile_account_present
FROM browser_visit_observations AS v
JOIN browser_history_pages AS h
  ON h.profile_id = v.profile_id
 AND h.native_url_id = v.native_url_id
JOIN browser_profiles AS p
  ON p.profile_id = v.profile_id
LIMIT 10;
```

For scheduled differential collection, keep the query name and SQL stable, do not
enable snapshot mode, and set `removed` to `false`.

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
