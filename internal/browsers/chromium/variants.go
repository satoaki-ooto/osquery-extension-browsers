package chromium

import (
	"os"
	"path/filepath"
	"runtime"
)

// BrowserVariant represents a specific variant of a Chromium-based browser
type BrowserVariant struct {
	Name    string
	Paths   []string
	Process string
}

// DetectBrowserVariants returns a list of detected Chromium-based browser variants
func DetectBrowserVariants() []BrowserVariant {
	var variants []BrowserVariant
	homeDirectory := os.Getenv("HOME")
	localAppData := os.Getenv("LOCALAPPDATA")

	switch runtime.GOOS {
	case "windows":
		variants = append(variants, BrowserVariant{
			Name:    "chrome",
			Paths:   []string{filepath.Join(localAppData, "Google", "Chrome", "User Data")},
			Process: "chrome.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "edge",
			Paths:   []string{filepath.Join(localAppData, "Microsoft", "Edge", "User Data")},
			Process: "msedge.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "chromium",
			Paths:   []string{filepath.Join(localAppData, "Chromium", "User Data")},
			Process: "chromium.exe",
		})

		variants = append(variants, BrowserVariant{
			Name: "brave",
			Paths: []string{filepath.Join(
				localAppData,
				"BraveSoftware",
				"Brave-Browser",
				"User Data",
			)},
			Process: "brave.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "vivaldi",
			Paths:   []string{filepath.Join(localAppData, "Vivaldi", "User Data")},
			Process: "vivaldi.exe",
		})

	case "darwin":
		variants = append(variants, BrowserVariant{
			Name: "chrome",
			Paths: []string{filepath.Join(
				homeDirectory,
				"Library",
				"Application Support",
				"Google",
				"Chrome",
			)},
			Process: "Google Chrome",
		})

		variants = append(variants, BrowserVariant{
			Name: "edge",
			Paths: []string{filepath.Join(
				homeDirectory,
				"Library",
				"Application Support",
				"Microsoft Edge",
			)},
			Process: "Microsoft Edge",
		})

		variants = append(variants, BrowserVariant{
			Name:    "chromium",
			Paths:   []string{filepath.Join(homeDirectory, "Library", "Application Support", "Chromium")},
			Process: "Chromium",
		})

		variants = append(variants, BrowserVariant{
			Name: "brave",
			Paths: []string{filepath.Join(
				homeDirectory,
				"Library",
				"Application Support",
				"BraveSoftware",
				"Brave-Browser",
			)},
			Process: "Brave Browser",
		})

		variants = append(variants, BrowserVariant{
			Name:    "vivaldi",
			Paths:   []string{filepath.Join(homeDirectory, "Library", "Application Support", "Vivaldi")},
			Process: "Vivaldi",
		})

		variants = append(variants, BrowserVariant{
			Name:    "comet",
			Paths:   []string{filepath.Join(homeDirectory, "Library", "Application Support", "Comet")},
			Process: "Comet",
		})

	default:
		variants = append(variants, BrowserVariant{
			Name:    "chrome",
			Paths:   []string{filepath.Join(homeDirectory, ".config", "google-chrome")},
			Process: "google-chrome",
		})

		variants = append(variants, BrowserVariant{
			Name:    "edge",
			Paths:   []string{filepath.Join(homeDirectory, ".config", "microsoft-edge")},
			Process: "microsoft-edge",
		})

		variants = append(variants, BrowserVariant{
			Name:    "chromium",
			Paths:   []string{filepath.Join(homeDirectory, ".config", "chromium")},
			Process: "chromium",
		})

		variants = append(variants, BrowserVariant{
			Name:    "brave",
			Paths:   []string{filepath.Join(homeDirectory, ".config", "BraveSoftware", "Brave-Browser")},
			Process: "brave",
		})

		variants = append(variants, BrowserVariant{
			Name:    "vivaldi",
			Paths:   []string{filepath.Join(homeDirectory, ".config", "vivaldi")},
			Process: "vivaldi",
		})
	}

	return variants
}
