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

	switch runtime.GOOS {
	case "windows":
		variants = append(variants, BrowserVariant{
			Name:    "Chrome",
			Paths:   []string{filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data")},
			Process: "chrome.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Edge",
			Paths:   []string{filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data")},
			Process: "msedge.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Chromium",
			Paths:   []string{filepath.Join(os.Getenv("LOCALAPPDATA"), "Chromium", "User Data")},
			Process: "chromium.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Brave",
			Paths:   []string{filepath.Join(os.Getenv("LOCALAPPDATA"), "BraveSoftware", "Brave-Browser", "User Data")},
			Process: "brave.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Vivaldi",
			Paths:   []string{filepath.Join(os.Getenv("LOCALAPPDATA"), "Vivaldi", "User Data")},
			Process: "vivaldi.exe",
		})

	case "darwin":
		variants = append(variants, BrowserVariant{
			Name:    "Chrome",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Google", "Chrome")},
			Process: "Google Chrome",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Edge",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Microsoft Edge")},
			Process: "Microsoft Edge",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Chromium",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Chromium")},
			Process: "Chromium",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Brave",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "BraveSoftware", "Brave-Browser")},
			Process: "Brave Browser",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Vivaldi",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Vivaldi")},
			Process: "Vivaldi",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Comet",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Comet")},
			Process: "Comet",
		})

	default:
		variants = append(variants, BrowserVariant{
			Name:    "Chrome",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".config", "google-chrome")},
			Process: "google-chrome",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Edge",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".config", "microsoft-edge")},
			Process: "microsoft-edge",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Chromium",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".config", "chromium")},
			Process: "chromium",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Brave",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".config", "BraveSoftware", "Brave-Browser")},
			Process: "brave",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Vivaldi",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".config", "vivaldi")},
			Process: "vivaldi",
		})
	}

	return variants
}
