# Data Model: Periodic Diff Table

**Feature**: Periodic Diff Table for Browser History
**Date**: 2025-10-22
**Status**: Design

## Overview

This document defines the data entities, their attributes, relationships, and validation rules for the periodic diff table feature. The model consists of two primary components:

1. **Execution Timestamp Records** - Persistent state tracking last query execution times
2. **Filtered Browser History Entries** - Time-bounded browser history data returned by the diff table

## Entity Definitions

### 1. Execution Timestamp Record

**Purpose**: Track the last execution time for each combination of table, browser type, and profile to enable time-based filtering.

#### Attributes

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `table_name` | TEXT | NOT NULL, part of PRIMARY KEY | Identifier for the osquery table (e.g., 'browser_history_diff') |
| `browser_type` | TEXT | NOT NULL, part of PRIMARY KEY | Browser identifier: 'chrome', 'firefox', 'safari', 'edge' |
| `profile_id` | TEXT | NOT NULL, part of PRIMARY KEY | Browser profile identifier (e.g., 'Default', 'Profile 1', 'dev-edition-default') |
| `last_execution_time` | INTEGER | NOT NULL, >= 0 | Unix timestamp in seconds representing last successful query execution |

#### Primary Key
`(table_name, browser_type, profile_id)`

#### Indexes
- `idx_table_browser ON (table_name, browser_type)` - Optimize lookups for all profiles of a browser type

#### Storage
SQLite database at `/tmp/osquery_browser_diff_timestamps.db` (configurable via `--timestamp-db-path` flag)

#### Validation Rules

| Rule ID | Rule | Violation Behavior |
|---------|------|-------------------|
| VR-TS-001 | `table_name` must match pattern `^[a-z_]+$` (lowercase alphanumeric + underscore) | Reject write, log error |
| VR-TS-002 | `browser_type` must be one of: 'chrome', 'firefox', 'safari', 'edge' | Reject write, log error |
| VR-TS-003 | `profile_id` must not be empty string | Reject write, log error |
| VR-TS-004 | `last_execution_time` must be >= 0 (no negative timestamps) | Reject write, log error |
| VR-TS-005 | `last_execution_time` must not be more than 1 hour in the future (clock skew tolerance) | Log warning, allow write |

#### Lifecycle

**Creation**: First query execution for a (table_name, browser_type, profile_id) combination creates a new record with current timestamp.

**Update**: Each subsequent query execution updates the `last_execution_time` to the current time via `REPLACE INTO` (upsert operation).

**Deletion**: Not implemented in MVP. Future enhancement: purge timestamps for browsers/profiles no longer detected.

**State Transitions**:
```
[No Record] --first query--> [Record: timestamp=T0]
[Record: timestamp=T0] --query at T1--> [Record: timestamp=T1]
[Record: timestamp=Tn] --extension restart--> [Record: timestamp=Tn] (persisted)
```

#### Example Records

```sql
-- Chrome Default profile last queried at 2025-10-22 10:00:00 UTC (1729638000)
table_name='browser_history_diff', browser_type='chrome', profile_id='Default', last_execution_time=1729638000

-- Firefox dev profile last queried at 2025-10-22 10:05:00 UTC (1729638300)
table_name='browser_history_diff', browser_type='firefox', profile_id='dev-edition-default', last_execution_time=1729638300

-- Chrome Profile 1 never queried via diff table (standard browser_history table has separate timestamp)
table_name='browser_history', browser_type='chrome', profile_id='Profile 1', last_execution_time=1729638600
```

---

### 2. Browser History Entry (Diff Table View)

**Purpose**: Represents a single browser history visit that is included in the diff table query results (i.e., timestamp > last execution time).

This entity is **derived** (not persisted) - it's the filtered view of browser history entries from the underlying browser databases.

#### Attributes

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `time` | TEXT | NOT NULL, ISO 8601 format | Visit timestamp: 'YYYY-MM-DD HH:MM:SS' |
| `url` | TEXT | NOT NULL | Full URL visited (e.g., 'https://example.com/page') |
| `title` | TEXT | May be empty | Page title from browser history |
| `visit_count` | TEXT | NOT NULL, numeric string | Number of times URL visited (from browser DB) |
| `profile` | TEXT | NOT NULL | Profile identifier matching Execution Timestamp Record |
| `browser_type` | TEXT | NOT NULL | Browser identifier: 'chrome', 'firefox', 'safari', 'edge' |
| `browser_variant` | TEXT | May be empty | Specific browser variant (e.g., 'Chrome', 'Brave', 'Firefox Developer Edition') |

#### Validation Rules

