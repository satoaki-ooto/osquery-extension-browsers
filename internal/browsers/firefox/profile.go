package firefox

import (
	"errors"
	"os"
	"path/filepath"

	"osquery-extension-browsers/internal/browsers/common"

	"github.com/go-ini/ini"
)

// FindProfiles discovers all profiles for Firefox browsers
func FindProfiles() ([]common.Profile, error) {
	var profiles []common.Profile

	// Get the paths to Firefox browser data directories
	firefoxPaths := FindFirefoxPaths()

	for _, profilesDir := range firefoxPaths {
		// Check if the profiles directory exists
		if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
			continue
		}

		// Read the profiles.ini file
		profilesIniPath := filepath.Join(filepath.Dir(profilesDir), "profiles.ini")
		if _, err := os.Stat(profilesIniPath); os.IsNotExist(err) {
			// If profiles.ini doesn't exist, try to find profiles in the directory
			profile, err := findProfilesInDirectory(profilesDir)
			if err == nil {
				profiles = append(profiles, profile)
			}
			continue
		}

		// Parse the profiles.ini file
		profilesFromIni, err := readProfilesIni(profilesIniPath, profilesDir)
		if err != nil {
			continue
		}

		profiles = append(profiles, profilesFromIni...)
	}

	return profiles, nil
}

// readProfilesIni reads profile information from the profiles.ini file
func readProfilesIni(profilesIniPath, profilesDir string) ([]common.Profile, error) {
	var profiles []common.Profile

	// Load the INI file
	cfg, err := ini.Load(profilesIniPath)
	if err != nil {
		return profiles, err
	}

	// Iterate through sections
	for _, section := range cfg.Sections() {
		// Skip the default section
		if section.Name() == "DEFAULT" {
			continue
		}

		// Check if this is a profile section
		if section.HasKey("Path") {
			profile, err := parseProfileSection(section, profilesDir)
			if err != nil {
				continue
			}

			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// parseProfileSection parses a profile section from the profiles.ini file
func parseProfileSection(section *ini.Section, profilesDir string) (common.Profile, error) {
	profile := common.Profile{
		BrowserType:    "Firefox",
		BrowserVariant: "Firefox",
	}

	// Get the profile name
	if nameKey := section.Key("Name"); nameKey != nil {
		profile.Name = nameKey.String()
	}

	// Get the profile path
	pathKey := section.Key("Path")
	if pathKey == nil {
		return profile, errors.New("profile path not found")
	}

	profilePath := pathKey.String()

	// Check if the path is relative
	isRelative := true
	if isRelativeKey := section.Key("IsRelative"); isRelativeKey != nil {
		isRelative = isRelativeKey.MustBool(true)
	}

	if isRelative {
		profile.Path = filepath.Join(profilesDir, profilePath)
	} else {
		profile.Path = profilePath
	}

	// Set the profile ID based on the directory name
	profile.ID = filepath.Base(profile.Path)

	return profile, nil
}

// findProfilesInDirectory finds profiles in a directory when profiles.ini is not available
func findProfilesInDirectory(profilesDir string) (common.Profile, error) {
	profile := common.Profile{
		ID:             "default",
		Name:           "Default",
		Path:           profilesDir,
		BrowserType:    "Firefox",
		BrowserVariant: "Firefox",
	}

	return profile, nil
}
