package firefox

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"osquery-extension-browsers/internal/browsers/common"
)

func TestFindHistory(t *testing.T) {
	t.Run("missing_places_sqlite_returns_empty_slice", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := ioutil.TempDir("", "firefox_test_")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a profile with a non-existent places.sqlite
		profile := common.Profile{
			ID:             "test-profile",
			Name:           "Test Profile",
			Path:           tempDir,
			BrowserType:    "firefox",
			BrowserVariant: "firefox",
		}

		// Call FindHistory - should return empty slice with no error
		historyEntries, err := FindHistory(profile)

		// Verify no error occurred
		if err != nil {
			t.Errorf("Expected no error for missing places.sqlite, got: %v", err)
		}

		// Verify empty slice is returned
		if historyEntries == nil {
			t.Error("Expected empty slice, got nil")
		}

		if len(historyEntries) != 0 {
			t.Errorf("Expected empty slice, got %d entries", len(historyEntries))
		}
	})

	t.Run("existing_empty_database_works", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := ioutil.TempDir("", "firefox_test_")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create an empty SQLite database file
		dbPath := filepath.Join(tempDir, "places.sqlite")
		file, err := os.Create(dbPath)
		if err != nil {
			t.Fatalf("Failed to create empty database file: %v", err)
		}
		file.Close()

		// Create a profile
		profile := common.Profile{
			ID:             "test-profile",
			Name:           "Test Profile",
			Path:           tempDir,
			BrowserType:    "firefox",
			BrowserVariant: "firefox",
		}

		// Call FindHistory - should attempt to open the database
		// This will likely fail due to invalid database format, but that's expected
		historyEntries, err := FindHistory(profile)

		// We expect either an error (due to invalid database) or empty results
		// The important thing is that it didn't skip due to missing file
		if err != nil {
			// This is acceptable - empty file isn't a valid SQLite database
			t.Logf("Expected error for invalid database format: %v", err)
		} else if historyEntries == nil {
			t.Error("Expected empty slice or error, got nil")
		}
	})

	t.Run("inaccessible_directory_returns_empty_slice", func(t *testing.T) {
		// Test with a non-existent directory
		profile := common.Profile{
			ID:             "test-profile",
			Name:           "Test Profile",
			Path:           "/nonexistent/directory/path",
			BrowserType:    "firefox",
			BrowserVariant: "firefox",
		}

		// Call FindHistory - should return empty slice with no error
		historyEntries, err := FindHistory(profile)

		// Verify no error occurred
		if err != nil {
			t.Errorf("Expected no error for inaccessible directory, got: %v", err)
		}

		// Verify empty slice is returned
		if historyEntries == nil {
			t.Error("Expected empty slice, got nil")
		}

		if len(historyEntries) != 0 {
			t.Errorf("Expected empty slice, got %d entries", len(historyEntries))
		}
	})

	t.Run("concurrent_access_to_missing_database", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := ioutil.TempDir("", "firefox_test_")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		profile := common.Profile{
			ID:             "test-profile",
			Name:           "Test Profile",
			Path:           tempDir,
			BrowserType:    "firefox",
			BrowserVariant: "firefox",
		}

		// Test concurrent access
		const numGoroutines = 10
		results := make(chan []common.HistoryEntry, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				historyEntries, err := FindHistory(profile)
				results <- historyEntries
				errors <- err
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			entries := <-results
			err := <-errors

			if err != nil {
				t.Errorf("Concurrent access failed: %v", err)
			}

			if entries == nil {
				t.Error("Expected empty slice, got nil")
			}

			if len(entries) != 0 {
				t.Errorf("Expected empty slice, got %d entries", len(entries))
			}
		}
	})
}

