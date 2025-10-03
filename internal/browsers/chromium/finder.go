package chromium

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"osquery-extension-browsers/internal/browsers/common"
)

// FindChromiumPaths returns the paths to Chromium-based browser data directories for all users
func FindChromiumPaths() []string {
	users, err := common.UsersFromContext()
	if err != nil || len(users) == 0 {
		return []string{}
	}

	// Filter accessible users
	var accessibleUsers []common.UserInfo
	for _, user := range users {
		if user.IsAccessible {
			accessibleUsers = append(accessibleUsers, user)
		}
	}

	if len(accessibleUsers) == 0 {
		return []string{}
	}

	// Use worker pool for better performance and resource management
	allPaths := scanUsersWithWorkerPool(accessibleUsers, findChromiumPathsForUser)

	return allPaths
}

// findChromiumPathsForUser returns Chromium-based browser paths for a specific user
func findChromiumPathsForUser(user common.UserInfo) []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Windows paths for Chromium-based browsers
		localAppData := filepath.Join(user.HomeDir, "AppData", "Local")
		paths = append(paths, filepath.Join(localAppData, "Google", "Chrome", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "Microsoft", "Edge", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "Chromium", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data"))
		paths = append(paths, filepath.Join(localAppData, "Vivaldi", "User Data"))

	case "darwin":
		// macOS paths for Chromium-based browsers
		appSupport := filepath.Join(user.HomeDir, "Library", "Application Support")
		paths = append(paths, filepath.Join(appSupport, "Google", "Chrome"))
		paths = append(paths, filepath.Join(appSupport, "Microsoft Edge"))
		paths = append(paths, filepath.Join(appSupport, "Chromium"))
		paths = append(paths, filepath.Join(appSupport, "BraveSoftware", "Brave-Browser"))
		paths = append(paths, filepath.Join(appSupport, "Vivaldi"))
		paths = append(paths, filepath.Join(appSupport, "Comet"))

	default:
		// Linux paths for Chromium-based browsers
		configDir := filepath.Join(user.HomeDir, ".config")
		paths = append(paths, filepath.Join(configDir, "google-chrome"))
		paths = append(paths, filepath.Join(configDir, "microsoft-edge"))
		paths = append(paths, filepath.Join(configDir, "chromium"))
		paths = append(paths, filepath.Join(configDir, "BraveSoftware", "Brave-Browser"))
		paths = append(paths, filepath.Join(configDir, "vivaldi"))
	}

	// Filter paths that exist
	var existingPaths []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}

	return existingPaths
}

// scanUsersWithWorkerPool scans users concurrently using a worker pool pattern
func scanUsersWithWorkerPool(users []common.UserInfo, scanFunc func(common.UserInfo) []string) []string {
	// Determine optimal number of workers based on system and user count
	maxWorkers := runtime.NumCPU()
	if len(users) < maxWorkers {
		maxWorkers = len(users)
	}
	if maxWorkers > 10 {
		maxWorkers = 10 // Cap at 10 to avoid excessive resource usage
	}

	// Create channels for work distribution
	userChan := make(chan common.UserInfo, len(users))
	resultChan := make(chan []string, len(users))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for user := range userChan {
				paths := scanFunc(user)
				resultChan <- paths
			}
		}()
	}

	// Send work to workers
	go func() {
		defer close(userChan)
		for _, user := range users {
			userChan <- user
		}
	}()

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allPaths []string
	for paths := range resultChan {
		allPaths = append(allPaths, paths...)
	}

	return allPaths
}
