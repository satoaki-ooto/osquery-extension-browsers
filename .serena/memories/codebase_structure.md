# Codebase Structure (updated)

- cmd/browser_extend_extension/main.go: extension entrypoint
- internal/browsers/common: interfaces, detector, process, retry, timestamp
- internal/browsers/chromium: finder, history, profile, variants (+ tests)
- internal/browsers/firefox: finder, history, profile, variants (+ tests)
- internal/browsers/*_test.go: unit/integration/benchmark/worker pool tests
- .kiro/specs/multi-user-browser-detection: design, requirements, tasks
- Makefile: build/test/lint helpers
- AGENTS.md: agent configuration
- FIREFOX_HISTORY_CHANGES.md: notes on Firefox history schema changes