package chromium

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"osquery-extension-browsers/internal/browsers/common"

	_ "github.com/mattn/go-sqlite3"
)

// getHistoryDBPath returns the path to the history database for a given profile
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

// FindHistory discovers history entries for a specific profile
func FindHistory(profile common.Profile) ([]common.HistoryEntry, error) {
	historyDBPath := getHistoryDBPath(profile.Path)

	// Open the SQLite database
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro&immutable=1", historyDBPath))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query the history entries
	// We're using a simple query to get the most recent visits
	query := `
		SELECT id, url, title, last_visit_time, visit_count
		FROM urls
		ORDER BY last_visit_time DESC
	`

	rows, err := db.Query(query)
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
			BrowserType:    "Chromium",
			BrowserVariant: profile.BrowserVariant,
		}

		historyEntries = append(historyEntries, historyEntry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return historyEntries, nil
}
