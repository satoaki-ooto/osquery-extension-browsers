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

	switch runtime.GOOS {
	case "windows":
		variants = append(variants, BrowserVariant{
			Name:    "Firefox",
			Paths:   []string{filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles")},
			Process: "firefox.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Firefox Developer Edition",
			Paths:   []string{filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles")},
			Process: "firefox.exe",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Firefox Nightly",
			Paths:   []string{filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles")},
			Process: "firefox.exe",
		})

	case "darwin":
		variants = append(variants, BrowserVariant{
			Name:    "Firefox",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles")},
			Process: "Firefox",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Firefox Developer Edition",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles")},
			Process: "Firefox Developer Edition",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Firefox Nightly",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles")},
			Process: "Firefox Nightly",
		})

	default:
		variants = append(variants, BrowserVariant{
			Name:    "Firefox",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox")},
			Process: "firefox",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Firefox Developer Edition",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox")},
			Process: "firefox",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Firefox Nightly",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox")},
			Process: "firefox",
		})

		variants = append(variants, BrowserVariant{
			Name:    "Zen Browser",
			Paths:   []string{filepath.Join(os.Getenv("HOME"), ".zen"), filepath.Join(os.Getenv("HOME"), ".var", "app", "app.zen_browser.zen", ".zen")},
			Process: "zen",
		})
	}

	return variants
}
