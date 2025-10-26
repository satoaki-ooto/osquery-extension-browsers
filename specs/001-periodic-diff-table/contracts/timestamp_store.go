// Package timestamp_store provides interface contracts for execution timestamp persistence.
// This is a contract definition file (not executable code).
package timestamp_store

import "time"

// TimestampStore manages execution timestamp persistence for osquery table queries.
// Implementations must be thread-safe and support concurrent reads with serialized writes.
type TimestampStore interface {
	// Get retrieves the last execution time for a specific table/browser/profile combination.
	//
	// Returns:
	//   - time.Time: Last execution timestamp (zero value if no record exists)
	//   - error: Non-nil if storage layer fails (e.g., DB unavailable, corruption)
	//
	// Behavior:
	//   - First query (no record): Returns time.Time{} (zero time), nil error
	//   - Subsequent queries: Returns stored timestamp, nil error
	//   - Storage failure: Returns time.Time{}, error describing failure
	//
	// Thread-safety: Safe for concurrent calls (read-only operation)
	Get(tableName, browserType, profileID string) (time.Time, error)

	// Update sets the last execution time for a specific table/browser/profile combination.
	// Uses REPLACE/UPSERT semantics: creates new record if none exists, updates if present.
	//
	// Parameters:
	//   - tableName: osquery table name (e.g., "browser_history_diff")
	//   - browserType: Browser identifier - must be one of: "chrome", "firefox", "safari", "edge"
	//   - profileID: Browser profile identifier (e.g., "Default", "Profile 1")
	//   - timestamp: Unix timestamp to store (should be time.Now() at query execution)
	//
	// Returns:
	//   - error: Non-nil if update fails (e.g., validation error, DB unavailable)
	//
	// Validation:
	//   - browserType must match allowed values (enforced by DB trigger or application logic)
	//   - profileID must not be empty
	//   - timestamp must be non-negative (>= Unix epoch)
	//
	// Thread-safety: Safe for concurrent calls (writes are serialized by SQLite)
	Update(tableName, browserType, profileID string, timestamp time.Time) error

	// BatchUpdate atomically updates multiple timestamps in a single database transaction.
	// Provides all-or-nothing semantics: either all updates succeed or none are applied.
	//
	// Parameters:
	//   - updates: Slice of TimestampUpdate structs containing timestamp data
	//
	// Returns:
	//   - error: Non-nil if any update fails or transaction cannot commit
	//
	// Behavior:
	//   - Empty slice: No-op, returns nil
	//   - Partial validation failure: Rolls back entire transaction, returns first error encountered
	//   - Commit failure: Rolls back, returns commit error
	//
	// Use case: Update all browser profile timestamps at end of diff table query
	//
	// Thread-safety: Safe for concurrent calls (transactions are serialized)
	BatchUpdate(updates []TimestampUpdate) error

	// Close gracefully shuts down the timestamp store and releases resources.
	// Must be called before application exit to ensure pending writes are flushed.
	//
	// Returns:
	//   - error: Non-nil if cleanup fails (e.g., DB connection close error)
	//
	// Behavior after Close():
	//   - All subsequent Get/Update/BatchUpdate calls should return error
	//   - Pending transactions should be committed or rolled back
	//   - Database connections should be closed
	//
	// Thread-safety: NOT safe for concurrent calls with Get/Update/BatchUpdate
	//                (caller must ensure no concurrent operations during shutdown)
	Close() error
}

// TimestampUpdate represents a single timestamp update operation for batch processing.
type TimestampUpdate struct {
	TableName   string    // osquery table name (e.g., "browser_history_diff")
	BrowserType string    // Browser identifier: "chrome", "firefox", "safari", "edge"
	ProfileID   string    // Browser profile identifier (e.g., "Default")
	Timestamp   time.Time // Unix timestamp to store
}

// NewSQLiteStore creates a TimestampStore implementation backed by SQLite.
//
// Parameters:
//   - dbPath: File path to SQLite database (created if not exists)
//
// Returns:
//   - TimestampStore: Initialized store ready for use
//   - error: Non-nil if database cannot be opened or schema initialization fails
//
// Implementation notes:
//   - Enables WAL mode for better concurrency
//   - Creates schema if database is new
//   - Configures connection pool (max open connections: 1 for write serialization)
//
// Example:
//   store, err := timestamp_store.NewSQLiteStore("/tmp/osquery_browser_diff_timestamps.db")
//   if err != nil {
//       log.Fatalf("Failed to initialize timestamp store: %v", err)
//   }
//   defer store.Close()
func NewSQLiteStore(dbPath string) (TimestampStore, error) {
	panic("Contract definition - implement in timestamp_store/store.go")
}

// Error types for timestamp store operations
var (
	// ErrInvalidBrowserType indicates an unsupported browser type was provided
	// Valid types: "chrome", "firefox", "safari", "edge"
	ErrInvalidBrowserType = newTimestampStoreError("invalid browser type")

	// ErrEmptyProfileID indicates an empty profile ID was provided
	ErrEmptyProfileID = newTimestampStoreError("profile ID cannot be empty")

	// ErrNegativeTimestamp indicates a negative timestamp was provided
	ErrNegativeTimestamp = newTimestampStoreError("timestamp cannot be negative")

	// ErrStoreClosed indicates an operation was attempted on a closed store
	ErrStoreClosed = newTimestampStoreError("timestamp store is closed")
)

// TimestampStoreError wraps timestamp store operation errors
type TimestampStoreError struct {
	Message string
	Cause   error
}

func (e *TimestampStoreError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *TimestampStoreError) Unwrap() error {
	return e.Cause
}

func newTimestampStoreError(message string) *TimestampStoreError {
	return &TimestampStoreError{Message: message}
}

// Behavioral Contracts (not enforced by Go compiler, documented for implementers)
//
// 1. Idempotency:
//    - Update(table, browser, profile, T1) followed by Update(same params, T1) is idempotent
//    - Get() is always idempotent
//
// 2. Ordering guarantees:
//    - Update(T1) → Update(T2): If T2 > T1, later update always wins
//    - Concurrent Update(T1) || Update(T2): Undefined which wins (last write wins, determined by SQLite)
//
// 3. Durability:
//    - After Update() returns nil error, timestamp is persisted (survives process crash)
//    - BatchUpdate() commit provides atomic durability guarantee
//
// 4. Isolation:
//    - Get() during concurrent Update(): Returns either old or new value (read committed isolation)
//    - BatchUpdate() is isolated: external reads see all updates or none
//
// 5. Consistency:
//    - Timestamps always increase or stay same (no backward time travel)
//    - Primary key constraint enforced: max 1 record per (table, browser, profile)
//
// 6. Error handling philosophy:
//    - Storage errors (DB unavailable): Return error, caller decides fail-safe vs fail-closed
//    - Validation errors (invalid input): Return error immediately, no side effects
//    - Close() errors: Best-effort cleanup, log but don't panic
