# Research: Periodic Diff Table Implementation

**Feature**: Periodic Diff Table for Browser History
**Date**: 2025-10-22
**Purpose**: Resolve technical decisions for timestamp persistence and diff table implementation

## Research Questions

This document addresses the 5 key technical decisions identified in the implementation plan:

1. Storage location for execution timestamps
2. Timestamp storage schema design
3. Concurrency control mechanism
4. Error handling strategy
5. Integration pattern with existing code

---

## Decision 1: Storage Location for Execution Timestamps

### Question
Should execution timestamps be stored in a dedicated SQLite database or a file-based format (JSON/TOML)?

### Decision
**Use a dedicated SQLite database**

### Rationale

**Chosen approach advantages**:
- **ACID guarantees**: Atomic timestamp updates prevent race conditions when multiple queries execute
- **Queryability**: SQL interface simplifies timestamp lookups by browser type, table name, etc.
- **Consistency with existing code**: Project already uses go-sqlite3 for browser history reading; no new dependencies
- **Restart durability**: Survives osqueryd restarts automatically
- **Concurrent access**: SQLite handles multiple readers and writer serialization via built-in locking
- **Schema enforcement**: Type safety for timestamp values, browser types, table identifiers

**Storage location**: `/tmp/osquery_browser_diff_timestamps.db` (or user-configurable via extension flag)
- Aligns with existing extension log location pattern (`/tmp/browser_extend_extension.log`)
- Separate from browser history databases (read-only constraint preserved)
- Easily cleared for testing/reset scenarios

### Alternatives Considered

**File-based JSON/TOML**:
- ❌ No atomic update guarantees (file write + move required)
- ❌ Manual locking mechanism needed for concurrency
- ❌ Schema validation required at application layer
- ✅ Simpler for human inspection (minor benefit vs operational risk)

**In-memory only (no persistence)**:
- ❌ Fails FR-006: timestamps lost on extension restart
- ❌ Violates core requirement for monitoring continuity

---

## Decision 2: Timestamp Storage Schema Design

### Question
Should we use a single table with a composite key or separate tables per browser type?

### Decision
**Single table with composite key (table_name, browser_type, profile_id)**

### Rationale

**Schema**:
```sql
CREATE TABLE IF NOT EXISTS execution_timestamps (
    table_name TEXT NOT NULL,      -- e.g., 'browser_history_diff'
    browser_type TEXT NOT NULL,    -- 'chrome', 'firefox', 'safari', 'edge'
    profile_id TEXT NOT NULL,      -- Profile identifier from common.Profile
    last_execution_time INTEGER NOT NULL,  -- Unix timestamp in seconds
    PRIMARY KEY (table_name, browser_type, profile_id)
);

CREATE INDEX idx_table_browser ON execution_timestamps(table_name, browser_type);
```

**Advantages**:
- **Scalability**: Adding new browser types requires no schema changes
- **Multi-table support**: FR-008 requires distinguishing `browser_history` vs `browser_history_diff` table timestamps - table_name column enables this
- **Profile isolation**: Each browser profile maintains independent timestamps (FR-005)
- **Simple queries**: Single SELECT with WHERE clause retrieves relevant timestamp
- **Atomic updates**: REPLACE or INSERT OR REPLACE updates timestamp in one operation

**Example usage**:
```go
// Retrieve timestamp for Chrome Default profile in diff table
SELECT last_execution_time
FROM execution_timestamps
WHERE table_name = 'browser_history_diff'
  AND browser_type = 'chrome'
  AND profile_id = 'Default'

// Update timestamp after query completes
REPLACE INTO execution_timestamps (table_name, browser_type, profile_id, last_execution_time)
VALUES ('browser_history_diff', 'chrome', 'Default', 1729638000)
```

### Alternatives Considered

**Separate tables per browser** (`chrome_timestamps`, `firefox_timestamps`, etc.):
- ❌ Schema proliferation (4+ tables)
- ❌ Dynamic table creation complexity
- ❌ Harder to query "all browsers last execution time"

**Flat key-value store** (e.g., `key='chrome:Default:browser_history_diff'`, `value=timestamp`):
- ❌ No schema enforcement
- ❌ Complex key parsing logic
- ❌ Limited query capabilities (e.g., "find all timestamps for Chrome")

---

## Decision 3: Concurrency Control Mechanism

### Question
How do we handle concurrent timestamp reads/writes when multiple osqueryd queries execute simultaneously?

### Decision
**Use SQLite's built-in WAL (Write-Ahead Logging) mode with IMMEDIATE transactions**

### Rationale

