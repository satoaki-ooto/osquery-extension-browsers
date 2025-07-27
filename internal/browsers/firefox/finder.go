package firefox

import (
	"os"
	"path/filepath"
	"runtime"
)

// FindFirefoxPaths returns the paths to Firefox browser data directories
func FindFirefoxPaths() []string {
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
	}

	return paths
}