func TestGetHistoryDBPath(t *testing.T) {
	tests := []struct {
		name        string
		profilePath string
		expected    string
	}{
		{
			name:        "unix_path",
			profilePath: "/home/user/.mozilla/firefox/profile",
			expected:    filepath.Join("/home/user/.mozilla/firefox/profile", "places.sqlite"),
		},
		{
			name:        "windows_path",
			profilePath: "C:\\Users\\user\\AppData\\Roaming\\Mozilla\\Firefox\\Profiles\\profile",
			expected:    filepath.Join("C:\\Users\\user\\AppData\\Roaming\\Mozilla\\Firefox\\Profiles\\profile", "places.sqlite"),
		},
		{
			name:        "relative_path",
			profilePath: "profile",
			expected:    filepath.Join("profile", "places.sqlite"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHistoryDBPath(tt.profilePath)
			if result != tt.expected {
				t.Errorf("getHistoryDBPath(%s) = %s, expected %s", tt.profilePath, result, tt.expected)
			}
		})
	}
}

func TestParseUnixTime(t *testing.T) {
	tests := []struct {
		name     string
		unixTime int64
		expected time.Time
	}{
		{
			name:     "zero_time",
			unixTime: 0,
			expected: time.Time{},
		},
		{
			name:     "firefox_timestamp",
			unixTime: 1640995200000000, // 2022-01-01 00:00:00 UTC in microseconds
			expected: time.Unix(1640995200, 0),
		},
		{
			name:     "current_time_microseconds",
			unixTime: 1640995200123456,                 // With microseconds
			expected: time.Unix(1640995200, 123456000), // Convert to nanoseconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseUnixTime(tt.unixTime)

			if tt.name == "zero_time" {
				if !result.IsZero() {
					t.Errorf("parseUnixTime(%d) should return zero time, got %v", tt.unixTime, result)
				}
			} else {
				// For non-zero times, check that the Unix seconds match
				if result.Unix() != tt.expected.Unix() {
					t.Errorf("parseUnixTime(%d) = %v, expected Unix seconds %d, got %d",
						tt.unixTime, result, tt.expected.Unix(), result.Unix())
				}
			}
		})
	}
}

func TestFindHistoryErrorHandling(t *testing.T) {
	t.Run("profile_with_zen_browser_variant", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := ioutil.TempDir("", "zen_test_")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a Zen browser profile
		profile := common.Profile{
			ID:             "zen-profile",
			Name:           "Zen Profile",
			Path:           tempDir,
			BrowserType:    "zen",
			BrowserVariant: "zen",
		}

		// Call FindHistory - should return empty slice with no error
		historyEntries, err := FindHistory(profile)

		// Verify no error occurred
		if err != nil {
			t.Errorf("Expected no error for missing places.sqlite in Zen profile, got: %v", err)
		}

		// Verify empty slice is returned
		if historyEntries == nil {
			t.Error("Expected empty slice, got nil")
		}

		if len(historyEntries) != 0 {
			t.Errorf("Expected empty slice, got %d entries", len(historyEntries))
		}
	})

	t.Run("profile_with_floorp_variant", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := ioutil.TempDir("", "floorp_test_")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a Floorp profile
		profile := common.Profile{
			ID:             "floorp-profile",
			Name:           "Floorp Profile",
			Path:           tempDir,
			BrowserType:    "floorp",
			BrowserVariant: "floorp",
		}

		// Call FindHistory - should return empty slice with no error
		historyEntries, err := FindHistory(profile)

		// Verify no error occurred
		if err != nil {
			t.Errorf("Expected no error for missing places.sqlite in Floorp profile, got: %v", err)
		}

		// Verify empty slice is returned
		if historyEntries == nil {
			t.Error("Expected empty slice, got nil")
		}

		if len(historyEntries) != 0 {
			t.Errorf("Expected empty slice, got %d entries", len(historyEntries))
		}
	})

	t.Run("profile_with_empty_path", func(t *testing.T) {
		// Create a profile with empty path
		profile := common.Profile{
			ID:             "empty-path-profile",
			Name:           "Empty Path Profile",
			Path:           "",
			BrowserType:    "firefox",
			BrowserVariant: "firefox",
		}

		// Call FindHistory - should return empty slice with no error
		historyEntries, err := FindHistory(profile)

		// Verify no error occurred
		if err != nil {
			t.Errorf("Expected no error for empty path profile, got: %v", err)
		}

		// Verify empty slice is returned
		if historyEntries == nil {
			t.Error("Expected empty slice, got nil")
		}

		if len(historyEntries) != 0 {
			t.Errorf("Expected empty slice, got %d entries", len(historyEntries))
		}
	})
}
