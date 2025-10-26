// Package diff_table provides interface contracts for the periodic diff table implementation.
// This is a contract definition file (not executable code).
package diff_table

import (
	"context"

	"github.com/osquery/osquery-go/plugin/table"
)

// Plugin creates and returns an osquery table plugin for browser_history_diff.
//
// The returned plugin registers a table with osqueryd that provides time-based
// filtered browser history entries (only entries added since last query).
//
// Table schema:
//   - time (TEXT): Visit timestamp in 'YYYY-MM-DD HH:MM:SS' format
//   - url (TEXT): Full URL visited
//   - title (TEXT): Page title from browser history
//   - visit_count (TEXT): Number of times URL visited (numeric string)
//   - profile (TEXT): Browser profile identifier (e.g., "Default")
//   - browser_type (TEXT): Browser identifier ("chrome", "firefox", "safari", "edge")
//   - browser_variant (TEXT): Specific browser variant (e.g., "Chrome", "Brave")
//
// Parameters:
//   - store: TimestampStore implementation for tracking query execution times
//
// Returns:
//   - *table.Plugin: osquery table plugin ready for registration
//
// Example usage in main.go:
//   store, _ := timestamp_store.NewSQLiteStore("/tmp/osquery_browser_diff_timestamps.db")
//   diffTable := diff_table.NewPlugin(store)
//   server.RegisterPlugin(diffTable)
//
// Behavior:
//   - First query per browser/profile: Returns all available history entries
//   - Subsequent queries: Returns only entries with timestamp > last execution time
//   - Empty results: Valid response when no new history entries exist
//   - Storage failure: Logs warning, returns all entries (fail-safe)
//
// Thread-safety: Plugin callbacks are invoked serially by osqueryd (no concurrent access)
func NewPlugin(store TimestampStore) *table.Plugin {
	panic("Contract definition - implement in diff_table/table.go")
}

// TimestampStore defines the interface required by the diff table for timestamp persistence.
// This interface should be satisfied by timestamp_store.TimestampStore.
type TimestampStore interface {
	Get(tableName, browserType, profileID string) (time.Time, error)
	BatchUpdate(updates []TimestampUpdate) error
}

// TimestampUpdate represents a timestamp update for batch operations.
// Must match timestamp_store.TimestampUpdate structure.
type TimestampUpdate struct {
	TableName   string
	BrowserType string
	ProfileID   string
	Timestamp   time.Time
}

// generateDiffHistory is the callback function invoked by osqueryd when querying browser_history_diff.
//
// Implementation outline:
//  1. Discover browser profiles (chromium.FindProfiles, firefox.FindProfiles, etc.)
//  2. For each (browser_type, profile_id):
//     a. Retrieve last execution timestamp from store
//     b. Query browser history database (chromium.FindHistory, etc.)
//     c. Filter entries where entry.VisitTime > lastExecutionTime
//     d. Accumulate filtered entries in results slice
//     e. Record (browser_type, profile_id, time.Now()) for batch timestamp update
//  3. Batch update all timestamps atomically
//  4. Return filtered entries to osqueryd
//
// Parameters:
//   - ctx: Context for cancellation (osqueryd may cancel long-running queries)
//   - queryContext: Query constraints from osqueryd (unused in MVP - no WHERE clause support)
//
// Returns:
//   - []map[string]string: Rows matching table schema (keys: time, url, title, etc.)
//   - error: Non-nil if critical failure prevents returning results
//
// Error handling:
//   - Timestamp retrieval failure: Log warning, treat as first query (include all entries)
//   - Browser history read failure: Log error, skip that browser/profile (continue with others)
//   - Timestamp update failure: Log error, still return query results (fail-safe)
//   - Empty results: Return empty slice (not an error condition)
//
// Performance considerations:
//   - Expected query time: 100-500ms depending on browser history size
//   - Optimization: Could parallelize browser history queries (future enhancement)
//   - Memory usage: Accumulates all filtered entries in memory before returning
//
// Contract definition - actual implementation in diff_table/table.go
func generateDiffHistory(ctx context.Context, queryContext table.QueryContext,
	store TimestampStore) ([]map[string]string, error) {
	panic("Contract definition - implement in diff_table/table.go")
}

