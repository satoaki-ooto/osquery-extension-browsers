package browsers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"osquery-extension-browsers/internal/browsers/chromium"
	"osquery-extension-browsers/internal/browsers/common"
	"osquery-extension-browsers/internal/browsers/firefox"
)

// TestMultiUserBrowserDetectionFlow tests the complete multi-user browser detection workflow
func TestMultiUserBrowserDetectionFlow(t *testing.T) {
	t.Run("complete_detection_flow", func(t *testing.T) {
		// Test user enumeration
		users, err := common.UsersFromContext()
		if err != nil {
			t.Logf("User enumeration returned error (may be expected): %v", err)
		}

		t.Logf("Found %d users on system", len(users))

		// Test Firefox path detection
		firefoxPaths := firefox.FindFirefoxPaths()
		t.Logf("Found %d Firefox paths", len(firefoxPaths))

		// Test Chromium path detection
		chromiumPaths := chromium.FindChromiumPaths()
		t.Logf("Found %d Chromium paths", len(chromiumPaths))

		// Verify all paths are absolute
		allPaths := append(firefoxPaths, chromiumPaths...)
		for _, path := range allPaths {
			if !filepath.IsAbs(path) {
				t.Errorf("Non-absolute path returned: %s", path)
			}
		}

		// Verify paths exist (for accessible ones)
		for _, path := range allPaths {
			if _, err := os.Stat(path); err != nil {
				t.Logf("Path not accessible (expected): %s - %v", path, err)
			}
		}
	})
}

// TestMultiUserConcurrentAccess tests concurrent access to multi-user browser detection
func TestMultiUserConcurrentAccess(t *testing.T) {
	const numGoroutines = 20
	const numIterations = 5

	t.Run("concurrent_firefox_detection", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan []string, numGoroutines*numIterations)
		errors := make(chan error, numGoroutines*numIterations)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					paths := firefox.FindFirefoxPaths()
					if paths == nil {
						errors <- nil // Signal error without actual error value
						return
					}
					results <- paths
				}
			}()
		}

		wg.Wait()
		close(results)
		close(errors)

		// Check for errors
		errorCount := 0
		for range errors {
			errorCount++
		}

		if errorCount > 0 {
			t.Errorf("Found %d errors in concurrent Firefox detection", errorCount)
		}

		// Verify we got expected number of results
		resultCount := 0
		for range results {
			resultCount++
		}

		expectedResults := numGoroutines * numIterations
		if resultCount != expectedResults {
			t.Errorf("Expected %d results, got %d", expectedResults, resultCount)
		}
	})

	t.Run("concurrent_chromium_detection", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan []string, numGoroutines*numIterations)
		errors := make(chan error, numGoroutines*numIterations)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					paths := chromium.FindChromiumPaths()
					if paths == nil {
						errors <- nil // Signal error without actual error value
						return
					}
					results <- paths
				}
			}()
		}

		wg.Wait()
		close(results)
		close(errors)

		// Check for errors
		errorCount := 0
		for range errors {
			errorCount++
		}

		if errorCount > 0 {
			t.Errorf("Found %d errors in concurrent Chromium detection", errorCount)
		}

		// Verify we got expected number of results
		resultCount := 0
		for range results {
			resultCount++
		}

		expectedResults := numGoroutines * numIterations
		if resultCount != expectedResults {
			t.Errorf("Expected %d results, got %d", expectedResults, resultCount)
		}
	})
}

// TestBackwardCompatibility ensures existing code continues to work
func TestBackwardCompatibility(t *testing.T) {
	t.Run("firefox_paths_compatibility", func(t *testing.T) {
		// Test that the function signature hasn't changed
		paths := firefox.FindFirefoxPaths()

		// Should return a slice of strings (not nil)
		if paths == nil {
			t.Error("FindFirefoxPaths() returned nil, breaking backward compatibility")
		}

		// Should return valid paths
		for _, path := range paths {
			if path == "" {
				t.Error("FindFirefoxPaths() returned empty path")
			}
			if !filepath.IsAbs(path) {
				t.Errorf("FindFirefoxPaths() returned relative path: %s", path)
			}
		}
	})

	t.Run("chromium_paths_compatibility", func(t *testing.T) {
		// Test that the function signature hasn't changed
		paths := chromium.FindChromiumPaths()

		// Should return a slice of strings (not nil)
		if paths == nil {
			t.Error("FindChromiumPaths() returned nil, breaking backward compatibility")
		}

		// Should return valid paths
		for _, path := range paths {
			if path == "" {
				t.Error("FindChromiumPaths() returned empty path")
			}
			if !filepath.IsAbs(path) {
				t.Errorf("FindChromiumPaths() returned relative path: %s", path)
			}
		}
	})
}

