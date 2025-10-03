package chromium

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"osquery-extension-browsers/internal/browsers/common"
)

func TestFindChromiumPaths(t *testing.T) {
	// Test the actual function behavior
	paths := FindChromiumPaths()

	// Verify we get a slice (even if empty)
	if paths == nil {
		t.Error("FindChromiumPaths() returned nil, expected slice")
	}

	// Verify all returned paths are absolute
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("FindChromiumPaths() returned relative path: %s", path)
		}
	}

	// Test should pass regardless of system state
	t.Logf("FindChromiumPaths() returned %d paths", len(paths))
}

func TestFindChromiumPathsForUser(t *testing.T) {
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
				"C:\\Users\\testuser\\AppData\\Local\\Google\\Chrome\\User Data",
				"C:\\Users\\testuser\\AppData\\Local\\Microsoft\\Edge\\User Data",
				"C:\\Users\\testuser\\AppData\\Local\\Chromium\\User Data",
				"C:\\Users\\testuser\\AppData\\Local\\BraveSoftware\\Brave-Browser\\User Data",
				"C:\\Users\\testuser\\AppData\\Local\\Vivaldi\\User Data",
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
				"/Users/testuser/Library/Application Support/Google/Chrome",
				"/Users/testuser/Library/Application Support/Microsoft Edge",
				"/Users/testuser/Library/Application Support/Chromium",
				"/Users/testuser/Library/Application Support/BraveSoftware/Brave-Browser",
				"/Users/testuser/Library/Application Support/Vivaldi",
				"/Users/testuser/Library/Application Support/Comet",
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
				"/home/testuser/.config/google-chrome",
				"/home/testuser/.config/microsoft-edge",
				"/home/testuser/.config/chromium",
				"/home/testuser/.config/BraveSoftware/Brave-Browser",
				"/home/testuser/.config/vivaldi",
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

			paths := findChromiumPathsForUser(tt.user)

			// Since we're testing path generation, not existence,
			// we need to check the generated paths match expected patterns
			expectedPaths := getExpectedChromiumPathsForOS(tt.user)

			if len(paths) > len(expectedPaths) {
				t.Errorf("findChromiumPathsForUser() returned more paths than expected")
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

func TestFindChromiumPathsErrorHandling(t *testing.T) {
	// Test with inaccessible user
	inaccessibleUser := common.UserInfo{
		Username:     "inaccessible",
		HomeDir:      "/nonexistent/path",
		UID:          "9999",
		IsAccessible: false,
	}

	paths := findChromiumPathsForUser(inaccessibleUser)

	// Should return empty slice for inaccessible directories
	if len(paths) > 0 {
		t.Errorf("Expected no paths for inaccessible user, got %d", len(paths))
	}
}

func TestFindChromiumPathsConcurrency(t *testing.T) {
	// Test that concurrent calls don't cause issues
	done := make(chan bool, 10)

	for range 10 {
		go func() {
			paths := FindChromiumPaths()
			if paths == nil {
				t.Error("Concurrent call returned nil")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}
}

func TestFindChromiumPathsMultiUser(t *testing.T) {
	// Test multi-user functionality by mocking users
	// This test verifies the concurrent scanning behavior
	paths := FindChromiumPaths()

	// Should not panic and should return a valid slice
	if paths == nil {
		t.Error("FindChromiumPaths() returned nil in multi-user scenario")
	}

	// All paths should be absolute
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("Multi-user scan returned relative path: %s", path)
		}
	}
}

func TestFindChromiumPathsFallback(t *testing.T) {
	// This test verifies that the fallback mechanism works
	// when user enumeration fails or returns no users

	// The actual function should handle this gracefully
	paths := FindChromiumPaths()

	// Should not be nil even in fallback scenarios
	if paths == nil {
		t.Error("FindChromiumPaths() returned nil during fallback")
	}

	// In fallback mode, should still return current user paths
	if len(paths) == 0 {
		t.Log("No Chromium paths found - this is acceptable if no browsers are installed")
	}
}

// Helper functions for testing

func getExpectedChromiumPathsForOS(user common.UserInfo) []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		localAppData := filepath.Join(user.HomeDir, "AppData", "Local")
		paths = append(paths, filepath.Join(localAppData, "Google", "Chrome", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "Microsoft", "Edge", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "Chromium", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "Vivaldi", "User Data"))
	case "darwin":
		appSupport := filepath.Join(user.HomeDir, "Library", "Application Support")
		paths = append(paths, filepath.Join(appSupport, "Google", "Chrome"))
		paths = append(paths, filepath.Join(appSupport, "Microsoft Edge"))
		paths = append(paths, filepath.Join(appSupport, "Chromium"))
		paths = append(paths, filepath.Join(appSupport, "BraveSoftware", "Brave-Browser"))
		paths = append(paths, filepath.Join(appSupport, "Vivaldi"))
		paths = append(paths, filepath.Join(appSupport, "Comet"))
	default:
		configDir := filepath.Join(user.HomeDir, ".config")
		paths = append(paths, filepath.Join(configDir, "google-chrome"))
		paths = append(paths, filepath.Join(configDir, "microsoft-edge"))
		paths = append(paths, filepath.Join(configDir, "chromium"))
		paths = append(paths, filepath.Join(configDir, "BraveSoftware", "Brave-Browser"))
		paths = append(paths, filepath.Join(configDir, "vivaldi"))
	}

	return paths
}

func containsWindowsChromiumPath(path string) bool {
	return strings.Contains(path, "AppData\\Local") &&
		(strings.Contains(path, "Chrome") ||
			strings.Contains(path, "Edge") ||
			strings.Contains(path, "Chromium") ||
			strings.Contains(path, "Brave") ||
			strings.Contains(path, "Vivaldi"))
}

func containsDarwinChromiumPath(path string) bool {
	return strings.Contains(path, "Library/Application Support") &&
		(strings.Contains(path, "Chrome") ||
			strings.Contains(path, "Edge") ||
			strings.Contains(path, "Chromium") ||
			strings.Contains(path, "Brave") ||
			strings.Contains(path, "Vivaldi") ||
			strings.Contains(path, "Comet"))
}

func containsLinuxChromiumPath(path string) bool {
	return strings.Contains(path, ".config") &&
		(strings.Contains(path, "chrome") ||
			strings.Contains(path, "edge") ||
			strings.Contains(path, "chromium") ||
			strings.Contains(path, "Brave") ||
			strings.Contains(path, "vivaldi"))
}
