# Quickstart: Periodic Diff Table

**Feature**: Periodic Diff Table for Browser History
**Audience**: Developers implementing the feature
**Estimated time**: 30 minutes to understand, 4-6 hours to implement

## Overview

This guide provides a step-by-step walkthrough for implementing the periodic diff table feature, which adds time-based filtering to the osquery browser history extension. By the end, you'll have a working `browser_history_diff` table that returns only new browser history entries since the last query.

## Prerequisites

- Go 1.22+ installed
- osquery or osqueryd running locally
- Familiarity with the existing codebase (`internal/browsers/`, `cmd/browser_extend_extension/`)
- SQLite3 command-line tool (for debugging)

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│ osqueryd: SELECT * FROM browser_history_diff;                │
└────────────────────┬─────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────┐
│ diff_table.generateDiffHistory()                             │
│  ├─ 1. Discover browser profiles                            │
│  ├─ 2. For each browser/profile:                            │
│  │     ├─ Get last execution timestamp                      │
│  │     ├─ Query browser history DB                          │
│  │     └─ Filter entries (visit_time > timestamp)           │
│  ├─ 3. Batch update timestamps                              │
│  └─ 4. Return filtered entries                              │
└────────────────────┬─────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────┐
│ timestamp_store (SQLite)                                     │
│  /tmp/osquery_browser_diff_timestamps.db                     │
│  ├─ execution_timestamps table                               │
│  └─ Stores: (table_name, browser_type, profile_id) → timestamp │
└──────────────────────────────────────────────────────────────┘
```

## Implementation Steps

### Step 1: Create Timestamp Store Package (2 hours)

**File**: `internal/timestamp_store/store.go`

1. **Define the interface** (copy from `contracts/timestamp_store.go`):
   ```go
   type TimestampStore interface {
       Get(tableName, browserType, profileID string) (time.Time, error)
       Update(tableName, browserType, profileID string, timestamp time.Time) error
       BatchUpdate(updates []TimestampUpdate) error
       Close() error
   }
   ```

2. **Implement SQLite storage**:
   ```go
   type sqliteStore struct {
       db *sql.DB
   }

   func NewSQLiteStore(dbPath string) (TimestampStore, error) {
       db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL", dbPath))
       if err != nil {
           return nil, fmt.Errorf("open database: %w", err)
       }

       // Initialize schema
       if err := initSchema(db); err != nil {
           db.Close()
           return nil, fmt.Errorf("init schema: %w", err)
       }

       return &sqliteStore{db: db}, nil
   }
   ```

3. **Initialize schema** (see `data-model.md` for SQL):
   - Create `execution_timestamps` table
   - Add indexes and validation triggers
   - Set `PRAGMA user_version = 1`

4. **Implement Get() method**:
   ```go
   func (s *sqliteStore) Get(tableName, browserType, profileID string) (time.Time, error) {
       var unixTime int64
       err := s.db.QueryRow(`
           SELECT last_execution_time
           FROM execution_timestamps
           WHERE table_name=? AND browser_type=? AND profile_id=?
       `, tableName, browserType, profileID).Scan(&unixTime)

       if err == sql.ErrNoRows {
           return time.Time{}, nil // Zero time = first query
       }
       if err != nil {
           return time.Time{}, fmt.Errorf("query timestamp: %w", err)
       }

       return time.Unix(unixTime, 0), nil
   }
   ```

5. **Implement BatchUpdate() method** (see `data-model.md` for transaction pattern)

**Testing** (`internal/timestamp_store/store_test.go`):
- Test Get() with no record → returns zero time
- Test Update() → subsequent Get() returns updated time
- Test BatchUpdate() → all timestamps updated atomically
- Test concurrent reads (spawn 10 goroutines calling Get())

---

### Step 2: Create Diff Table Package (2 hours)

**File**: `internal/diff_table/table.go`

1. **Define NewPlugin() function**:
   ```go
   func NewPlugin(store timestamp_store.TimestampStore) *table.Plugin {
       columns := []table.ColumnDefinition{
           table.TextColumn("time"),
           table.TextColumn("url"),
           table.TextColumn("title"),
           table.TextColumn("visit_count"),
           table.TextColumn("profile"),
           table.TextColumn("browser_type"),
           table.TextColumn("browser_variant"),
       }

       return table.NewPlugin("browser_history_diff", columns, func(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
           return generateDiffHistory(ctx, queryContext, store)
       })
   }
   ```

2. **Implement generateDiffHistory()** (core logic):
   ```go
   func generateDiffHistory(ctx context.Context, queryContext table.QueryContext, store timestamp_store.TimestampStore) ([]map[string]string, error) {
       var results []map[string]string
       var timestampUpdates []timestamp_store.TimestampUpdate

       // 1. Discover and query Chromium profiles
       chromiumProfiles, _ := chromium.FindProfiles()
       for _, profile := range chromiumProfiles {
           entries, err := processProfile(ctx, store, profile, "chrome")
           if err != nil {
               log.Printf("Error processing Chrome profile %s: %v", profile.ID, err)
               continue
           }
           results = append(results, entries...)
           timestampUpdates = append(timestampUpdates, timestamp_store.TimestampUpdate{
               TableName:   "browser_history_diff",
               BrowserType: "chrome",
               ProfileID:   profile.ID,
               Timestamp:   time.Now(),
           })
       }

       // 2. Discover and query Firefox profiles (similar pattern)
       // ...

       // 3. Batch update timestamps
       if err := store.BatchUpdate(timestampUpdates); err != nil {
           log.Printf("WARNING: Failed to update timestamps: %v", err)
           // Don't return error - fail-safe approach
       }

       return results, nil
   }
   ```

3. **Implement processProfile() helper**:
   ```go
   func processProfile(ctx context.Context, store timestamp_store.TimestampStore, profile common.Profile, browserType string) ([]map[string]string, error) {
       // Get last execution timestamp
       lastExec, err := store.Get("browser_history_diff", browserType, profile.ID)
       if err != nil {
           log.Printf("WARNING: Failed to get timestamp for %s/%s: %v - using zero time", browserType, profile.ID, err)
           lastExec = time.Time{} // Fail-safe: treat as first query
       }

       // Query browser history
       var historyEntries []common.HistoryEntry
       switch browserType {
       case "chrome":
           historyEntries, err = chromium.FindHistory(profile)
       case "firefox":
           historyEntries, err = firefox.FindHistory(profile)
       // ... other browsers
       }

       if err != nil {
           return nil, fmt.Errorf("find history: %w", err)
       }

       // Filter entries
       filtered := filterEntries(historyEntries, lastExec)

       // Format for osquery
       var results []map[string]string
       for _, entry := range filtered {
           results = append(results, map[string]string{
               "time":            entry.VisitTime.Format("2006-01-02 15:04:05"),
               "url":             entry.URL,
               "title":           entry.Title,
               "visit_count":     strconv.Itoa(entry.VisitCount),
               "profile":         entry.ProfileID,
               "browser_type":    entry.BrowserType,
               "browser_variant": entry.BrowserVariant,
           })
       }

       return results, nil
   }
   ```

**File**: `internal/diff_table/filter.go`

4. **Implement filterEntries()**:
   ```go
   func filterEntries(entries []common.HistoryEntry, lastExecutionTime time.Time) []common.HistoryEntry {
       if lastExecutionTime.IsZero() {
           return entries // First query: return all
       }

       var filtered []common.HistoryEntry
       for _, entry := range entries {
           if entry.VisitTime.After(lastExecutionTime) { // Strict greater-than
               filtered = append(filtered, entry)
           }
       }

       return filtered
   }
   ```

**Testing** (`internal/diff_table/table_test.go`):
- Mock TimestampStore interface for unit tests
- Test filterEntries() with various timestamp scenarios
- Test generateDiffHistory() with mock store and browser history

---

### Step 3: Update Extension Main (30 minutes)

**File**: `cmd/browser_extend_extension/main.go`

1. **Initialize timestamp store** (add before server creation):
   ```go
   // Initialize timestamp store
   timestampDBPath := "/tmp/osquery_browser_diff_timestamps.db"
   timestampStore, err := timestamp_store.NewSQLiteStore(timestampDBPath)
   if err != nil {
       log.Fatalf("Failed to initialize timestamp store: %v", err)
   }
   defer timestampStore.Close()
   ```

2. **Register diff table plugin** (add after existing browser_history table):
   ```go
   // Register browser_history_diff table
   debugLog("Registering browser_history_diff table plugin...")
   browserHistoryDiffTable := diff_table.NewPlugin(timestampStore)
   server.RegisterPlugin(browserHistoryDiffTable)
   debugLog("✓ browser_history_diff plugin registered successfully")
   ```

3. **Optional: Add configuration flags**:
   ```go
   timestampDBPath := flag.String("timestamp-db-path", "/tmp/osquery_browser_diff_timestamps.db", "Path to timestamp storage database")
   timestampReset := flag.Bool("timestamp-reset", false, "Clear all timestamps on startup")
   ```

---

### Step 4: Build and Test (1 hour)

1. **Build the extension**:
   ```bash
   make build-current
   # Or: go build -o build/browser_extend_extension ./cmd/browser_extend_extension
   ```

2. **Start osqueryd** (if not running):
   ```bash
   sudo osqueryd --ephemeral --disable_database \
       --extensions_autoload=/path/to/extensions.load \
       --extensions_socket=/tmp/osquery.sock
   ```

3. **Load the extension**:
   ```bash
   # Create extensions.load file
   echo "/path/to/browser_extend_extension" > /tmp/extensions.load
   ```

4. **Query the diff table** (first query):
   ```bash
   osqueryi --socket /tmp/osquery.sock
   ```
   ```sql
   SELECT COUNT(*) FROM browser_history_diff;
   -- Should return all browser history entries (first query)
   ```

5. **Query again** (subsequent query):
   ```sql
   SELECT * FROM browser_history_diff;
   -- Should return empty (no new history since last query)
   ```

6. **Add new browser history** (open a browser, visit a site)

7. **Query again**:
   ```sql
   SELECT * FROM browser_history_diff WHERE browser_type='chrome';
   -- Should return only the newly visited site
   ```

8. **Verify timestamps** (debugging):
   ```bash
   sqlite3 /tmp/osquery_browser_diff_timestamps.db \
       "SELECT table_name, browser_type, profile_id, datetime(last_execution_time, 'unixepoch') FROM execution_timestamps;"
   ```

---

## Testing Strategy

### Unit Tests

**timestamp_store package**:
```bash
go test ./internal/timestamp_store -v
```

Test cases:
- Get() with no record
- Update() + Get() roundtrip
- BatchUpdate() atomicity
- Concurrent reads (spawn 10 goroutines)
- Invalid input (empty profile_id, negative timestamp)

**diff_table package**:
```bash
go test ./internal/diff_table -v
```

Test cases:
- filterEntries() with zero time (first query)
- filterEntries() with valid timestamp (subsequent query)
- filterEntries() with timestamp equal to entry time (should exclude)
- generateDiffHistory() with mock store (table-driven tests)

### Integration Tests

**File**: `tests/integration/diff_table_test.go`

```go
func TestDiffTableIntegration(t *testing.T) {
    // 1. Setup: Create temp timestamp DB
    tmpDB := filepath.Join(t.TempDir(), "timestamps.db")
    store, _ := timestamp_store.NewSQLiteStore(tmpDB)
    defer store.Close()

    // 2. First query: Should return all entries
    results1, _ := generateDiffHistory(context.Background(), nil, store)
    assert.Greater(t, len(results1), 0, "First query should return entries")

    // 3. Immediate second query: Should return empty (no new history)
    results2, _ := generateDiffHistory(context.Background(), nil, store)
    assert.Equal(t, 0, len(results2), "Second query should be empty")

    // 4. Verify timestamps were updated
    lastExec, _ := store.Get("browser_history_diff", "chrome", "Default")
    assert.False(t, lastExec.IsZero(), "Timestamp should be set after query")
}
```

### Manual Testing Checklist

- [ ] Extension starts without errors
- [ ] First query returns all browser history entries
- [ ] Second immediate query returns empty results
- [ ] Timestamps persist across extension restarts
- [ ] Multiple browsers (Chrome, Firefox) work independently
- [ ] Standard `browser_history` table unaffected by diff table queries
- [ ] Logs show warnings if timestamp storage fails (simulate by deleting DB during query)

---

## Configuration Options

### Extension Flags

Add these optional flags to `main.go` for customization:

| Flag | Default | Description |
|------|---------|-------------|
| `--timestamp-db-path` | `/tmp/osquery_browser_diff_timestamps.db` | Path to timestamp storage database |
| `--timestamp-reset` | `false` | Clear all timestamps on extension startup (testing mode) |
| `--diff-table-enabled` | `true` | Enable/disable diff table registration |

Example:
```bash
./browser_extend_extension --socket=/tmp/osquery.sock \
    --timestamp-db-path=/var/lib/osquery/timestamps.db \
    --timestamp-reset=false