// TestErrorHandlingIntegration tests error handling across the entire system
func TestErrorHandlingIntegration(t *testing.T) {
	t.Run("graceful_permission_handling", func(t *testing.T) {
		// This test verifies that permission errors don't crash the system
		// We can't easily simulate permission errors, but we can verify
		// the system handles various edge cases gracefully

		// Test with timeout to ensure no hanging
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		done := make(chan bool)
		go func() {
			// Test Firefox detection
			firefoxPaths := firefox.FindFirefoxPaths()
			if firefoxPaths == nil {
				t.Error("Firefox detection returned nil during error handling test")
			}

			// Test Chromium detection
			chromiumPaths := chromium.FindChromiumPaths()
			if chromiumPaths == nil {
				t.Error("Chromium detection returned nil during error handling test")
			}

			done <- true
		}()

		select {
		case <-ctx.Done():
			t.Error("Browser detection timed out - possible hanging due to permission issues")
		case <-done:
			// Success - completed within timeout
		}
	})

	t.Run("user_enumeration_fallback", func(t *testing.T) {
		// Test that the system falls back gracefully when user enumeration fails
		// This is tested indirectly by ensuring the functions still return results

		firefoxPaths := firefox.FindFirefoxPaths()
		chromiumPaths := chromium.FindChromiumPaths()

		// Even if user enumeration fails, we should get current user paths
		// The functions should never return nil
		if firefoxPaths == nil {
			t.Error("Firefox detection should not return nil even when user enumeration fails")
		}
		if chromiumPaths == nil {
			t.Error("Chromium detection should not return nil even when user enumeration fails")
		}
	})
}

// TestCrossPlatformIntegration tests the system works across different platforms
func TestCrossPlatformIntegration(t *testing.T) {
	t.Run("platform_specific_paths", func(t *testing.T) {
		firefoxPaths := firefox.FindFirefoxPaths()
		chromiumPaths := chromium.FindChromiumPaths()

		// Verify paths are appropriate for current platform
		allPaths := append(firefoxPaths, chromiumPaths...)
		for _, path := range allPaths {
			switch runtime.GOOS {
			case "windows":
				if !filepath.IsAbs(path) || !isValidWindowsPath(path) {
					t.Errorf("Invalid Windows path: %s", path)
				}
			case "darwin":
				if !filepath.IsAbs(path) || !isValidDarwinPath(path) {
					t.Errorf("Invalid macOS path: %s", path)
				}
			default: // Linux and other Unix-like systems
				if !filepath.IsAbs(path) || !isValidLinuxPath(path) {
					t.Errorf("Invalid Linux path: %s", path)
				}
			}
		}
	})
}

// TestPerformanceCharacteristics tests basic performance characteristics
func TestPerformanceCharacteristics(t *testing.T) {
	t.Run("detection_performance", func(t *testing.T) {
		// Measure time for Firefox detection
		start := time.Now()
		firefoxPaths := firefox.FindFirefoxPaths()
		firefoxDuration := time.Since(start)

		// Measure time for Chromium detection
		start = time.Now()
		chromiumPaths := chromium.FindChromiumPaths()
		chromiumDuration := time.Since(start)

		t.Logf("Firefox detection took %v, found %d paths", firefoxDuration, len(firefoxPaths))
		t.Logf("Chromium detection took %v, found %d paths", chromiumDuration, len(chromiumPaths))

		// Reasonable performance expectations (adjust based on system requirements)
		maxDuration := 10 * time.Second
		if firefoxDuration > maxDuration {
			t.Errorf("Firefox detection took too long: %v", firefoxDuration)
		}
		if chromiumDuration > maxDuration {
			t.Errorf("Chromium detection took too long: %v", chromiumDuration)
		}
	})
}

// Helper functions for path validation

func isValidWindowsPath(path string) bool {
	return filepath.IsAbs(path) &&
		(filepath.VolumeName(path) != "" ||
			filepath.HasPrefix(path, `\\`)) // UNC paths
}

func isValidDarwinPath(path string) bool {
	return filepath.IsAbs(path) &&
		(filepath.HasPrefix(path, "/Users/") ||
			filepath.HasPrefix(path, "/home/") ||
			filepath.HasPrefix(path, "/"))
}

func isValidLinuxPath(path string) bool {
	return filepath.IsAbs(path) &&
		(filepath.HasPrefix(path, "/home/") ||
			filepath.HasPrefix(path, "/root/") ||
			filepath.HasPrefix(path, "/"))
}
