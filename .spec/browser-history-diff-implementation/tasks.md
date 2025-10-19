---
description: "Task list for browser history diff detection implementation"
---

# Tasks: ブラウザ履歴差分検出機能

**Input**: Design documents from `.spec/browser-history-diff-implementation/`
**Prerequisites**: plan.md

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Project structure validation and dependency checks

- [ ] T001 Verify Go dependencies (osquery-go, go-sqlite3) in go.mod
- [ ] T002 [P] Configure development environment and build tools
- [ ] T003 [P] Run existing tests to establish baseline: `go test ./...`

---

## Phase 2: Core Infrastructure (State Management)

**Purpose**: Implement state management system for tracking last fetch times

### Implementation

- [ ] T004 Create `internal/browsers/common/state.go` with StateManager struct
  - JSON serialization/deserialization
  - File path: `/tmp/browser_history_state.json`
  - Thread-safe access with `sync.RWMutex`
  - Methods: `Load()`, `Save()`, `GetLastFetchTime()`, `UpdateLastFetchTime()`

- [ ] T005 Add unit tests for state management in `internal/browsers/common/state_test.go`
  - Test Load/Save operations
  - Test concurrent access scenarios
  - Test file not found handling (first run)
  - Test JSON parsing errors

---

## Phase 3: FindHistory() Extension

**Purpose**: Add filtering capabilities to history retrieval functions

### Data Model Changes

- [ ] T006 Update `internal/browsers/common/interfaces.go`
  - Add `FindHistoryOptions` struct with MinTime, MaxTime, Limit fields
  - Update `Browser` interface to include `FindHistoryWithOptions(profile Profile, opts *FindHistoryOptions) ([]HistoryEntry, error)`
  - Keep existing `FindHistory()` for backward compatibility

### Chromium Implementation

- [ ] T007 Update `internal/browsers/chromium/history.go`
  - Implement `FindHistoryWithOptions()` function
  - Add dynamic SQL query generation based on options
  - Add time range filtering (MinTime/MaxTime)
  - Add LIMIT clause support
  - Keep existing `FindHistory()` calling new function with nil options

- [ ] T008 Add tests for Chromium history filtering in `internal/browsers/chromium/history_test.go`
  - Test with no options (all history)
  - Test with MinTime filter
  - Test with MaxTime filter
  - Test with Limit
  - Test with combined filters

### Firefox Implementation

- [ ] T009 Update `internal/browsers/firefox/history.go`
  - Implement `FindHistoryWithOptions()` function
  - Add dynamic SQL query generation (Firefox uses Unix timestamps directly)
  - Add time range filtering
  - Add LIMIT clause support
  - Keep existing `FindHistory()` calling new function with nil options

- [ ] T010 Add tests for Firefox history filtering in `internal/browsers/firefox/history_test.go`
  - Test with no options (all history)
  - Test with MinTime filter
  - Test with MaxTime filter
  - Test with Limit
  - Test with combined filters

---

## Phase 4: Table Schema Updates

**Purpose**: Add new columns to support change tracking

### Schema Changes

- [ ] T011 Update table schema in `cmd/browser_extend_extension/main.go`
  - Add `unix_time` column (IntegerColumn) to browser_history table
  - Add `change_type` column (TextColumn) to browser_history table
  - Update `generateBrowserHistory()` to populate new columns

---

## Phase 5: Monitoring Table Implementation

**Purpose**: Implement browser_history_mon table for differential updates

### Core Implementation

- [ ] T012 Add `browserHistoryMonTablePlugin()` in `cmd/browser_extend_extension/main.go`
  - Define same schema as browser_history
  - Register new table plugin with name "browser_history_mon"

- [ ] T013 Implement `generateBrowserHistoryMon()` in `cmd/browser_extend_extension/main.go`
  - Initialize StateManager
  - For each browser type and profile:
    - Get last fetch time from state
    - Call FindHistoryWithOptions() with MinTime = last fetch time
    - Determine change_type for each entry (new vs updated based on visit_count)
    - Update state with current timestamp
  - Return results with change_type populated
  - Handle errors gracefully (continue with other profiles)

- [ ] T014 Add change type determination logic
  - Create helper function `determineChangeType(entry HistoryEntry, lastFetchTime time.Time) string`
  - Logic: if lastFetchTime.IsZero() → "new"
  - Logic: if visit_count == 1 → "new"
  - Logic: else → "updated"

---

## Phase 6: Integration & Testing

**Purpose**: Verify complete functionality with integration tests

### Integration Tests

- [ ] T015 Create `internal/browsers/integration_test.go` for monitoring table
  - Test first run (no state file): should return all history
  - Test second run: should return only new entries
  - Test updated entries: revisit URL and verify change_type = "updated"
  - Test multiple profiles: verify state isolation
  - Test concurrent access: multiple queries at same time

- [ ] T016 Test error handling scenarios
  - State file permission errors
  - Browser DB access errors
  - Corrupted state file handling
  - Empty profile lists

### Manual Testing

- [ ] T017 Build and test with osqueryi
  - Run: `go build -o osquery-browser-history cmd/browser_extend_extension/main.go`
  - Test: `SELECT * FROM browser_history;` (full history)
  - Test: `SELECT * FROM browser_history_mon;` (differential)
  - Verify: Run browser_history_mon twice, second run shows only new entries
  - Verify: WHERE clauses work on both tables

---

## Phase 7: Performance & Polish

**Purpose**: Optimize performance and improve code quality

### Performance

- [ ] T018 [P] Add benchmarks in `internal/browsers/benchmark_test.go`
  - Benchmark full history retrieval
  - Benchmark differential retrieval
  - Benchmark state load/save operations
  - Compare performance with/without filters

