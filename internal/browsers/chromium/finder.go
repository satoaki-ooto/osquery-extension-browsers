package chromium

import (
	"os"
	"path/filepath"
	"runtime"
)

// FindChromiumPaths returns the paths to Chromium-based browser data directories
func FindChromiumPaths() []string {
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
