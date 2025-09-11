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

	var allPaths []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use goroutines to scan users concurrently
	for _, user := range users {
		if !user.IsAccessible {
			continue
		}

		wg.Add(1)
		go func(u common.UserInfo) {
			defer wg.Done()
			userPaths := findFirefoxPathsForUser(u)

			mu.Lock()
			allPaths = append(allPaths, userPaths...)
			mu.Unlock()
		}(user)
	}

	wg.Wait()

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
