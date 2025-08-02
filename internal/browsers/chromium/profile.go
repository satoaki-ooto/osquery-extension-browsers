package chromium

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"osquery-extension-browsers/internal/browsers/common"
)

// ProfileInfo represents the structure of the Preferences file
type ProfileInfo struct {
	Name string `json:"name"`
}

// findProfileDirectories returns a list of profile directories within the user data directory
func findProfileDirectories(userDataDir string) ([]string, error) {
	var profileDirs []string

	// Check if the user data directory exists
	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		return profileDirs, err
	}

	// Read the contents of the user data directory
	entries, err := ioutil.ReadDir(userDataDir)
	if err != nil {
		return profileDirs, err
	}

	// Regular expression to match profile directories
	profileDirRegex := regexp.MustCompile(`^(Default|Profile \d+)$`)

	for _, entry := range entries {
		if entry.IsDir() && profileDirRegex.MatchString(entry.Name()) {
			profileDirs = append(profileDirs, filepath.Join(userDataDir, entry.Name()))
		}
	}

	return profileDirs, nil
}

// readProfileInfo reads profile information from the Preferences file
func readProfileInfo(profileDir string) (common.Profile, error) {
	profile := common.Profile{
		Path: profileDir,
	}

	// Set the profile ID based on the directory name
	dirName := filepath.Base(profileDir)
	profile.ID = dirName

	// Set the profile name based on the directory name
	if dirName == "Default" {
		profile.Name = "Default"
	} else {
		profile.Name = dirName
	}

	// Read the Preferences file
	preferencesPath := filepath.Join(profileDir, "Preferences")
	if _, err := os.Stat(preferencesPath); os.IsNotExist(err) {
		// If Preferences file doesn't exist, return the basic profile info
		return profile, nil
	}

	data, err := ioutil.ReadFile(preferencesPath)
	if err != nil {
		return profile, err
	}

	var profileInfo ProfileInfo
	if err := json.Unmarshal(data, &profileInfo); err != nil {
		return profile, err
	}

	// Update the profile name if it exists in the Preferences file
	if profileInfo.Name != "" {
		profile.Name = profileInfo.Name
	}

	return profile, nil
}

// FindProfiles discovers all profiles for Chromium-based browsers
func FindProfiles() ([]common.Profile, error) {
	var profiles []common.Profile

	// Get the paths to Chromium-based browser data directories
	chromiumPaths := FindChromiumPaths()

	for _, userDataDir := range chromiumPaths {
		// Find profile directories within each user data directory
		profileDirs, err := findProfileDirectories(userDataDir)
		if err != nil {
			// If we can't read a directory, continue with the next one
			continue
		}

		// Read profile information for each profile directory
		for _, profileDir := range profileDirs {
			profile, err := readProfileInfo(profileDir)
			if err != nil {
				// If we can't read a profile, continue with the next one
				continue
			}

			// Set browser type and variant
			profile.BrowserVariant = getBrowserVariant(userDataDir)
			profile.BrowserType = strings.ToLower(profile.BrowserVariant)

			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// getBrowserVariant determines the browser variant based on the user data directory path
func getBrowserVariant(userDataDir string) string {
	switch {
	case strings.Contains(strings.ToLower(userDataDir), "chrome"):
		return "chrome"
	case strings.Contains(strings.ToLower(userDataDir), "edge"):
		return "edge"
	case strings.Contains(strings.ToLower(userDataDir), "chromium"):
		return "chromium"
	case strings.Contains(strings.ToLower(userDataDir), "brave"):
		return "brave"
	case strings.Contains(strings.ToLower(userDataDir), "vivaldi"):
		return "vivaldi"
	default:
		return "chromium"
	}
}
