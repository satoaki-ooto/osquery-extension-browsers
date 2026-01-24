package chromium

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"osquery-extension-browsers/internal/browsers/common"

	_ "github.com/mattn/go-sqlite3"
)

// getHistoryDBPath returns path to the history database for a given profile
func getHistoryDBPath(profilePath string) string {
	return filepath.Join(profilePath, "History")
}

// parseChromeTime converts Chrome's timestamp format to time.Time
// Chrome's timestamp is in microseconds since Windows epoch (1601-01-01 00:00:00 UTC)
func parseChromeTime(chromeTime int64) time.Time {
	// Windows epoch starts at 1601-01-01 00:00:00 UTC
	// Unix epoch starts at 1970-01-01 00:00:00 UTC
	// Difference is 11644473600 seconds
	const windowsEpochOffset = 11644473600 * 1000000 // in microseconds

	if chromeTime == 0 {
		return time.Time{}
	}

	// Convert microseconds to nanoseconds for time.Unix
	unixMicroseconds := chromeTime - windowsEpochOffset
	return time.Unix(0, unixMicroseconds*1000)
}

// toChromeTime converts time.Time to Chrome's timestamp format
// Chrome's timestamp is in microseconds since Windows epoch (1601-01-01 00:00:00 UTC)
func toChromeTime(t time.Time) int64 {
	// Windows epoch starts at 1601-01-01 00:00:00 UTC
	// Unix epoch starts at 1970-01-01 00:00:00 UTC
	// Difference is 11644473600 seconds
	const windowsEpochOffset = 11644473600 * 1000000 // in microseconds

	if t.IsZero() {
		return 0
	}

	// Convert nanoseconds to microseconds and add Windows epoch offset
	return t.UnixNano()/1000 + windowsEpochOffset
}

// FindHistory discovers history entries for a specific profile.
// Optional parameters can be provided to filter results (e.g., by time or limit).
func FindHistory(profile common.Profile, opts ...common.HistoryOption) ([]common.HistoryEntry, error) {
	historyDBPath := getHistoryDBPath(profile.Path)

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
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var historyEntries []common.HistoryEntry

	for rows.Next() {
		var id int64
		var url, title string
		var lastVisitTime int64
		var visitCount int

		err := rows.Scan(&id, &url, &title, &lastVisitTime, &visitCount)
		if err != nil {
			return nil, err
		}

		historyEntry := common.HistoryEntry{
			ID:             id,
			URL:            url,
			Title:          title,
			VisitTime:      parseChromeTime(lastVisitTime),
			VisitCount:     visitCount,
			ProfileID:      profile.ID,
			BrowserType:    strings.ToLower(profile.BrowserVariant),
			BrowserVariant: profile.BrowserVariant,
		}

		historyEntries = append(historyEntries, historyEntry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return historyEntries, nil
}