**Configuration**:
```go
db, err := sql.Open("sqlite3", "file:/tmp/osquery_browser_diff_timestamps.db?mode=rwc&_journal_mode=WAL")
```

**Transaction pattern**:
```go
// Read timestamp (shared lock, allows concurrent reads)
tx, _ := db.Begin()
row := tx.QueryRow("SELECT last_execution_time FROM execution_timestamps WHERE ...")
timestamp := ...
tx.Commit()

// Write timestamp (exclusive lock via IMMEDIATE)
tx, _ := db.Begin()
tx.Exec("BEGIN IMMEDIATE")  // Acquire write lock early
tx.Exec("REPLACE INTO execution_timestamps ...")
tx.Commit()
```

**Advantages**:
- **Concurrent reads**: Multiple queries can read timestamps simultaneously (WAL mode)
- **Serialized writes**: SQLite automatically serializes timestamp updates
- **No deadlocks**: SQLite's two-lock model (shared + exclusive) prevents deadlocks
- **Crash recovery**: WAL provides durability without fsync on every write
- **No application-level locking**: SQLite handles all concurrency control

**Performance impact**: Sub-millisecond lock contention for timestamp operations (negligible vs browser history query time)

### Alternatives Considered

**Application-level mutex**:
- ❌ Doesn't protect across osqueryd process restarts
- ❌ Complex coordination if multiple extension instances run
- ✅ Simpler code (but SQLite already provides this)

**Lock files**:
- ❌ Manual cleanup required
- ❌ Stale lock detection logic needed
- ❌ Doesn't help with database corruption scenarios

---

## Decision 4: Error Handling Strategy

### Question
Should the system fail-safe (return all entries) or fail-closed (return error) when timestamp storage is unavailable?

### Decision
**Fail-safe with logging: Return all browser history entries when timestamp storage fails**

### Rationale

**Behavior**:
```go
func getLastExecutionTime(store *TimestampStore, browser, profile string) (time.Time, error) {
    timestamp, err := store.Get(browser, profile)
    if err != nil {
        log.Printf("WARNING: Failed to retrieve timestamp for %s/%s: %v - returning zero time (all entries)", browser, profile, err)
        return time.Time{}, nil  // Zero time = include all entries
    }
    return timestamp, nil
}
```

**Advantages**:
- **Monitoring continuity**: FR-006 goal achieved - queries still produce output
- **Operational visibility**: Logs clearly indicate timestamp storage issues
- **Graceful degradation**: System remains functional (though less efficient)
- **Prevents false negatives**: Operators won't miss browser history events due to storage failures
- **Aligns with FR-004**: First-run behavior (no timestamp) returns all entries - same pattern for errors

**Trade-off accepted**: Temporary duplicate entries in logs if storage is intermittently failing (better than missing security events)

**Mitigation**: Extension health checks can monitor timestamp storage and alert on repeated failures

### Alternatives Considered

**Fail-closed (return error)**:
- ❌ Violates monitoring continuity goal
- ❌ Operators lose visibility into browser history during storage outages
- ✅ Prevents duplicate log entries (minor benefit vs operational risk)

**Retry with exponential backoff**:
- ❌ Adds latency to query response (blocks osqueryd)
- ❌ Complex retry logic for potentially unfixable errors (e.g., disk full)
- ✅ Might fix transient issues (but SQLite is usually reliable)

---

## Decision 5: Integration Pattern with Existing Code

### Question
Should the diff table use a decorator pattern (wrap existing history functions) or a separate table implementation?

### Decision
**Separate table implementation with shared history retrieval functions**

### Rationale

**Architecture**:
```
internal/diff_table/
├── table.go              # New table plugin: browser_history_diff
│   ├── RegisterDiffTable()  # Registers table with osqueryd
│   ├── generateDiffHistory() # Callback for table queries
│   └── filterByTimestamp()   # Applies time-based filtering
│
└── Uses existing functions:
    ├── chromium.FindProfiles() + chromium.FindHistory()
    ├── firefox.FindProfiles() + firefox.FindHistory()
    └── timestamp_store.Get() + timestamp_store.Update()
```

**Advantages**:
- **FR-008 compliance**: Separate table maintains independent timestamps from `browser_history` table
- **No breaking changes**: Existing `browser_history` table behavior unchanged
- **Clear separation**: Diff logic isolated in dedicated package
- **Testability**: Can mock timestamp store and browser history functions independently
- **User choice**: Operators query `browser_history` (all entries) or `browser_history_diff` (new only) based on use case

