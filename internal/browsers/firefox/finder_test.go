package firefox

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"osquery-extension-browsers/internal/browsers/common"
)

func TestFindFirefoxPaths(t *testing.T) {
	// Test the actual function behavior
	paths := FindFirefoxPaths()

	// Verify we get a slice (even if empty)
	if paths == nil {
		t.Error("FindFirefoxPaths() returned nil, expected slice")
	}

	// Verify all returned paths are absolute
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("FindFirefoxPaths() returned relative path: %s", path)
		}
	}

	// Test should pass regardless of system state
	t.Logf("FindFirefoxPaths() returned %d paths", len(paths))
}

func TestFindFirefoxPathsForUser(t *testing.T) {
	tests := []struct {
		name     string
		user     common.UserInfo
		expected []string
	}{
		{
			name: "windows user",
			user: common.UserInfo{
				Username: "testuser",
				HomeDir:  "C:\\Users\\testuser",
				UID:      "1001",
			},
			expected: []string{
				"C:\\Users\\testuser\\AppData\\Roaming\\Mozilla\\Firefox\\Profiles",
			},
		},
		{
			name: "darwin user",
			user: common.UserInfo{
				Username: "testuser",
				HomeDir:  "/Users/testuser",
				UID:      "501",
			},
			expected: []string{
				"/Users/testuser/Library/Application Support/Firefox/Profiles",
				"/Users/testuser/Library/Application Support/zen/Profiles",
				"/Users/testuser/Library/Application Support/Floorp/Profiles",
			},
		},
		{
			name: "linux user",
			user: common.UserInfo{
				Username: "testuser",
				HomeDir:  "/home/testuser",
				UID:      "1001",
			},
			expected: []string{
				"/home/testuser/.mozilla/firefox",
				"/home/testuser/.zen",
				"/home/testuser/.var/app/app.zen_browser.zen/.zen",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip test if not on the expected OS
			switch tt.name {
			case "windows user":
				if runtime.GOOS != "windows" {
					t.Skip("Skipping Windows test on non-Windows OS")
				}
			case "darwin user":
				if runtime.GOOS != "darwin" {
					t.Skip("Skipping macOS test on non-macOS OS")
				}
			case "linux user":
				if runtime.GOOS != "linux" {
					t.Skip("Skipping Linux test on non-Linux OS")
				}
			}

			paths := findFirefoxPathsForUser(tt.user)

			// Since we're testing path generation, not existence,
			// we need to check the generated paths match expected patterns
			expectedPaths := getExpectedPathsForOS(tt.user)

			if len(paths) > len(expectedPaths) {
				t.Errorf("findFirefoxPathsForUser() returned more paths than expected")
			}

			// Verify each returned path is in the expected list
			for _, path := range paths {
				found := false
				for _, expected := range expectedPaths {
					if path == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected path returned: %s", path)
				}
			}
		})
	}
}

func TestFindFirefoxPathsErrorHandling(t *testing.T) {
	// Test with inaccessible user
	inaccessibleUser := common.UserInfo{
		Username:     "inaccessible",
		HomeDir:      "/nonexistent/path",
		UID:          "9999",
		IsAccessible: false,
	}

	paths := findFirefoxPathsForUser(inaccessibleUser)

	// Should return empty slice for inaccessible directories
	if len(paths) > 0 {
		t.Errorf("Expected no paths for inaccessible user, got %d", len(paths))
	}
}

func TestFindFirefoxPathsConcurrency(t *testing.T) {
	// Test that concurrent calls don't cause issues
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			paths := FindFirefoxPaths()
			if paths == nil {
				t.Error("Concurrent call returned nil")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper functions for testing

func getExpectedPathsForOS(user common.UserInfo) []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		appData := filepath.Join(user.HomeDir, "AppData", "Roaming")
		paths = append(paths, filepath.Join(appData, "Mozilla", "Firefox", "Profiles"))
	case "darwin":
		paths = append(paths, filepath.Join(user.HomeDir, "Library", "Application Support", "Firefox", "Profiles"))
		paths = append(paths, filepath.Join(user.HomeDir, "Library", "Application Support", "zen", "Profiles"))
		paths = append(paths, filepath.Join(user.HomeDir, "Library", "Application Support", "Floorp", "Profiles"))
	default:
		paths = append(paths, filepath.Join(user.HomeDir, ".mozilla", "firefox"))
		paths = append(paths, filepath.Join(user.HomeDir, ".zen"))
		paths = append(paths, filepath.Join(user.HomeDir, ".var", "app", "app.zen_browser.zen", ".zen"))
	}

	return paths
}

func containsWindowsFirefoxPath(path string) bool {
	return strings.Contains(path, "Mozilla\\Firefox") ||
		strings.Contains(path, "AppData\\Roaming")
}

func containsDarwinFirefoxPath(path string) bool {
	return strings.Contains(path, "Library/Application Support/Firefox") ||
		strings.Contains(path, "Library/Application Support/zen") ||
		strings.Contains(path, "Library/Application Support/Floorp") ||
		strings.Contains(path, "Application Support")
}

func containsLinuxFirefoxPath(path string) bool {
	return strings.Contains(path, ".mozilla/firefox") ||
		strings.Contains(path, ".zen") ||
		strings.Contains(path, "zen_browser")
}
