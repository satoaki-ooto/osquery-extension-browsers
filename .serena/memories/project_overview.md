# Osquery Browser History Extension

## Project Purpose
An osquery extension exposing a browser_history virtual table to query browsing history across Chromium-based and Firefox-based browsers and profiles on Windows, macOS, and Linux.

## Key Features
- Multi-browser support: Chromium (Chrome, Edge, Chromium, Brave, Vivaldi) and Firefox (Firefox, ESR, Dev Edition, Nightly; Zen on Linux)
- Multi-platform: Windows, macOS (Darwin), Linux
- Multi-profile detection per browser
- Robust utilities: process detection, retry logic, timestamp conversion

## Tech Stack
- Language: Go (per go.mod)
- Notable packages: osquery-go, gopsutil, go-sqlite3, go-ini

## Architecture
- Entry: cmd/browser_extend_extension/main.go
- Abstractions: internal/browsers/common/interfaces.go
- Implementations: internal/browsers/chromium/* and internal/browsers/firefox/*
- Utilities: internal/browsers/common/*

## Status
- Core Chromium and Firefox history retrieval implemented
- Tests present: unit and integration tests under internal/browsers/*_test.go
- Specs under .kiro/specs for multi-user browser detection

## Commands
- Build: go build -o osquery-browser-history or use Makefile targets
- Test: go test ./...
- Lint: golangci-lint run (if installed)