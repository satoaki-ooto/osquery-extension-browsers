// Package timestamp_store provides functionality for persisting and retrieving
// execution timestamps for browser history queries.
package timestamp_store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store manages persistence of execution timestamps for browser history queries
type Store struct {
	db   *sql.DB
	path string
	mu   sync.Mutex
}

// NewStore creates a new timestamp store at the specified path
func NewStore(path string) (*Store, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create timestamp store directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open timestamp store: %w", err)
	}

	s := &Store{
		db:   db,
		path: path,
	}

	// Initialize schema
	if err := s.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize timestamp store: %w", err)
	}

	return s, nil
}

// init initializes the database schema
func (s *Store) init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS execution_timestamps (
			browser_type TEXT NOT NULL,
			table_name TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			PRIMARY KEY (browser_type, table_name)
		);
	`)
	return err
}

// GetLastExecutionTime returns the last execution time for the given browser type and table
func (s *Store) GetLastExecutionTime(ctx context.Context, browserType, tableName string) (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var timestamp int64
	err := s.db.QueryRowContext(
		ctx,
		"SELECT timestamp FROM execution_timestamps WHERE browser_type = ? AND table_name = ?",
		browserType, tableName,
	).Scan(&timestamp)

	switch {
	case err == sql.ErrNoRows:
		// No previous timestamp, return zero time
		return time.Time{}, nil
	case err != nil:
		return time.Time{}, fmt.Errorf("failed to get last execution time: %w", err)
	default:
		return time.Unix(0, timestamp), nil
	}
}

// UpdateExecutionTime updates the last execution time for the given browser type and table
func (s *Store) UpdateExecutionTime(ctx context.Context, browserType, tableName string, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := t.UnixNano()
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO execution_timestamps (browser_type, table_name, timestamp)
		 VALUES (?, ?, ?)
		 ON CONFLICT(browser_type, table_name) 
		 DO UPDATE SET timestamp = excluded.timestamp`,
		browserType, tableName, timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to update execution time: %w", err)
	}

	return nil
}

// Close closes the store and releases any database resources
func (s *Store) Close() error {
	return s.db.Close()
}