```

---

## Debugging Tips

### View Timestamp Database

```bash
sqlite3 /tmp/osquery_browser_diff_timestamps.db
```

```sql
-- View all timestamps
SELECT
    table_name,
    browser_type,
    profile_id,
    datetime(last_execution_time, 'unixepoch') as last_executed_at,
    (strftime('%s', 'now') - last_execution_time) as seconds_ago
FROM execution_timestamps
ORDER BY last_execution_time DESC;

-- Reset timestamps for testing
DELETE FROM execution_timestamps WHERE table_name = 'browser_history_diff';

-- Check schema version
PRAGMA user_version;
```

### Enable Debug Logging

```bash
./browser_extend_extension --socket=/tmp/osquery.sock --debug
```

Logs will show:
- Timestamp retrieval attempts
- Browser history query results
- Filtering operations
- Timestamp update successes/failures

### Common Issues

**Issue**: "Failed to open timestamp database: unable to open database file"
- **Cause**: `/tmp` directory doesn't exist or is not writable
- **Solution**: Create directory or use `--timestamp-db-path` with writable location

**Issue**: Query returns all entries every time (no filtering)
- **Cause**: Timestamps not persisting
- **Solution**: Check timestamp DB file exists, verify BatchUpdate() isn't failing silently

**Issue**: "No such table: browser_history_diff" in osqueryi
- **Cause**: Extension not loaded or table registration failed
- **Solution**: Check extension is in `extensions.load`, verify logs show "plugin registered successfully"

---

## Performance Tuning

### Expected Performance

| Operation | Target | Notes |
|-----------|--------|-------|
| Timestamp lookup | < 1ms | Indexed query on small table |
| Browser history query | 50-200ms | Depends on history size (1k-100k entries) |
| Filtering | 1-5ms | In-memory operation |
| Timestamp batch update | < 5ms | 4 profiles × 1 UPDATE each |
| **Total query time** | **100-500ms** | Dominated by browser history read |

### Optimization Opportunities (Future)

1. **Parallelize browser queries**: Query Chrome/Firefox concurrently using goroutines
2. **Index browser history**: Add index on `last_visit_time` in browser DBs (READ ONLY!)
3. **Streaming results**: Return results as they're filtered (avoid accumulating in memory)
4. **Connection pooling**: Reuse browser history DB connections across queries

---

## Deployment

### Production Considerations

1. **Timestamp DB location**: Use persistent path (not `/tmp`):
   ```bash
   --timestamp-db-path=/var/lib/osquery/browser_diff_timestamps.db
   ```

2. **Backup strategy**: Timestamps are re-creatable (first query returns all history)
   - Backup not critical, but useful for avoiding large result sets after outage

3. **Monitoring**:
   - Log aggregation: Monitor warnings about timestamp storage failures
   - osquery logging: Configure `logger_plugin` to capture diff table queries

4. **Graceful shutdown**: Extension calls `timestampStore.Close()` via defer
   - Ensures pending transactions commit before exit

### Security Considerations

1. **Read-only browser history**: Extension never modifies browser DBs (immutable=1 flag)
2. **Timestamp DB permissions**: Default 0644 (user read/write, others read)
   - Consider 0600 if multi-user system
3. **SQL injection**: Prepared statements used throughout (safe from injection)

---

## Next Steps

After implementing the feature:

1. **Run `/speckit.tasks`** to generate implementation task breakdown
2. **Review `tasks.md`** for prioritized implementation order
3. **Create tests first** (TDD approach if constitution requires)
4. **Implement in order**: timestamp_store → diff_table → main.go integration
5. **Iterate**: Test after each component, don't wait until end

---

## References

- **Feature Spec**: [spec.md](./spec.md)
- **Research Decisions**: [research.md](./research.md)
- **Data Model**: [data-model.md](./data-model.md)
- **Interface Contracts**: [contracts/timestamp_store.go](./contracts/timestamp_store.go), [contracts/diff_table.go](./contracts/diff_table.go)
- **osquery-go docs**: https://pkg.go.dev/github.com/osquery/osquery-go
- **SQLite WAL mode**: https://www.sqlite.org/wal.html

---

## Support

If you encounter issues during implementation:

1. Review error messages in extension logs (`/tmp/browser_extend_extension.log`)
2. Check timestamp database with `sqlite3` tool
3. Verify browser history databases are accessible (permissions, corruption)
4. Test components in isolation (unit tests for timestamp_store, diff_table)
5. Compare against existing `browser_history` table implementation in `main.go`

**Happy coding!** 🚀
