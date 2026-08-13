package chromium

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"osquery-extension-browsers/internal/browsers/common"
)

const webkitEpochOffsetMicroseconds int64 = 11644473600000000

// FindVisitObservations reads stable native Chromium visit rows.
func FindVisitObservations(
	profile common.Profile,
	nativeURLIDs []int64,
) (observations []common.VisitObservation, err error) {
	databasePath := filepath.Join(profile.Path, "History")
	exists, err := regularFileExists(databasePath)
	if err != nil {
		return nil, fmt.Errorf(
			"inspect %s history database for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}
	if !exists {
		return []common.VisitObservation{}, nil
	}

	database, err := common.OpenBrowserSQLite(databasePath)
	if err != nil {
		return nil, fmt.Errorf(
			"open %s history database for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}
	defer func() {
		if closeErr := database.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf(
				"close %s history database for profile %q: %w",
				profile.BrowserVariant,
				profile.Path,
				closeErr,
			)
		}
	}()

	query := "SELECT id, url, visit_time FROM visits"
	query, arguments := addInt64Filter(query, "url", nativeURLIDs)
	query += " ORDER BY id"

	rows, err := database.Query(query, arguments...)
	if err != nil {
		return nil, fmt.Errorf(
			"query %s visit observations for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}
	defer func() {
		if closeErr := rows.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf(
				"close %s visit rows for profile %q: %w",
				profile.BrowserVariant,
				profile.Path,
				closeErr,
			)
		}
	}()

	for rows.Next() {
		var nativeVisitID int64
		var nativeURLID int64
		var nativeVisitTime int64
		if err := rows.Scan(&nativeVisitID, &nativeURLID, &nativeVisitTime); err != nil {
			return nil, fmt.Errorf(
				"scan %s visit observation for profile %q: %w",
				profile.BrowserVariant,
				profile.Path,
				err,
			)
		}

		visitTimeUS := nativeVisitTime - webkitEpochOffsetMicroseconds
		observations = append(observations, common.VisitObservation{
			ObservationID: common.GenerateObservationID(
				profile.BrowserVariant,
				profile.Path,
				nativeVisitID,
			),
			ProfileID:      profile.ProfileID,
			BrowserFamily:  chromiumFamily,
			BrowserVariant: profile.BrowserVariant,
			NativeVisitID:  nativeVisitID,
			NativeURLID:    nativeURLID,
			VisitTime:      visitTimeUS / 1000000,
			VisitTimeUS:    visitTimeUS,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate %s visit observations for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}

	return observations, nil
}

// FindHistoryPages reads current Chromium URL/page state.
func FindHistoryPages(
	profile common.Profile,
	nativeURLIDs []int64,
) (pages []common.HistoryPage, err error) {
	databasePath := filepath.Join(profile.Path, "History")
	exists, err := regularFileExists(databasePath)
	if err != nil {
		return nil, fmt.Errorf(
			"inspect %s history database for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}
	if !exists {
		return []common.HistoryPage{}, nil
	}

	database, err := common.OpenBrowserSQLite(databasePath)
	if err != nil {
		return nil, fmt.Errorf(
			"open %s history database for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}
	defer func() {
		if closeErr := database.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf(
				"close %s history database for profile %q: %w",
				profile.BrowserVariant,
				profile.Path,
				closeErr,
			)
		}
	}()

	query := "SELECT id, url, title, visit_count, last_visit_time, hidden FROM urls"
	query, arguments := addInt64Filter(query, "id", nativeURLIDs)
	query += " ORDER BY id"

	rows, err := database.Query(query, arguments...)
	if err != nil {
		return nil, fmt.Errorf(
			"query %s history pages for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}
	defer func() {
		if closeErr := rows.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf(
				"close %s page rows for profile %q: %w",
				profile.BrowserVariant,
				profile.Path,
				closeErr,
			)
		}
	}()

	for rows.Next() {
		var nativeURLID int64
		var pageURL string
		var title sql.NullString
		var visitCount int64
		var nativeLastVisitTime sql.NullInt64
		var hidden sql.NullInt64
		if err := rows.Scan(
			&nativeURLID,
			&pageURL,
			&title,
			&visitCount,
			&nativeLastVisitTime,
			&hidden,
		); err != nil {
			return nil, fmt.Errorf(
				"scan %s history page for profile %q: %w",
				profile.BrowserVariant,
				profile.Path,
				err,
			)
		}

		page := common.HistoryPage{
			ProfileID:      profile.ProfileID,
			NativeURLID:    nativeURLID,
			BrowserFamily:  chromiumFamily,
			BrowserVariant: profile.BrowserVariant,
			URL:            pageURL,
			Title:          title.String,
			TitlePresent:   title.Valid,
			VisitCount:     visitCount,
			Hidden:         hidden.Int64,
			HiddenPresent:  hidden.Valid,
		}
		if nativeLastVisitTime.Valid && nativeLastVisitTime.Int64 != 0 {
			page.LastVisitTime = (nativeLastVisitTime.Int64 - webkitEpochOffsetMicroseconds) / 1000000
			page.LastVisitTimePresent = true
		}
		pages = append(pages, page)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate %s history pages for profile %q: %w",
			profile.BrowserVariant,
			profile.Path,
			err,
		)
	}

	return pages, nil
}

func regularFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("%q is a directory", path)
	}
	return true, nil
}

func addInt64Filter(query, column string, values []int64) (string, []any) {
	if len(values) == 0 {
		return query, nil
	}

	placeholders := make([]string, len(values))
	arguments := make([]any, len(values))
	for index, value := range values {
		placeholders[index] = "?"
		arguments[index] = value
	}
	return query + " WHERE " + column + " IN (" + strings.Join(placeholders, ",") + ")", arguments
}
