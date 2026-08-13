package firefox

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-ini/ini"

	"osquery-extension-browsers/internal/browsers/common"
)

// FindProfiles discovers all current Firefox-family profiles.
func FindProfiles() ([]common.Profile, error) {
	roots, err := FindRoots()
	if err != nil {
		return nil, fmt.Errorf("find Firefox roots: %w", err)
	}

	var profiles []common.Profile
	for _, root := range roots {
		rootProfiles, err := FindProfilesForRoot(root)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, rootProfiles...)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].ProfileID < profiles[j].ProfileID
	})
	return profiles, nil
}

// FindProfilesForRoot reads profiles.ini metadata or falls back to directory inventory.
func FindProfilesForRoot(root common.BrowserRoot) ([]common.Profile, error) {
	if _, err := os.Stat(root.Path); os.IsNotExist(err) {
		return []common.Profile{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("inspect %s root %q: %w", root.BrowserVariant, root.Path, err)
	}

	profilesINIPath, found, err := findProfilesINI(root.Path)
	if err != nil {
		return nil, fmt.Errorf(
			"inspect %s profiles.ini under %q: %w",
			root.BrowserVariant,
			root.Path,
			err,
		)
	}
	if found {
		return readProfilesINI(root, profilesINIPath)
	}
	return findProfilesInDirectory(root)
}

func findProfilesINI(rootPath string) (path string, found bool, err error) {
	candidates := []string{filepath.Join(rootPath, "profiles.ini")}
	if filepath.Base(rootPath) == "Profiles" {
		candidates = append(candidates, filepath.Join(filepath.Dir(rootPath), "profiles.ini"))
	}

	for _, candidate := range candidates {
		info, statErr := os.Stat(candidate)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return "", false, statErr
		}
		if !info.IsDir() {
			return candidate, true, nil
		}
	}
	return "", false, nil
}

func readProfilesINI(root common.BrowserRoot, profilesINIPath string) ([]common.Profile, error) {
	configuration, err := ini.Load(profilesINIPath)
	if err != nil {
		return nil, fmt.Errorf("read %s profiles.ini %q: %w", root.BrowserVariant, profilesINIPath, err)
	}

	profiles := make([]common.Profile, 0)
	for _, section := range configuration.Sections() {
		if !strings.HasPrefix(section.Name(), "Profile") || !section.HasKey("Path") {
			continue
		}

		profilePath := section.Key("Path").String()
		if section.Key("IsRelative").MustBool(true) {
			profilePath = filepath.Join(filepath.Dir(profilesINIPath), profilePath)
		}

		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("inspect %s profile %q: %w", root.BrowserVariant, profilePath, err)
		}

		displayName := section.Key("Name").String()
		profile, err := newProfile(root, profilePath, displayName, displayName != "")
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].ProfileID < profiles[j].ProfileID
	})
	return profiles, nil
}

func findProfilesInDirectory(root common.BrowserRoot) ([]common.Profile, error) {
	entries, err := os.ReadDir(root.Path)
	if err != nil {
		return nil, fmt.Errorf("read %s profile root %q: %w", root.BrowserVariant, root.Path, err)
	}

	profiles := make([]common.Profile, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		profile, err := newProfile(
			root,
			filepath.Join(root.Path, entry.Name()),
			"",
			false,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func newProfile(
	root common.BrowserRoot,
	profilePath string,
	displayName string,
	displayNamePresent bool,
) (common.Profile, error) {
	canonicalPath, err := common.CanonicalProfilePath(profilePath)
	if err != nil {
		return common.Profile{}, fmt.Errorf(
			"canonicalize %s profile %q: %w",
			root.BrowserVariant,
			profilePath,
			err,
		)
	}

	return common.Profile{
		ProfileID:          common.GenerateProfileID(root.BrowserVariant, canonicalPath),
		BrowserFamily:      firefoxFamily,
		BrowserVariant:     root.BrowserVariant,
		OSUserName:         root.OSUserName,
		Directory:          filepath.Base(filepath.Clean(profilePath)),
		Path:               canonicalPath,
		DisplayName:        displayName,
		DisplayNamePresent: displayNamePresent,
		AccountPresent:     false,
	}, nil
}
