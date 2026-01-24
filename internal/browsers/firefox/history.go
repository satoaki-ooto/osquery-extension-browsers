package firefox

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"osquery-extension-browsers/internal/browsers/common"

	_ "github.com/mattn/go-sqlite3"
)

// FindHistory discovers history entries for a specific Firefox profile.
//
// The function automatically handles missing places.sqlite databases by returning
// an empty slice with no error (silent skip). This ensures that profiles without
// history databases don't cause the entire browser history collection to fail.
//
// Behavior:
//   - If places.sqlite doesn't exist: returns empty []common.HistoryEntry with nil error
//   - If places.sqlite exists but is invalid: returns error from database operations
//   - If places.sqlite exists and is valid: returns history entries or database errors
//
// Optional parameters can be provided to filter results (e.g., by time or limit).
//
// This graceful handling aligns with the robust error handling pattern used throughout
// the extension, where individual profile failures don't stop overall processing.
func FindHistory(profile common.Profile, opts ...common.HistoryOption) ([]common.HistoryEntry, error) {
	historyDBPath := getHistoryDBPath(profile.Path)

	// Check if places.sqlite exists before attempting to open it
	if _, err := os.Stat(historyDBPath); os.IsNotExist(err) {
		// Return empty slice with no error (silent skip)
		return []common.HistoryEntry{}, nil
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro&immutable=1", historyDBPath))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Apply options
	options := common.ApplyHistoryOptions(opts)

	// Build query with optional filters
	query := `
		SELECT p.id, p.url, p.title, h.visit_date, p.visit_count
		FROM moz_places p
		JOIN moz_historyvisits h ON p.id = h.place_id
	`

	args := []interface{}{}
	if options.Since != nil {
		unixTime := toUnixTime(*options.Since)
		query += ` WHERE h.visit_date > ?`
		args = append(args, unixTime)
	}

	query += ` ORDER BY h.visit_date DESC`

	if options.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, options.Limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var historyEntries []common.HistoryEntry

	for rows.Next() {
		var id int64
		var url string
		var title sql.NullString
		var visitDate int64
		var visitCount int

		err := rows.Scan(&id, &url, &title, &visitDate, &visitCount)
		if err != nil {
			return nil, err
		}

		historyEntry := common.HistoryEntry{
			ID:             id,
			URL:            url,
			Title:          title.String,
			VisitTime:      parseUnixTime(visitDate),
			VisitCount:     visitCount,
			ProfileID:      profile.ID,
			BrowserType:    profile.BrowserType,
			BrowserVariant: profile.BrowserVariant,
		}

		historyEntries = append(historyEntries, historyEntry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return historyEntries, nil
}

// getHistoryDBPath returns path to history database for a given profile
func getHistoryDBPath(profilePath string) string {
	return filepath.Join(profilePath, "places.sqlite")
}

// parseUnixTime converts Unix timestamp to time.Time
// Firefox's timestamp is in microseconds since Unix epoch (1970-01-01 00:00:00 UTC)
func parseUnixTime(unixTime int64) time.Time {
	if unixTime == 0 {
		return time.Time{}
	}

	// Convert microseconds to nanoseconds for time.Unix
	return time.Unix(0, unixTime*1000)
}

// toUnixTime converts time.Time to Firefox's timestamp format
// Firefox's timestamp is in microseconds since Unix epoch (1970-01-01 00:00:00 UTC)
func toUnixTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	// Convert nanoseconds to microseconds
	return t.UnixNano() / 1000
}
