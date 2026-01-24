# Design: Incremental Browser History Collection

**Change ID:** `add-periodic-browser-history-collection`

## Architecture Overview

The incremental collection feature will be implemented as a simple extension to the existing query interface:

```
┌─────────────────────────────────────────────────────────────┐
│                    osquery Extension                        │
│                                                             │
│  ┌──────────────────┐      ┌──────────────────┐            │
│  │  Virtual Tables  │      │  State Manager   │            │
│  │  (with filters)  │      │  (JSON file)     │            │
│  └────────┬─────────┘      └────────┬─────────┘            │
│           │                         │                       │
│  ┌────────▼─────────────────────────▼─────────┐            │
│  │         Browser History Collectors         │            │
│  │  (Chromium & Firefox with since param)     │            │
│  └────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. Browser History Collectors

**Location**: `internal/browsers/chromium/history.go`, `internal/browsers/firefox/history.go`

**Current Interface**:
```go
func FindHistory(profile common.Profile) ([]common.HistoryEntry, error)
```

**New Interface**:
```go
type HistoryOptions struct {
    Since *time.Time
    Limit int
}

func FindHistory(profile common.Profile, opts ...HistoryOption) ([]common.HistoryEntry, error)
```

**Implementation**:
- Add optional `since` parameter to filter by visit time
- Use browser database's indexed timestamp columns for efficient filtering
- Maintain backward compatibility (nil `since` means get all history)

**Chromium Example**:
```go
func FindHistory(profile common.Profile, opts ...HistoryOption) ([]common.HistoryEntry, error) {
    options := &HistoryOptions{}
    for _, opt := range opts {
        opt(options)
    }
    
    query := `
        SELECT id, url, title, last_visit_time, visit_count
        FROM urls
    `
    
    args := []interface{}{}
    if options.Since != nil {
        chromeTime := toChromeTime(*options.Since)
        query += ` WHERE last_visit_time > ?`
        args = append(args, chromeTime)
    }
    
    query += ` ORDER BY last_visit_time DESC`
    
    if options.Limit > 0 {
        query += ` LIMIT ?`
        args = append(args, options.Limit)
    }
    
    rows, err := db.Query(query, args...)
    // ...
}
```

### 2. Virtual Table Schema

**Location**: `cmd/browser_extend_extension/main.go`

**Current Schema**:
```go
columns := []table.ColumnDefinition{
    table.TextColumn("time"),
    table.TextColumn("title"),
    table.TextColumn("url"),
    table.TextColumn("profile"),
    table.TextColumn("browser_type"),
}
```

**New Schema**:
```go
columns := []table.ColumnDefinition{
    table.TextColumn("time"),
    table.BigIntColumn("visit_time"),     // Unix timestamp for filtering
    table.TextColumn("title"),
    table.TextColumn("url"),
    table.TextColumn("profile"),
    table.TextColumn("browser_type"),
    table.TextColumn("browser_variant"),
    table.IntegerColumn("visit_count"),
}
```

**Query Support**:
```sql
-- Get all history
SELECT * FROM browser_history;

-- Get history since timestamp (for incremental collection)
SELECT * FROM browser_history WHERE visit_time > 1234567890;

-- Get recent history with limit
SELECT * FROM browser_history ORDER BY visit_time DESC LIMIT 100;
```

### 3. State Management

**Location**: `internal/state/manager.go`

**Purpose**: Track last collection timestamp for incremental queries

**Implementation**:
```go
type StateManager struct {
    stateFile string
    mu        sync.RWMutex
}

type CollectionState struct {
    LastCollectionTime int64 `json:"last_collection_time"`
    ProfileStates      map[string]int64 `json:"profile_states"`
}

func (sm *StateManager) GetLastCollectionTime() (time.Time, error)
func (sm *StateManager) UpdateLastCollectionTime(t time.Time) error
```

**State File Format** (JSON):
```json
{
  "last_collection_time": 1234567890,
  "profile_states": {
    "chrome_default": 1234567890,
    "firefox_default": 1234567880
  }
}
```

**Default Location**: `~/.osquery/browser_history_state.json`

### 4. Configuration

**Command-Line Flags**:
```go
--state-file string    Path to state file (default: "~/.osquery/browser_history_state.json")
```

**osquery Scheduled Query Configuration**:
```json
{
  "schedule": {
    "browser_history_collection": {
      "query": "SELECT * FROM browser_history WHERE visit_time > (SELECT MAX(visit_time) FROM browser_history)",
      "interval": 300
    }
  }
}
```

## Performance Considerations

### Query Performance
- **Indexed Columns**: Browser databases already index timestamp columns
- **Push-down Filters**: osquery pushes WHERE clauses to the extension
- **Efficient Pagination**: Use LIMIT and OFFSET for large result sets

### State Management
- **Minimal I/O**: State file only updated after successful collection
- **Concurrent Safety**: Use file locking or atomic writes
- **Error Recovery**: Graceful handling of corrupted state files

## Error Handling Strategy

### Collection Errors
- **Database Locks**: Retry with exponential backoff
- **Missing Databases**: Log and continue
- **Parse Errors**: Skip invalid entries, log warnings

### State Management Errors
- **Corrupted State**: Reset to zero timestamp (full collection)
- **File Permission Errors**: Log error, disable incremental collection
- **Concurrent Access**: Use file locking to prevent corruption

## Testing Strategy

### Unit Tests
- History collection with `since` parameter
- State file read/write operations
- Error handling for corrupted state
- Backward compatibility (nil `since`)

### Integration Tests
- Full incremental collection workflow
- State persistence across restarts
- Multiple profile handling
- Concurrent query scenarios

### Performance Tests
- Large history database filtering
- State file I/O performance
- Memory usage with incremental queries

## Deployment Considerations

### Migration Path
- **New Installation**: Clean state file, full initial collection
- **Existing Installation**: Backward compatible, no migration needed
- **Rollback**: Remove `since` parameter usage

### Monitoring
- **Query Performance**: Log slow queries
- **Collection Metrics**: Track entries collected per query
- **Error Rates**: Monitor state file errors

### Operational Procedures
- **State File Backup**: Include in regular backup strategy
- **Cleanup**: Manual deletion of state file triggers full collection
- **Troubleshooting**: Check logs for state file errors