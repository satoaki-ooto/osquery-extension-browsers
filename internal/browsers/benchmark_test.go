package browsers

import (
	"testing"

	"osquery-extension-browsers/internal/browsers/chromium"
	"osquery-extension-browsers/internal/browsers/common"
	"osquery-extension-browsers/internal/browsers/firefox"
)

// BenchmarkUserEnumeration benchmarks the user enumeration process
func BenchmarkUserEnumeration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := common.UsersFromContext()
		if err != nil {
			b.Logf("User enumeration error (may be expected): %v", err)
		}
		_ = users // Prevent optimization
	}
}

// BenchmarkFirefoxPathDetection benchmarks Firefox path detection
func BenchmarkFirefoxPathDetection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		paths := firefox.FindFirefoxPaths()
		_ = paths // Prevent optimization
	}
}

// BenchmarkChromiumPathDetection benchmarks Chromium path detection
func BenchmarkChromiumPathDetection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		paths := chromium.FindChromiumPaths()
		_ = paths // Prevent optimization
	}
}

// BenchmarkCombinedBrowserDetection benchmarks both Firefox and Chromium detection together
func BenchmarkCombinedBrowserDetection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		firefoxPaths := firefox.FindFirefoxPaths()
		chromiumPaths := chromium.FindChromiumPaths()
		_ = firefoxPaths  // Prevent optimization
		_ = chromiumPaths // Prevent optimization
	}
}

// BenchmarkParallelFirefoxDetection benchmarks parallel Firefox detection
func BenchmarkParallelFirefoxDetection(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			paths := firefox.FindFirefoxPaths()
			_ = paths // Prevent optimization
		}
	})
}

// BenchmarkParallelChromiumDetection benchmarks parallel Chromium detection
func BenchmarkParallelChromiumDetection(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			paths := chromium.FindChromiumPaths()
			_ = paths // Prevent optimization
		}
	})
}

// BenchmarkParallelCombinedDetection benchmarks parallel combined browser detection
func BenchmarkParallelCombinedDetection(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			firefoxPaths := firefox.FindFirefoxPaths()
			chromiumPaths := chromium.FindChromiumPaths()
			_ = firefoxPaths  // Prevent optimization
			_ = chromiumPaths // Prevent optimization
		}
	})
}

// BenchmarkUserEnumerationMemory benchmarks memory usage of user enumeration
func BenchmarkUserEnumerationMemory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := common.UsersFromContext()
		if err != nil {
			b.Logf("User enumeration error (may be expected): %v", err)
		}
		_ = users // Prevent optimization
	}
}

// BenchmarkFirefoxDetectionMemory benchmarks memory usage of Firefox detection
func BenchmarkFirefoxDetectionMemory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		paths := firefox.FindFirefoxPaths()
		_ = paths // Prevent optimization
	}
}

// BenchmarkChromiumDetectionMemory benchmarks memory usage of Chromium detection
func BenchmarkChromiumDetectionMemory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		paths := chromium.FindChromiumPaths()
		_ = paths // Prevent optimization
	}
}

// BenchmarkScalabilityWithUsers benchmarks performance with different numbers of users
func BenchmarkScalabilityWithUsers(b *testing.B) {
	// Get actual users for realistic testing
	users, err := common.UsersFromContext()
	if err != nil || len(users) == 0 {
		b.Skip("No users available for scalability testing")
	}

	b.Run("single_user", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate single user by limiting to first user
			if len(users) > 0 {
				paths := firefox.FindFirefoxPaths()
				_ = paths
			}
		}
	})

	if len(users) > 1 {
		b.Run("multiple_users", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use all available users
				paths := firefox.FindFirefoxPaths()
				_ = paths
			}
		})
	}
}

// BenchmarkConcurrentUserScanning benchmarks the concurrent user scanning implementation
func BenchmarkConcurrentUserScanning(b *testing.B) {
	users, err := common.UsersFromContext()
	if err != nil || len(users) == 0 {
		b.Skip("No users available for concurrent scanning benchmark")
	}

	b.Run("firefox_concurrent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			paths := firefox.FindFirefoxPaths()
			_ = paths
		}
	})

	b.Run("chromium_concurrent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			paths := chromium.FindChromiumPaths()
			_ = paths
		}
	})
}

// BenchmarkMemoryEfficiency tests memory efficiency of the multi-user implementation
func BenchmarkMemoryEfficiency(b *testing.B) {
	b.Run("firefox_memory_efficiency", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			paths := firefox.FindFirefoxPaths()

			// Simulate processing the paths to measure realistic memory usage
			for _, path := range paths {
				_ = len(path) // Simple processing to prevent optimization
			}
		}
	})

	b.Run("chromium_memory_efficiency", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			paths := chromium.FindChromiumPaths()

			// Simulate processing the paths to measure realistic memory usage
			for _, path := range paths {
				_ = len(path) // Simple processing to prevent optimization
			}
		}
	})
}

// BenchmarkWorstCaseScenario benchmarks performance in worst-case scenarios
func BenchmarkWorstCaseScenario(b *testing.B) {
	b.Run("repeated_calls", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate repeated calls that might happen in real usage
			for j := 0; j < 5; j++ {
				firefoxPaths := firefox.FindFirefoxPaths()
				chromiumPaths := chromium.FindChromiumPaths()
				_ = firefoxPaths
				_ = chromiumPaths
			}
		}
	})

	b.Run("high_concurrency", func(b *testing.B) {
		b.SetParallelism(100) // High concurrency
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				firefoxPaths := firefox.FindFirefoxPaths()
				chromiumPaths := chromium.FindChromiumPaths()
				_ = firefoxPaths
				_ = chromiumPaths
			}
		})
	})
}
