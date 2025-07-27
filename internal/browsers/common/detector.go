package common

import (
	"os"
	"path/filepath"
	"runtime"
)

// DetectInstalledBrowsers returns a list of installed browsers
func DetectInstalledBrowsers() []string {
	var browsers []string

	// Check for Chrome
	if isBrowserInstalled(getChromePaths()) {
		browsers = append(browsers, "Chrome")
	}

	// Check for Edge
	if isBrowserInstalled(getEdgePaths()) {
		browsers = append(browsers, "Edge")
	}

	// Check for Chromium
	if isBrowserInstalled(getChromiumPaths()) {
		browsers = append(browsers, "Chromium")
	}

	// Check for Firefox
	if isBrowserInstalled(getFirefoxPaths()) {
		browsers = append(browsers, "Firefox")
	}

	return browsers
}

// isBrowserInstalled checks if a browser is installed by checking if any of its paths exist
func isBrowserInstalled(paths []string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// getChromePaths returns the paths where Chrome might be installed
func getChromePaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES"), "Google", "Chrome", "Application", "chrome.exe"))
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Google", "Chrome", "Application", "chrome.exe"))
	case "darwin":
		paths = append(paths, "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	default:
		paths = append(paths, "/usr/bin/google-chrome")
		paths = append(paths, "/usr/bin/google-chrome-stable")
		paths = append(paths, "/usr/bin/chromium-browser")
	}

	return paths
}

// getEdgePaths returns the paths where Edge might be installed
func getEdgePaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES"), "Microsoft", "Edge", "Application", "msedge.exe"))
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Microsoft", "Edge", "Application", "msedge.exe"))
	case "darwin":
		paths = append(paths, "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge")
	default:
		paths = append(paths, "/usr/bin/microsoft-edge")
		paths = append(paths, "/usr/bin/microsoft-edge-stable")
	}

	return paths
}

// getChromiumPaths returns the paths where Chromium might be installed
func getChromiumPaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES"), "Chromium", "Application", "chrome.exe"))
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Chromium", "Application", "chrome.exe"))
	case "darwin":
		paths = append(paths, "/Applications/Chromium.app/Contents/MacOS/Chromium")
	default:
		paths = append(paths, "/usr/bin/chromium")
		paths = append(paths, "/usr/bin/chromium-browser")
	}

	return paths
}

// getFirefoxPaths returns the paths where Firefox might be installed
func getFirefoxPaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES"), "Mozilla Firefox", "firefox.exe"))
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Mozilla Firefox", "firefox.exe"))
	case "darwin":
		paths = append(paths, "/Applications/Firefox.app/Contents/MacOS/firefox")
	default:
		paths = append(paths, "/usr/bin/firefox")
	}

	return paths
}
