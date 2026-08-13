package firefox

import (
	"os"
	"path/filepath"
	"runtime"
)

// BrowserVariant represents a specific variant of a Firefox-based browser
type BrowserVariant struct {
	Name    string
	Paths   []string
	Process string
}

// DetectBrowserVariants returns a list of detected Firefox-based browser variants
func DetectBrowserVariants() []BrowserVariant {
	var variants []BrowserVariant
	appData := os.Getenv("APPDATA")
	homeDirectory := os.Getenv("HOME")

	switch runtime.GOOS {
	case "windows":
		variants = append(variants, BrowserVariant{
			Name:    "firefox",
			Paths:   []string{filepath.Join(appData, "Mozilla", "Firefox", "Profiles")},
			Process: "firefox.exe",
		})

	case "darwin":
		variants = append(variants, BrowserVariant{
			Name: "firefox",
			Paths: []string{filepath.Join(
				homeDirectory,
				"Library",
				"Application Support",
				"Firefox",
				"Profiles",
			)},
			Process: "Firefox",
		})

		variants = append(variants, BrowserVariant{
			Name: "zen",
			Paths: []string{filepath.Join(
				homeDirectory,
				"Library",
				"Application Support",
				"zen",
				"Profiles",
			)},
			Process: "zen",
		})

		variants = append(variants, BrowserVariant{
			Name: "floorp",
			Paths: []string{filepath.Join(
				homeDirectory,
				"Library",
				"Application Support",
				"Floorp",
				"Profiles",
			)},
			Process: "floorp",
		})

	default:
		variants = append(variants, BrowserVariant{
			Name:    "firefox",
			Paths:   []string{filepath.Join(homeDirectory, ".mozilla", "firefox")},
			Process: "firefox",
		})

		variants = append(variants, BrowserVariant{
			Name: "zen",
			Paths: []string{
				filepath.Join(homeDirectory, ".zen"),
				filepath.Join(
					homeDirectory,
					".var",
					"app",
					"app.zen_browser.zen",
					".zen",
				),
			},
			Process: "zen",
		})
	}

	return variants
}
