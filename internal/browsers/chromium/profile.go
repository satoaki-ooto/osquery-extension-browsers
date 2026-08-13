package chromium

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"osquery-extension-browsers/internal/browsers/common"
)

var profileDirectoryPattern = regexp.MustCompile(`^(Default|Profile \d+)$`)

type localState struct {
	Profile struct {
		InfoCache map[string]localStateProfile `json:"info_cache"`
	} `json:"profile"`
}

type localStateProfile struct {
	Name     string `json:"name"`
	UserName string `json:"user_name"`
}

// FindProfiles discovers all current Chromium-family profiles.
func FindProfiles() ([]common.Profile, error) {
	roots, err := FindRoots()
	if err != nil {
		return nil, fmt.Errorf("find Chromium roots: %w", err)
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

// FindProfilesForRoot reads profile inventory and Local State once for one browser root.
func FindProfilesForRoot(root common.BrowserRoot) ([]common.Profile, error) {
	entries, err := os.ReadDir(root.Path)
	if os.IsNotExist(err) {
		return []common.Profile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s root %q: %w", root.BrowserVariant, root.Path, err)
	}

	metadata, err := readLocalState(root)
	if err != nil {
		return nil, err
	}

	profiles := make([]common.Profile, 0)
	for _, entry := range entries {
		if !entry.IsDir() || !profileDirectoryPattern.MatchString(entry.Name()) {
			continue
		}

		canonicalPath, err := common.CanonicalProfilePath(filepath.Join(root.Path, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf(
				"canonicalize %s profile %q: %w",
				root.BrowserVariant,
				entry.Name(),
				err,
			)
		}

		profileMetadata, metadataPresent := metadata[entry.Name()]
		profile := common.Profile{
			ProfileID:      common.GenerateProfileID(root.BrowserVariant, canonicalPath),
			BrowserFamily:  chromiumFamily,
			BrowserVariant: root.BrowserVariant,
			OSUserName:     root.OSUserName,
			Directory:      entry.Name(),
			Path:           canonicalPath,
		}
		if metadataPresent && profileMetadata.Name != "" {
			profile.DisplayName = profileMetadata.Name
			profile.DisplayNamePresent = true
		}
		if metadataPresent && profileMetadata.UserName != "" {
			profile.Account = profileMetadata.UserName
			profile.AccountPresent = true
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func readLocalState(root common.BrowserRoot) (map[string]localStateProfile, error) {
	localStatePath := filepath.Join(root.Path, "Local State")
	data, err := os.ReadFile(localStatePath)
	if os.IsNotExist(err) {
		return map[string]localStateProfile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s Local State %q: %w", root.BrowserVariant, localStatePath, err)
	}

	var state localState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse %s Local State %q: %w", root.BrowserVariant, localStatePath, err)
	}
	if state.Profile.InfoCache == nil {
		return map[string]localStateProfile{}, nil
	}
	return state.Profile.InfoCache, nil
}