| Rule ID | Rule | Violation Behavior |
|---------|------|-------------------|
| VR-HE-001 | `time` must parse as valid ISO 8601 datetime | Skip entry, log warning |
| VR-HE-002 | `url` must not be empty | Skip entry, log warning |
| VR-HE-003 | `url` must be valid URL format (basic check: starts with protocol) | Skip entry if malformed, log warning |
| VR-HE-004 | `visit_count` must be parseable as non-negative integer | Default to '1', log warning |

#### Filtering Logic

**Inclusion criteria for diff table**:
```
entry.visit_time > last_execution_time_for(table_name, entry.browser_type, entry.profile_id)
```

**Special cases**:
- **First query (no timestamp record)**: Include all entries (FR-004)
- **Timestamp storage failure**: Include all entries (fail-safe strategy from research.md)
- **Timestamp equal to last execution**: Exclude (strict greater-than comparison per FR-003)

#### Relationships

**Execution Timestamp Record → Browser History Entry**:
- Relationship: One-to-many (one timestamp record filters many history entries)
- Cardinality: 1:N
- Constraint: Timestamp record (browser_type, profile_id) must match history entry (browser_type, profile)
- Navigation: Look up timestamp record by (table_name='browser_history_diff', browser_type, profile_id), then filter history entries where visit_time > timestamp

---

## Data Flow

### Query Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ osqueryd: SELECT * FROM browser_history_diff;                   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│ diff_table.generateDiffHistory()                                │
│   1. Discover browser profiles (Chrome, Firefox, Safari, Edge)  │
│   2. For each (browser_type, profile_id):                       │
│      a. Retrieve timestamp from TimestampStore                  │
│      b. Query browser history DB (chromium.FindHistory, etc.)   │
│      c. Filter entries where visit_time > timestamp             │
│      d. Accumulate filtered entries                             │
│   3. Update all timestamps to current time (atomic batch)       │
│   4. Return filtered entries to osqueryd                        │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│ osqueryd logs results or sends to configured output             │
└─────────────────────────────────────────────────────────────────┘
```

### Timestamp Update Strategy

**Timing**: Update timestamps **after** successfully generating filtered results (FR-002: "after each successful query execution").

**Atomicity**: Batch update all timestamps for the current query in a single transaction to ensure consistency.

**Failure handling**: If timestamp update fails, log error but still return query results (fail-safe approach).

---

## Schema Definition (SQLite)

### execution_timestamps table

```sql
-- Schema version 1
PRAGMA user_version = 1;

CREATE TABLE IF NOT EXISTS execution_timestamps (
    table_name TEXT NOT NULL,
    browser_type TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    last_execution_time INTEGER NOT NULL,
    PRIMARY KEY (table_name, browser_type, profile_id)
) WITHOUT ROWID;

CREATE INDEX idx_table_browser
ON execution_timestamps(table_name, browser_type);

-- Validation trigger: Ensure valid browser_type
CREATE TRIGGER IF NOT EXISTS validate_browser_type
BEFORE INSERT ON execution_timestamps
FOR EACH ROW
WHEN NEW.browser_type NOT IN ('chrome', 'firefox', 'safari', 'edge')
BEGIN
    SELECT RAISE(ABORT, 'Invalid browser_type: must be chrome, firefox, safari, or edge');
END;

-- Validation trigger: Ensure non-negative timestamp
CREATE TRIGGER IF NOT EXISTS validate_timestamp
BEFORE INSERT ON execution_timestamps
FOR EACH ROW
WHEN NEW.last_execution_time < 0
BEGIN
    SELECT RAISE(ABORT, 'Invalid timestamp: must be non-negative');
END;

-- Validation trigger: Ensure non-empty profile_id
CREATE TRIGGER IF NOT EXISTS validate_profile_id
BEFORE INSERT ON execution_timestamps
FOR EACH ROW
WHEN NEW.profile_id = ''
BEGIN
    SELECT RAISE(ABORT, 'Invalid profile_id: must not be empty');
END;
```

### Migration Path

**Version 1 → Version 2** (future):
- Add `created_at` timestamp for audit trail
- Add `query_count` counter for analytics

```sql
-- Migration example (not implemented in MVP)
PRAGMA user_version = 2;

ALTER TABLE execution_timestamps
ADD COLUMN created_at INTEGER DEFAULT (strftime('%s', 'now'));

ALTER TABLE execution_timestamps
ADD COLUMN query_count INTEGER DEFAULT 1;
```

---

## Go Data Structures

### TimestampStore Interface

```go
package timestamp_store

import "time"

// TimestampStore manages execution timestamp persistence
type TimestampStore interface {
    // Get retrieves the last execution time for a table/browser/profile combination
    // Returns zero time if no record exists (first query case)
    Get(tableName, browserType, profileID string) (time.Time, error)

    // Update sets the last execution time for a table/browser/profile combination
    // Uses REPLACE semantics (insert or update)
    Update(tableName, browserType, profileID string, timestamp time.Time) error

    // BatchUpdate atomically updates multiple timestamps in a single transaction
    BatchUpdate(updates []TimestampUpdate) error

    // Close closes the underlying database connection
    Close() error
}

