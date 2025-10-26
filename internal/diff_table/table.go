// Package diff_table implements an osquery table plugin for browser history with periodic diff functionality.
package diff_table

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/osquery/osquery-go/plugin/table"
	"osquery-extension-browsers/internal/browsers/chromium"
	"osquery-extension-browsers/internal/browsers/common"
	"osquery-extension-browsers/internal/browsers/firefox"
	"osquery-extension-browsers/internal/timestamp_store"
)

// Table implements the osquery table plugin for browser history with periodic diff
type Table struct {
	store *timestamp_store.Store
}

// New creates a new instance of the diff table
func New() (*Table, error) {
	// Store timestamps in the user's cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	dbPath := filepath.Join(cacheDir, "osquery-browsers", "timestamps.db")
	store, err := timestamp_store.NewStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create timestamp store: %w", err)
	}

	return &Table{
		store: store,
	}, nil
}

// Columns returns the columns for the browser history diff table
func (t *Table) Columns() []table.ColumnDefinition {
	return []table.ColumnDefinition{
		table.TextColumn("time"),
		table.TextColumn("title"),
		table.TextColumn("url"),
		table.TextColumn("profile"),
		table.TextColumn("browser_type"),
		table.TextColumn("browser_variant"),
		table.TextColumn("visit_count"),
	}
}

// Generate generates the table data for browser history with periodic diff
func (t *Table) Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	// Get current time at the start of processing
	currentTime := time.Now()

	// Use a wait group to process different browser types in parallel
	var wg sync.WaitGroup
	resultsChan := make(chan []map[string]string, 2) // Buffer for Chromium and Firefox
	errChan := make(chan error, 2)

	// Process Chromium history
	wrappedFunc := func() {
		defer wg.Done()
		entries, err := t.getBrowserHistory(ctx, "chromium", chromium.FindProfiles, chromium.FindHistory)
		if err != nil {
			errChan <- fmt.Errorf("error getting Chromium history: %w", err)
			return
		}
		resultsChan <- entries
	}

	wg.Add(1)
	go wrappedFunc()

	// Process Firefox history
	wrappedFunc = func() {
		defer wg.Done()
		entries, err := t.getBrowserHistory(ctx, "firefox", firefox.FindProfiles, firefox.FindHistory)
		if err != nil {
			errChan <- fmt.Errorf("error getting Firefox history: %w", err)
			return
		}
		resultsChan <- entries
	}

	wg.Add(1)
	go wrappedFunc()

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errChan)
	}()

	// Collect results
	var results []map[string]string
	for entryBatch := range resultsChan {
		results = append(results, entryBatch...)
	}

	// Check for errors
	if len(errChan) > 0 {
		for err := range errChan {
			return nil, err
		}
	}

	// Update timestamps after successful query
	t.store.UpdateExecutionTime(ctx, "chromium", "browser_history_diff", currentTime)
	t.store.UpdateExecutionTime(ctx, "firefox", "browser_history_diff", currentTime)

	return results, nil
}

// getBrowserHistory retrieves browser history entries since the last execution time
func (t *Table) getBrowserHistory(
	ctx context.Context,
	browserType string,
	findProfilesFunc func() ([]common.Profile, error),
	findHistoryFunc func(profile common.Profile) ([]common.HistoryEntry, error),
) ([]map[string]string, error) {
	// Get last execution time for this browser type
	lastTime, err := t.store.GetLastExecutionTime(ctx, browserType, "browser_history_diff")
	if err != nil {
		return nil, fmt.Errorf("failed to get last execution time: %w", err)
	}

	// Find all profiles for this browser type
	profiles, err := findProfilesFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to find %s profiles: %w", browserType, err)
	}

	var results []map[string]string

	// Process each profile
	for _, profile := range profiles {
		historyEntries, err := findHistoryFunc(profile)
		if err != nil {
			// Log error but continue with other profiles
			continue
		}

		// Convert to map and filter by timestamp
		for _, entry := range historyEntries {
			if entry.VisitTime.After(lastTime) {
				results = append(results, map[string]string{
					"time":            entry.VisitTime.Format("2006-01-02 15:04:05"),
					"url":             entry.URL,
					"title":           entry.Title,
					"visit_count":     fmt.Sprintf("%d", entry.VisitCount),
					"profile":         entry.ProfileID,
					"browser_type":    entry.BrowserType,
					"browser_variant": entry.BrowserVariant,
				})
			}
		}
	}

	return results, nil
}

// Close releases any resources used by the table
func (t *Table) Close() error {
	if t.store != nil {
		return t.store.Close()
	}
	return nil
}