**Code reuse**:
- Diff table calls `chromium.FindHistory()` / `firefox.FindHistory()` directly
- Applies `entry.VisitTime > lastExecutionTime` filter in diff table code
- Updates timestamp after generating results

**Example registration in `main.go`**:
```go
// Existing table
browserHistoryTable := browserHistoryTablePlugin()
server.RegisterPlugin(browserHistoryTable)

// New diff table
browserHistoryDiffTable := diff_table.NewPlugin(timestampStore)
server.RegisterPlugin(browserHistoryDiffTable)
```

### Alternatives Considered

**Decorator pattern** (wrap existing `generateBrowserHistory` function):
- ❌ Harder to maintain independent timestamps (FR-008)
- ❌ Mixes concerns (timestamp logic inside existing table)
- ❌ Breaks separation of concerns

**Modify existing table with a parameter** (e.g., `?diff=true` query constraint):
- ❌ Changes existing table contract (potential breaking change)
- ❌ Harder to implement independent timestamp tracking
- ❌ Query syntax becomes more complex for users

---

## Summary of Decisions

| Decision | Chosen Approach | Key Benefit |
|----------|-----------------|-------------|
| **Storage** | Dedicated SQLite DB at `/tmp/osquery_browser_diff_timestamps.db` | ACID guarantees, consistent with existing patterns |
| **Schema** | Single table with composite key `(table_name, browser_type, profile_id)` | Scalable, multi-table support, profile isolation |
| **Concurrency** | SQLite WAL mode with IMMEDIATE transactions | Built-in concurrent reads, serialized writes, no deadlocks |
| **Error Handling** | Fail-safe: return all entries on storage errors | Maintains monitoring continuity, prevents missed events |
| **Integration** | Separate `browser_history_diff` table reusing existing history functions | FR-008 compliance, no breaking changes, clear separation |

## Best Practices Applied

### osquery Extension Patterns
- **Table plugin registration**: Follow existing `table.NewPlugin()` pattern from main.go
- **Column definitions**: Match existing `table.TextColumn()` usage for timestamp, url, title, etc.
- **Logging**: Use standard Go `log` package consistent with existing extension logging
- **Error handling**: Log warnings for non-fatal errors, return errors only for critical failures

### Go Project Standards
- **Package structure**: Internal packages for timestamp_store and diff_table (not exposed as public API)
- **Testing**: Table-driven tests matching existing `*_test.go` patterns (see `chromium/finder_test.go`)
- **Dependencies**: Reuse existing go-sqlite3 and osquery-go (no new external dependencies)
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` for error context (Go 1.13+ pattern)

### SQLite Best Practices
- **Schema versioning**: Include `PRAGMA user_version` for future schema migrations
- **Indexes**: Index on `(table_name, browser_type)` for common query patterns
- **Constraints**: Use PRIMARY KEY to enforce uniqueness and enable efficient REPLACE operations
- **WAL mode**: Enable for better concurrency (vs default DELETE journal mode)
- **Connection pooling**: Single `*sql.DB` instance shared across table queries (connection reuse)

## Implementation Notes

### Timestamp Precision
- Store timestamps as **Unix seconds** (INTEGER in SQLite)
- Browser history timestamps (microseconds in Chrome, seconds in Firefox) normalized to seconds for comparison
- Existing `common/timestamp.go` utilities may be relevant for normalization

### Testing Strategy
1. **Unit tests** (`timestamp_store/store_test.go`):
   - Get/Update operations
   - Concurrent access (spawn goroutines)
   - Error scenarios (missing DB file, disk full simulation)

2. **Integration tests** (`tests/integration/diff_table_test.go`):
   - Query diff table twice, verify second query excludes first query's entries
   - Simulate osqueryd restart (close/reopen DB)
   - Test with real browser history databases (fixtures)

3. **Table-driven tests**:
   - Multiple browser types (Chrome, Firefox, Safari)
   - Edge cases (zero timestamp, future timestamps)
   - Error paths (storage unavailable)

### Configuration
Add optional extension flags for customization:
```go
--timestamp-db-path   // Override default /tmp location
--timestamp-reset     // Clear all timestamps on startup (testing)
--diff-table-enabled  // Disable diff table if needed (default: true)
```

## References

- **SQLite WAL mode**: https://www.sqlite.org/wal.html
- **osquery-go table plugin**: https://pkg.go.dev/github.com/osquery/osquery-go/plugin/table
- **Go database/sql best practices**: https://go.dev/doc/database/manage-connections
- **Existing codebase patterns**: `internal/browsers/chromium/history.go`, `cmd/browser_extend_extension/main.go`
