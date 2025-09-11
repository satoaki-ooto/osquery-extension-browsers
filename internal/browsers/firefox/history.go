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
// This graceful handling aligns with the robust error handling pattern used throughout
// the extension, where individual profile failures don't stop overall processing.
func FindHistory(profile common.Profile) ([]common.HistoryEntry, error) {
	historyDBPath := getHistoryDBPath(profile.Path)

	// Check if places.sqlite exists before attempting to open it
	if _, err := os.Stat(historyDBPath); os.IsNotExist(err) {
		// Return empty slice with no error (silent skip)
		return []common.HistoryEntry{}, nil
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro&immutable=1", historyDBPath))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query the history entries
	// We're using a simple query to get the most recent visits
	query := `
		SELECT p.id, p.url, p.title, h.visit_date, p.visit_count
		FROM moz_places p
		JOIN moz_historyvisits h ON p.id = h.place_id
		ORDER BY h.visit_date DESC
	`

	rows, err := db.Query(query)
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

// getHistoryDBPath returns the path to the history database for a given profile
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