- [ ] T019 [P] Optimize initial run handling
  - Add limit to first run (prevent overwhelming results)
  - Document recommended Remote Setting interval settings

### Code Quality

- [ ] T020 [P] Add logging improvements
  - Add structured logging for state operations
  - Log metrics (entries found, time taken)
  - Add debug mode logging for troubleshooting

- [ ] T021 [P] Code cleanup and formatting
  - Run `go fmt ./...`
  - Run `golangci-lint run` if available
  - Add code comments for public APIs
  - Ensure consistent error messages

---

## Phase 8: Documentation

**Purpose**: Update documentation for new features

- [ ] T022 [P] Update README.md
  - Add section on browser_history_mon table
  - Add usage examples for Remote Setting
  - Add usage examples for osqueryi
  - Document state file location and format
  - Add troubleshooting guide

- [ ] T023 [P] Add inline documentation
  - Document FindHistoryOptions struct fields
  - Document StateManager methods
  - Add usage examples in code comments

- [ ] T024 [P] Create example configurations
  - Example Remote Setting config for monitoring
  - Example osqueryi queries for forensics
  - Example state file structure

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
- **Core Infrastructure (Phase 2)**: Depends on Setup - CRITICAL for all subsequent phases
- **FindHistory Extension (Phase 3)**: Depends on Phase 2 (needs state management design)
- **Table Schema Updates (Phase 4)**: Can start after Phase 1
- **Monitoring Table (Phase 5)**: Depends on Phase 2, 3, 4 completion
- **Integration & Testing (Phase 6)**: Depends on Phase 5 completion
- **Performance & Polish (Phase 7)**: Can start after Phase 5, should wait for Phase 6
- **Documentation (Phase 8)**: Depends on all implementation phases

### Task Dependencies Within Phases

**Phase 2 (Core Infrastructure)**:
- T005 depends on T004 (test after implementation)

**Phase 3 (FindHistory Extension)**:
- T006 must complete before T007, T009 (interface before implementation)
- T007 and T009 can run in parallel [P] (different files)
- T008 depends on T007 (test after implementation)
- T010 depends on T009 (test after implementation)

**Phase 4 (Table Schema)**:
- T011 is standalone (single file change)

**Phase 5 (Monitoring Table)**:
- T012, T013, T014 must be sequential (same file: main.go)
- T012 → T013 → T014

**Phase 6 (Integration & Testing)**:
- T015 and T016 can run in parallel [P] (different test files)
- T017 depends on all implementation (manual testing)

**Phase 7 (Performance & Polish)**:
- T018, T019, T020, T021 can all run in parallel [P] (different files/concerns)

**Phase 8 (Documentation)**:
- T022, T023, T024 can all run in parallel [P] (different files)

### Critical Path

```
T001-T003 (Setup)
    ↓
T004 (State Manager)
    ↓
T005 (State Tests)
    ↓
T006 (Interface Update)
    ↓
T007, T009 (Chromium & Firefox Implementation) [P]
    ↓
T008, T010 (Tests) [P]
    ↓
T011 (Schema Update)
    ↓
T012 → T013 → T014 (Monitoring Table)
    ↓
T015, T016 (Integration Tests) [P]
    ↓
T017 (Manual Testing)
    ↓
T018-T021 (Performance & Polish) [P]
    ↓
T022-T024 (Documentation) [P]
```

---

## Parallel Execution Examples

### After Phase 2 completion:

```bash
# Chromium and Firefox implementations can proceed in parallel:
Task: "Update internal/browsers/chromium/history.go with FindHistoryWithOptions()"
Task: "Update internal/browsers/firefox/history.go with FindHistoryWithOptions()"
```

### Phase 7 (Performance & Polish):

```bash
# All polish tasks can run in parallel:
Task: "Add benchmarks in internal/browsers/benchmark_test.go"
Task: "Optimize initial run handling"
Task: "Add logging improvements"
Task: "Code cleanup and formatting"
```

### Phase 8 (Documentation):

```bash
# All documentation tasks can run in parallel:
Task: "Update README.md with new features"
Task: "Add inline documentation"
Task: "Create example configurations"
```

---

## Implementation Strategy

### Incremental Development

1. **Phase 1-2**: Build foundation (state management)
   - Validate with unit tests
   - Ensure state persistence works correctly

2. **Phase 3-4**: Extend existing functionality
   - Add filtering to history retrieval
   - Maintain backward compatibility
   - Test both browsers (Chromium + Firefox)

3. **Phase 5**: Implement monitoring table
   - Use components from Phase 2-4
   - Test differential updates
   - Verify state tracking

4. **Phase 6**: Comprehensive testing
   - Integration tests
   - Error scenarios
   - Manual validation

5. **Phase 7-8**: Polish and document
   - Performance optimization
   - User-facing documentation
   - Example configurations

### Testing Strategy

- **Unit tests**: After each implementation task
- **Integration tests**: After Phase 5 completion
- **Manual testing**: Before Phase 7
- **Performance testing**: During Phase 7

### Validation Checkpoints

- **After T005**: State management works correctly
- **After T008, T010**: History filtering works for both browsers
- **After T014**: Monitoring table returns differential results
- **After T017**: Manual testing confirms expected behavior
- **After T021**: Code quality standards met

---

## Notes

- All paths are relative to repository root: `/Users/satoaki-oto/Workspace/repos/github.com/satoaki-ooto/osquery-extension-browsers/`
- [P] indicates tasks that can run in parallel (different files, no dependencies)
- Run tests after each implementation: `go test ./internal/browsers/...`
- Build command: `go build -o osquery-browser-history cmd/browser_extend_extension/main.go`
- State file location: `/tmp/browser_history_state.json`
- Maintain backward compatibility: existing browser_history table behavior unchanged
