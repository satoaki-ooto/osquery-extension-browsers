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
		// Fallback to current user paths
		return findChromiumPathsCurrentUser()
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
			userPaths := findChromiumPathsForUser(u)

			mu.Lock()
			allPaths = append(allPaths, userPaths...)
			mu.Unlock()
		}(user)
	}

	wg.Wait()

	// If no paths found from users, fallback to current user
	if len(allPaths) == 0 {
		allPaths = findChromiumPathsCurrentUser()
	}

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

// findChromiumPathsCurrentUser returns Chromium paths for the current user only (fallback)
func findChromiumPathsCurrentUser() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Windows paths for Chromium-based browsers
		paths = append(paths, filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data"))
		paths = append(paths, filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data"))
		paths = append(paths, filepath.Join(os.Getenv("LOCALAPPDATA"), "Chromium", "User Data"))

	case "darwin":
		// macOS paths for Chromium-based browsers
		paths = append(paths, filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Google", "Chrome"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Microsoft Edge"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Chromium"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "BraveSoftware", "Brave-Browser"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Vivaldi"))

	default:
		// Linux paths for Chromium-based browsers
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".config", "google-chrome"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".config", "microsoft-edge"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".config", "chromium"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".config", "BraveSoftware", "Brave-Browser"))
		paths = append(paths, filepath.Join(os.Getenv("HOME"), ".config", "vivaldi"))
	}

	return paths
}
