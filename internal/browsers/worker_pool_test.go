package browsers

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"osquery-extension-browsers/internal/browsers/chromium"
	"osquery-extension-browsers/internal/browsers/common"
	"osquery-extension-browsers/internal/browsers/firefox"
)

// TestWorkerPoolPerformance tests the performance characteristics of the worker pool
func TestWorkerPoolPerformance(t *testing.T) {
	t.Run("firefox_worker_pool_performance", func(t *testing.T) {
		start := time.Now()

		// Test Firefox worker pool by calling the main function
		// (which internally uses the worker pool)
		paths := firefox.FindFirefoxPaths()

		duration := time.Since(start)

		t.Logf("Firefox worker pool completed in %v, found %d paths", duration, len(paths))

		// Should complete reasonably quickly even with many users
		if duration > 30*time.Second {
			t.Errorf("Firefox worker pool took too long: %v", duration)
		}
	})

	t.Run("chromium_worker_pool_performance", func(t *testing.T) {
		start := time.Now()

		// Test Chromium worker pool by calling the main function
		paths := chromium.FindChromiumPaths()

		duration := time.Since(start)

		t.Logf("Chromium worker pool completed in %v, found %d paths", duration, len(paths))

		// Should complete reasonably quickly even with many users
		if duration > 30*time.Second {
			t.Errorf("Chromium worker pool took too long: %v", duration)
		}
	})
}