type TimestampUpdate struct {
    TableName   string
    BrowserType string
    ProfileID   string
    Timestamp   time.Time
}
```

### Implementation Example

```go
type sqliteStore struct {
    db *sql.DB
}

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

func (s *sqliteStore) BatchUpdate(updates []TimestampUpdate) error {
    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        REPLACE INTO execution_timestamps
        (table_name, browser_type, profile_id, last_execution_time)
        VALUES (?, ?, ?, ?)
    `)
    if err != nil {
        return fmt.Errorf("prepare statement: %w", err)
    }
    defer stmt.Close()

    for _, u := range updates {
        _, err := stmt.Exec(u.TableName, u.BrowserType, u.ProfileID, u.Timestamp.Unix())
        if err != nil {
            return fmt.Errorf("update timestamp for %s/%s/%s: %w",
                u.TableName, u.BrowserType, u.ProfileID, err)
        }
    }

    return tx.Commit()
}
```

---

## Data Integrity Constraints

### Cross-Entity Consistency

| Constraint ID | Description | Enforcement |
|---------------|-------------|-------------|
| CI-001 | browser_type values must be consistent between timestamp records and history entries | Application validates on write |
| CI-002 | profile_id values must match between timestamp records and browser profile discovery | Application validates on read |
| CI-003 | Timestamp updates must be atomic per query execution (all browsers or none) | Database transaction (BatchUpdate) |
| CI-004 | last_execution_time must always increase or stay same (no backward time travel) | Application logic (use time.Now() for updates) |

### Orphaned Data Handling

**Scenario**: Timestamp record exists but browser profile no longer detected (user deleted profile).
**Behavior**: Timestamp record remains (no automatic cleanup in MVP). No impact on functionality (profile won't be queried).

**Scenario**: Browser profile detected but no timestamp record exists.
**Behavior**: Treated as first query (return all history entries for that profile).

---

## Example Queries

### Retrieve All Timestamps for Debugging

```sql
SELECT
    table_name,
    browser_type,
    profile_id,
    datetime(last_execution_time, 'unixepoch') as last_executed_at
FROM execution_timestamps
ORDER BY last_execution_time DESC;
```

### Reset Timestamps for Testing

```sql
DELETE FROM execution_timestamps WHERE table_name = 'browser_history_diff';
```

### Find Profiles Not Queried Recently

```sql
SELECT
    browser_type,
    profile_id,
    datetime(last_execution_time, 'unixepoch') as last_executed_at,
    (strftime('%s', 'now') - last_execution_time) as seconds_since_query
FROM execution_timestamps
WHERE table_name = 'browser_history_diff'
  AND last_execution_time < (strftime('%s', 'now') - 3600)  -- More than 1 hour ago
ORDER BY last_execution_time ASC;
```

---

## Testing Considerations

### Unit Test Data

**Timestamp records for testing**:
```go
testRecords := []TimestampUpdate{
    {"browser_history_diff", "chrome", "Default", time.Unix(1729600000, 0)},
    {"browser_history_diff", "firefox", "default-release", time.Unix(1729600060, 0)},
    {"browser_history", "chrome", "Default", time.Unix(1729600120, 0)}, // Different table
}
```

### Integration Test Scenarios

1. **First query**: No timestamp records exist → return all history entries
2. **Subsequent query**: Timestamp exists → return only new entries since last query
3. **Mixed browsers**: Chrome has timestamp, Firefox does not → Chrome filtered, Firefox returns all
4. **Timestamp storage failure**: DB file deleted → return all entries (fail-safe)
5. **Concurrent queries**: Two queries start simultaneously → both should succeed with consistent timestamps

---

## Performance Characteristics

### Timestamp Lookup Performance
- **Expected**: < 1ms per lookup (indexed query on small table)
- **Scale**: O(1) with primary key or index lookup
- **Worst case**: O(N) full table scan if indexes missing (N = number of browser profiles, typically < 10)

### Timestamp Update Performance
- **Expected**: < 5ms for batch update of 4 browser profiles
- **Scale**: O(N) where N = number of profiles updated
- **Bottleneck**: SQLite write lock (serialized, but fast)

### Memory Usage
- **Timestamp storage**: ~100 bytes per record × number of profiles (~400 bytes for 4 profiles)
- **In-memory overhead**: Minimal (records loaded on-demand, not cached)

---

## Future Enhancements

1. **Timestamp expiration**: Automatically purge timestamps for profiles not seen in 30 days
2. **Query analytics**: Track query count, average time between queries per profile
3. **Audit trail**: Record timestamp change history for debugging
4. **Compression**: Store only date (not time) for older timestamps to save space
5. **Replication**: Sync timestamps across multiple osqueryd instances (multi-host coordination)