// filterEntries applies time-based filtering to browser history entries.
//
// Parameters:
//   - entries: Raw browser history entries from chromium.FindHistory() or firefox.FindHistory()
//   - lastExecutionTime: Timestamp from store.Get() (zero time if first query)
//
// Returns:
//   - []common.HistoryEntry: Entries where entry.VisitTime > lastExecutionTime
//
// Behavior:
//   - lastExecutionTime.IsZero(): Return all entries (first query case)
//   - Strict greater-than: Entries with VisitTime == lastExecutionTime are excluded
//   - Empty input: Return empty slice (not an error)
//
// Example:
//   lastExec := time.Unix(1729638000, 0)  // 2025-10-22 10:00:00 UTC
//   allEntries := chromium.FindHistory(profile)
//   // Returns only entries with VisitTime > 2025-10-22 10:00:00
//   filtered := filterEntries(allEntries, lastExec)
//
// Contract definition - actual implementation in diff_table/filter.go
func filterEntries(entries []HistoryEntry, lastExecutionTime time.Time) []HistoryEntry {
	panic("Contract definition - implement in diff_table/filter.go")
}

// HistoryEntry represents a browser history entry from internal/browsers/common.HistoryEntry.
// This is a placeholder for the actual type (imported from common package at runtime).
type HistoryEntry struct {
	ID             int64
	URL            string
	Title          string
	VisitTime      time.Time
	VisitCount     int
	ProfileID      string
	BrowserType    string
	BrowserVariant string
}

// formatHistoryEntry converts a common.HistoryEntry to the osquery table row format.
//
// Parameters:
//   - entry: Browser history entry from chromium.FindHistory() or firefox.FindHistory()
//
// Returns:
//   - map[string]string: Row with keys matching table schema (time, url, title, visit_count, profile, browser_type, browser_variant)
//
// Behavior:
//   - Time formatting: entry.VisitTime.Format("2006-01-02 15:04:05") for ISO 8601 compliance
//   - visit_count conversion: strconv.Itoa(entry.VisitCount) to convert int → string
//   - Empty title: Preserved as empty string (valid per schema)
//
// Example output:
//   {
//       "time": "2025-10-22 10:05:30",
//       "url": "https://example.com/page",
//       "title": "Example Page",
//       "visit_count": "5",
//       "profile": "Default",
//       "browser_type": "chrome",
//       "browser_variant": "Chrome"
//   }
//
// Contract definition - actual implementation in diff_table/table.go
func formatHistoryEntry(entry HistoryEntry) map[string]string {
	panic("Contract definition - implement in diff_table/table.go")
}

// Behavioral Contracts (documented for implementers)
//
// 1. Table registration:
//    - Table name must be "browser_history_diff" (distinct from "browser_history")
//    - Columns must match schema defined in NewPlugin() docstring
//    - Plugin must be registered after extension server creation, before server.Run()
//
// 2. Timestamp semantics:
//    - Timestamps updated AFTER successful query generation (not before)
//    - Batch update ensures atomic timestamp consistency across all browsers
//    - Failed timestamp updates don't prevent returning query results (fail-safe)
//
// 3. Query isolation:
//    - "browser_history" table and "browser_history_diff" table maintain independent timestamps
//    - Querying one table does not affect the other table's timestamps
//    - Enforced via table_name column in timestamp_store
//
// 4. Empty result handling:
//    - Empty slice (0 entries) is a valid response when no new history exists
//    - osqueryd logs empty results as normal (not an error condition)
//    - Timestamps still updated even for empty results (query executed successfully)
//
// 5. Error propagation:
//    - Storage errors: Log + continue (fail-safe approach from research.md)
//    - Browser detection errors: Log + skip that browser (continue with others)
//    - Critical errors (e.g., all browsers failed): Return error to osqueryd
//
// 6. Performance guarantees:
//    - Query timeout: Respect ctx.Done() for cancellation
//    - Memory bounds: Stream results if >10k entries (future enhancement)
//    - Concurrency: Single-threaded in MVP (osqueryd serializes queries)
//
// 7. Testing contracts:
//    - Mock TimestampStore for unit tests (no actual DB required)
//    - Table-driven tests for filterEntries() with various timestamp scenarios
//    - Integration tests with real browser history fixtures
