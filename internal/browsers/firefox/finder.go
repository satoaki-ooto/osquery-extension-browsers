package firefox

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"osquery-extension-browsers/internal/browsers/common"
)

// FindFirefoxPaths returns the paths to Firefox browser data directories for all users
func FindFirefoxPaths() []string {
	users, err := common.UsersFromContext()
	if err != nil || len(users) == 0 {
		// Fallback to current user paths
		return findFirefoxPathsCurrentUser()
	}

	// Filter accessible users
	var accessibleUsers []common.UserInfo
	for _, user := range users {
		if user.IsAccessible {
			accessibleUsers = append(accessibleUsers, user)
		}
	}

	if len(accessibleUsers) == 0 {
		return findFirefoxPathsCurrentUser()
	}

	// Use worker pool for better performance and resource management
	allPaths := scanUsersWithWorkerPool(accessibleUsers, findFirefoxPathsForUser)

	// If no paths found from users, fallback to current user
	if len(allPaths) == 0 {
		allPaths = findFirefoxPathsCurrentUser()
	}

	return allPaths
}

// findFirefoxPathsForUser returns Firefox paths for a specific user
func findFirefoxPathsForUser(user common.UserInfo) []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Windows paths for Firefox
		appData := filepath.Join(user.HomeDir, "AppData", "Roaming")
		paths = append(paths, filepath.Join(appData, "Mozilla", "Firefox", "Profiles"))

	case "darwin":
		// macOS paths for Firefox
		paths = append(paths, filepath.Join(user.HomeDir, "Library", "Application Support", "Firefox", "Profiles"))

	default:
		// Linux paths for Firefox
		paths = append(paths, filepath.Join(user.HomeDir, ".mozilla", "firefox"))
		// Linux paths for Zen Browser
		paths = append(paths, filepath.Join(user.HomeDir, ".zen"))
		// Linux paths for Zen Browser (Flatpak)
		paths = append(paths, filepath.Join(user.HomeDir, ".var", "app", "app.zen_browser.zen", ".zen"))
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

// findFirefoxPathsCurrentUser returns Firefox paths for the current user only (fallback)
func findFirefoxPathsCurrentUser() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Windows paths for Firefox
		paths = append(paths, filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles"))

	case "darwin":
		// macOS paths for Firefox
		paths = append(paths, filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles"))

	default:
		// Linux paths for Firefox
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox"))
		// Linux paths for Zen Browser
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".zen"))
		// Linux paths for Zen Browser (Flatpak)
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".var", "app", "app.zen_browser.zen", ".zen"))
	}

	return paths
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
