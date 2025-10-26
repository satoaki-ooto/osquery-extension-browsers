# Implementation Plan: Periodic Diff Table for Browser History

**Branch**: `001-periodic-diff-table` | **Date**: 2025-10-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-periodic-diff-table/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a periodic diff table mechanism for the osquery browser history extension to address the problem of missing logs when no browser history changes occur. The solution involves:

1. **Execution timestamp tracking** - Persistent storage of last query execution times per browser/table
2. **Time-based filtering** - Filter browser history entries to return only new entries since last query
3. **Dedicated table interface** - New `browser_history_diff` table with automatic timestamp management

This enables continuous monitoring visibility and compliance reporting by ensuring regular output even when browser history is unchanged.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: osquery-go v0.0.0-20250131154556-629f995b6947, mattn/go-sqlite3 v1.14.29
**Storage**: SQLite database for execution timestamp persistence
**Testing**: Go standard testing (`go test`), table-driven test patterns
**Target Platform**: macOS (Darwin), with potential Linux support
**Project Type**: Single project (osquery extension)
**Performance Goals**: Sub-100ms query response time for timestamp lookups, handle 10k+ browser history entries efficiently
**Constraints**: Must not interfere with existing browser_history table, atomic timestamp updates required, no external service dependencies
**Scale/Scope**: Single osquery extension with 1 new table, 1 new storage component, affects ~4 browser types (Chrome, Firefox, Safari, Edge)

### Key Technical Decisions Requiring Research

1. **NEEDS CLARIFICATION**: Storage location for execution timestamps (dedicated SQLite DB vs file-based JSON/TOML)
2. **NEEDS CLARIFICATION**: Timestamp storage schema design (single table vs per-browser tables)
3. **NEEDS CLARIFICATION**: Concurrency control mechanism for timestamp reads/writes
4. **NEEDS CLARIFICATION**: Error handling strategy when timestamp storage is unavailable (fail-safe vs fail-closed)
5. **NEEDS CLARIFICATION**: Integration pattern with existing history retrieval code (decorator vs separate table implementation)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Status**: No constitution file found (template placeholder exists). This project does not have formal constitution constraints defined yet.

**Assumed Best Practices Applied**:
- Follow existing project Go style and patterns (observed from codebase)
- Maintain test coverage for new components
- Use table-driven tests consistent with existing `*_test.go` files
- Preserve existing extension architecture (plugin pattern)
- No breaking changes to existing `browser_history` table

**Gate Status**: ✅ PASS (no violations possible without constitution)

## Project Structure

### Documentation (this feature)

```
specs/001-periodic-diff-table/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (pending)
├── data-model.md        # Phase 1 output (pending)
├── quickstart.md        # Phase 1 output (pending)
├── contracts/           # Phase 1 output (pending)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Existing structure (preserved)
cmd/
└── browser_extend_extension/
    └── main.go                          # Extension entry point (will add new table registration)

internal/
└── browsers/
    ├── chromium/
    │   ├── history.go                   # Existing Chromium history retrieval
    │   └── ...
    ├── firefox/
    │   ├── history.go                   # Existing Firefox history retrieval
    │   └── ...
    ├── common/
    │   ├── interfaces.go                # Existing common interfaces
    │   ├── timestamp.go                 # Existing timestamp utilities
    │   └── ...
    └── ...

# New structure (to be added)
internal/
└── timestamp_store/                     # NEW: Execution timestamp persistence
    ├── store.go                         # Core storage interface and implementation
    ├── store_test.go                    # Unit tests for storage
    └── schema.sql                       # SQLite schema definition

internal/
└── diff_table/                          # NEW: Periodic diff table implementation
    ├── table.go                         # Table plugin implementation
    ├── table_test.go                    # Table generation tests
    └── filter.go                        # Time-based filtering logic

tests/
├── integration/                         # NEW: End-to-end table query tests
│   └── diff_table_test.go
└── ...
```

**Structure Decision**:

This is a single-project Go extension following the existing architecture. New components are added as internal packages:

- `internal/timestamp_store/`: Encapsulates all timestamp persistence logic, providing clean interface for the diff table
- `internal/diff_table/`: Implements the new osquery table plugin, depends on timestamp_store and existing browser history retrieval code
- `cmd/browser_extend_extension/main.go`: Updated to register the new table plugin alongside existing browser_history table

This structure maintains separation of concerns (storage vs table logic) while preserving the existing extension architecture.

## Complexity Tracking

*No constitution violations to justify - constitution not defined*

| Decision | Justification |
|----------|---------------|
| New SQLite database for timestamps | Avoids modifying browser history databases (read-only), provides ACID guarantees for timestamp updates, survives extension restarts |
| Separate internal packages | Maintains testability and separation of storage vs presentation logic |
