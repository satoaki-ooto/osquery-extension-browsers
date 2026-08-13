package chromium

import (
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"osquery-extension-browsers/internal/browsers/common"
)

func TestVisitObservationsAndHistoryPages(t *testing.T) {
	profile := chromiumFixtureProfile(t)
	database := openChromiumFixture(t, filepath.Join(profile.Path, "History"))

	pageTimeUS := int64(1710000000123456)
	pageTime := webkitEpochOffsetMicroseconds + pageTimeUS
	if _, err := database.Exec(
		"INSERT INTO urls"+
			"(id, url, title, visit_count, last_visit_time, hidden) "+
			"VALUES (?, ?, ?, ?, ?, ?)",
		7,
		"https://example.com/a/b?secret=value#fragment",
		"Example",
		2,
		pageTime,
		0,
	); err != nil {
		t.Fatalf("insert Chromium page error = %v", err)
	}
	for visitID, delta := range map[int64]int64{101: 0, 102: 10} {
		if _, err := database.Exec(
			"INSERT INTO visits(id, url, visit_time) VALUES (?, ?, ?)",
			visitID,
			7,
			pageTime+delta,
		); err != nil {
			t.Fatalf("insert Chromium visit error = %v", err)
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
	if observations[0].ObservationID == observations[1].ObservationID {
		t.Error("two native visit IDs produced the same obs_id")
	}
	for _, observation := range observations {
		if observation.NativeURLID != 7 || observation.VisitTimeUS < pageTimeUS {
			t.Errorf("unexpected observation: %#v", observation)
		}
	}
	repeatedObservations, err := FindVisitObservations(profile, nil)
	if err != nil {
		t.Fatalf("second FindVisitObservations() error = %v", err)
	}
	if !reflect.DeepEqual(observations, repeatedObservations) {
		t.Errorf(
			"unchanged scan changed observations:\nfirst=%#v\nsecond=%#v",
			observations,
			repeatedObservations,
		)
	}

	pages, err := FindHistoryPages(profile, nil)
	if err != nil {
		t.Fatalf("FindHistoryPages() error = %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("FindHistoryPages() returned %d rows, want 1", len(pages))
	}
	page := pages[0]
	if page.URL != "https://example.com/a/b?secret=value#fragment" {
		t.Errorf("page URL = %q, full URL was not preserved", page.URL)
	}
	if !page.TitlePresent || page.Title != "Example" || page.LastVisitTime != pageTimeUS/1000000 {
		t.Errorf("unexpected page metadata: %#v", page)
	}
}

func TestExistingUnreadableHistoryReturnsContext(t *testing.T) {
	profile := chromiumFixtureProfile(t)
	if err := os.Mkdir(filepath.Join(profile.Path, "History"), 0o755); err != nil {
		t.Fatalf("os.Mkdir(History) error = %v", err)
	}

	_, err := FindVisitObservations(profile, nil)
	if err == nil {
		t.Fatal("FindVisitObservations() unexpectedly succeeded for a directory database")
	}
	for _, expected := range []string{profile.BrowserVariant, profile.Path, "is a directory"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("error %q does not contain %q", err, expected)
		}
	}
}

func chromiumFixtureProfile(t *testing.T) common.Profile {
	t.Helper()
	path, err := common.CanonicalProfilePath(t.TempDir())
	if err != nil {
		t.Fatalf("CanonicalProfilePath() error = %v", err)
	}
	return common.Profile{
		ProfileID:      common.GenerateProfileID("chrome", path),
		BrowserFamily:  "chromium",
		BrowserVariant: "chrome",
		Path:           path,
	}
}

func openChromiumFixture(t *testing.T, databasePath string) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	statements := []string{
		"CREATE TABLE visits (id INTEGER PRIMARY KEY, url INTEGER NOT NULL, visit_time INTEGER NOT NULL)",
		"CREATE TABLE urls (" +
			"id INTEGER PRIMARY KEY, " +
			"url TEXT NOT NULL, " +
			"title TEXT, " +
			"visit_count INTEGER NOT NULL, " +
			"last_visit_time INTEGER, " +
			"hidden INTEGER" +
			")",
	}
	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			t.Fatalf("database.Exec(%q) error = %v", statement, err)
		}
	}
	return database
}