// TestWorkerPoolConcurrency tests that the worker pool handles concurrent access correctly
func TestWorkerPoolConcurrency(t *testing.T) {
	const numGoroutines = 10
	const numIterations = 3

	t.Run("concurrent_firefox_worker_pool", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numIterations)
		results := make(chan int, numGoroutines*numIterations) // Store path counts

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					paths := firefox.FindFirefoxPaths()
					if paths == nil {
						errors <- nil // Signal error
						return
					}
					results <- len(paths)
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		close(results)

		// Check for errors
		errorCount := 0
		for range errors {
			errorCount++
		}

		if errorCount > 0 {
			t.Errorf("Found %d errors in concurrent Firefox worker pool test", errorCount)
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

	t.Run("concurrent_chromium_worker_pool", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numIterations)
		results := make(chan int, numGoroutines*numIterations)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					paths := chromium.FindChromiumPaths()
					if paths == nil {
						errors <- nil // Signal error
						return
					}
					results <- len(paths)
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		close(results)

		// Check for errors
		errorCount := 0
		for range errors {
			errorCount++
		}

		if errorCount > 0 {
			t.Errorf("Found %d errors in concurrent Chromium worker pool test", errorCount)
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

// TestWorkerPoolResourceManagement tests that the worker pool manages resources properly
func TestWorkerPoolResourceManagement(t *testing.T) {
	t.Run("worker_count_limits", func(t *testing.T) {
		// This test verifies that the worker pool doesn't create excessive goroutines
		// We can't directly test the internal worker count, but we can test that
		// the system remains responsive and doesn't exhaust resources

		const iterations = 50
		start := time.Now()

		for i := 0; i < iterations; i++ {
			firefoxPaths := firefox.FindFirefoxPaths()
			chromiumPaths := chromium.FindChromiumPaths()

			// Verify results are valid
			if firefoxPaths == nil || chromiumPaths == nil {
				t.Error("Worker pool returned nil results")
			}
		}

		duration := time.Since(start)
		t.Logf("Completed %d iterations in %v", iterations, duration)

		// Should complete without hanging or excessive resource usage
		if duration > 60*time.Second {
			t.Errorf("Worker pool resource management test took too long: %v", duration)
		}
	})

	t.Run("memory_usage_stability", func(t *testing.T) {
		// Test that repeated calls don't cause memory leaks
		runtime.GC() // Force garbage collection before test

		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Perform many operations
		for i := 0; i < 100; i++ {
			paths := firefox.FindFirefoxPaths()
			_ = paths // Prevent optimization

			if i%10 == 0 {
				runtime.GC() // Periodic garbage collection
			}
		}

		runtime.GC() // Force garbage collection after test
		runtime.ReadMemStats(&m2)

		// Calculate memory growth safely
		var memGrowth uint64
		if m2.Alloc > m1.Alloc {
			memGrowth = m2.Alloc - m1.Alloc
		} else {
			memGrowth = 0 // Memory actually decreased
		}

		t.Logf("Memory growth: %d bytes (from %d to %d)", memGrowth, m1.Alloc, m2.Alloc)

		// Allow some growth but not excessive (adjust threshold as needed)
		if memGrowth > 10*1024*1024 { // 10MB threshold
			t.Errorf("Excessive memory growth detected: %d bytes", memGrowth)
		}
	})
}

// TestWorkerPoolErrorHandling tests error handling in the worker pool
func TestWorkerPoolErrorHandling(t *testing.T) {
	t.Run("handles_user_enumeration_errors", func(t *testing.T) {
		// Test that worker pool handles cases where user enumeration fails
		// This is tested indirectly by ensuring the functions still work

		firefoxPaths := firefox.FindFirefoxPaths()
		chromiumPaths := chromium.FindChromiumPaths()

		// Should not return nil even if user enumeration has issues
		if firefoxPaths == nil {
			t.Error("Firefox worker pool returned nil during error handling test")
		}
		if chromiumPaths == nil {
			t.Error("Chromium worker pool returned nil during error handling test")
		}
	})

	t.Run("handles_timeout_scenarios", func(t *testing.T) {
		// Test that worker pool completes within reasonable time
		done := make(chan bool, 1)

		go func() {
			firefoxPaths := firefox.FindFirefoxPaths()
			chromiumPaths := chromium.FindChromiumPaths()

			// Verify results
			if firefoxPaths == nil || chromiumPaths == nil {
				t.Error("Worker pool returned nil in timeout test")
			}

			done <- true
		}()

		select {
		case <-done:
			// Success - completed within reasonable time
		case <-time.After(45 * time.Second):
			t.Error("Worker pool timed out - possible deadlock or hanging")
		}
	})
}

// TestWorkerPoolScalability tests how the worker pool scales with different user counts
func TestWorkerPoolScalability(t *testing.T) {
	// Get actual users for realistic testing
	users, err := common.UsersFromContext()
	if err != nil {
		t.Logf("User enumeration error (may be expected): %v", err)
	}

	userCount := len(users)
	t.Logf("Testing scalability with %d users", userCount)

	t.Run("scales_with_user_count", func(t *testing.T) {
		measurements := make(map[string]time.Duration)

		// Test Firefox scaling
		start := time.Now()
		firefoxPaths := firefox.FindFirefoxPaths()
		measurements["firefox"] = time.Since(start)

		// Test Chromium scaling
		start = time.Now()
		chromiumPaths := chromium.FindChromiumPaths()
		measurements["chromium"] = time.Since(start)

		t.Logf("Firefox: %v (%d paths), Chromium: %v (%d paths)",
			measurements["firefox"], len(firefoxPaths),
			measurements["chromium"], len(chromiumPaths))

		// Performance should be reasonable regardless of user count
		for browser, duration := range measurements {
			if duration > 20*time.Second {
				t.Errorf("%s worker pool scaling test took too long: %v", browser, duration)
			}
		}
	})
}

// Helper function to create mock users for testing
func createMockUsers(count int) []common.UserInfo {
	users := make([]common.UserInfo, count)
	for i := 0; i < count; i++ {
		users[i] = common.UserInfo{
			Username:     "testuser" + string(rune('0'+i%10)),
			HomeDir:      "/home/testuser" + string(rune('0'+i%10)),
			UID:          string(rune('1'+i%10)) + "000",
			IsAccessible: i%3 != 0, // Make some users inaccessible for realistic testing
		}
	}
	return users
}
