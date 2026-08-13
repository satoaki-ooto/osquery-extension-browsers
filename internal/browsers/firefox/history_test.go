package firefox

import (
	"database/sql"
	"path/filepath"
	"testing"

	"osquery-extension-browsers/internal/browsers/common"
)

func TestVisitObservationsUseFirefoxVisitID(t *testing.T) {
	profile := firefoxFixtureProfile(t)
	database, err := sql.Open("sqlite3", filepath.Join(profile.Path, "places.sqlite"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	statements := []string{
		"CREATE TABLE moz_historyvisits (" +
			"id INTEGER PRIMARY KEY, " +
			"place_id INTEGER NOT NULL, " +
			"visit_date INTEGER NOT NULL" +
			")",
		"CREATE TABLE moz_places (" +
			"id INTEGER PRIMARY KEY, " +
			"url TEXT NOT NULL, " +
			"title TEXT, " +
			"visit_count INTEGER NOT NULL, " +
			"last_visit_date INTEGER, " +
			"hidden INTEGER" +
			")",
		"INSERT INTO moz_places" +
			"(id, url, title, visit_count, last_visit_date, hidden) " +
			"VALUES (9, 'https://example.com/full?q=1#f', NULL, 2, 1710000000123456, 0)",
		"INSERT INTO moz_historyvisits(id, place_id, visit_date) VALUES (201, 9, 1710000000123456)",
		"INSERT INTO moz_historyvisits(id, place_id, visit_date) VALUES (202, 9, 1710000001123456)",
	}
	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			t.Fatalf("database.Exec(%q) error = %v", statement, err)
		}
	}
	if err := database.Close(); err != nil {
		t.Fatalf("database.Close() error = %v", err)
	}

	observations, err := FindVisitObservations(profile, nil)
	if err != nil {
		t.Fatalf("FindVisitObservations() error = %v", err)
	}
	if len(observations) != 2 {
		t.Fatalf("FindVisitObservations() returned %d rows, want 2", len(observations))
	}
	if observations[0].NativeVisitID != 201 || observations[1].NativeVisitID != 202 {
		t.Errorf(
			"native visit IDs = %d, %d; want 201, 202",
			observations[0].NativeVisitID,
			observations[1].NativeVisitID,
		)
	}
	if observations[0].ObservationID == observations[1].ObservationID {
		t.Error("two moz_historyvisits rows produced the same obs_id")
	}

	pages, err := FindHistoryPages(profile, []int64{9})
	if err != nil {
		t.Fatalf("FindHistoryPages() error = %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("FindHistoryPages() returned %d rows, want 1", len(pages))
	}
	if pages[0].TitlePresent {
		t.Errorf("NULL Firefox title was marked present: %#v", pages[0])
	}
	if pages[0].URL != "https://example.com/full?q=1#f" {
		t.Errorf("page URL = %q, full URL was not preserved", pages[0].URL)
	}
}

func firefoxFixtureProfile(t *testing.T) common.Profile {
	t.Helper()
	path, err := common.CanonicalProfilePath(t.TempDir())
	if err != nil {
		t.Fatalf("CanonicalProfilePath() error = %v", err)
	}
	return common.Profile{
		ProfileID:      common.GenerateProfileID("firefox", path),
		BrowserFamily:  "firefox",
		BrowserVariant: "firefox",
		Path:           path,
	}
}
