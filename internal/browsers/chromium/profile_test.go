package chromium

import (
	"os"
	"path/filepath"
	"testing"

	"osquery-extension-browsers/internal/browsers/common"
)

func TestFindProfilesForRootReadsLocalStateOnce(t *testing.T) {
	rootPath := t.TempDir()
	for _, directory := range []string{"Default", "Profile 4"} {
		if err := os.Mkdir(filepath.Join(rootPath, directory), 0o755); err != nil {
			t.Fatalf("os.Mkdir(%q) error = %v", directory, err)
		}
	}
	localState := `{
		"profile": {"info_cache": {
			"Default": {"name": "Work", "user_name": "person@example.com"},
			"Profile 4": {"name": "Empty account", "user_name": ""}
		}}
	}`
	localStatePath := filepath.Join(rootPath, "Local State")
	if err := os.WriteFile(localStatePath, []byte(localState), 0o600); err != nil {
		t.Fatalf("os.WriteFile(Local State) error = %v", err)
	}

	profiles, err := FindProfilesForRoot(common.BrowserRoot{
		Path:           rootPath,
		OSUserName:     "os-user",
		BrowserFamily:  "chromium",
		BrowserVariant: "chrome",
	})
	if err != nil {
		t.Fatalf("FindProfilesForRoot() error = %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("FindProfilesForRoot() returned %d profiles, want 2", len(profiles))
	}

	byDirectory := map[string]common.Profile{}
	for _, profile := range profiles {
		byDirectory[profile.Directory] = profile
		if !filepath.IsAbs(profile.Path) {
			t.Errorf("profile path %q is not absolute", profile.Path)
		}
		if profile.OSUserName != "os-user" || profile.BrowserVariant != "chrome" {
			t.Errorf("profile owner/variant = %q/%q", profile.OSUserName, profile.BrowserVariant)
		}
	}

	defaultProfile := byDirectory["Default"]
	if !defaultProfile.DisplayNamePresent || defaultProfile.DisplayName != "Work" {
		t.Errorf(
			"Default display name = %q present=%v",
			defaultProfile.DisplayName,
			defaultProfile.DisplayNamePresent,
		)
	}
	if !defaultProfile.AccountPresent || defaultProfile.Account != "person@example.com" {
		t.Errorf("Default account = %q present=%v", defaultProfile.Account, defaultProfile.AccountPresent)
	}
	profileFour := byDirectory["Profile 4"]
	if profileFour.AccountPresent {
		t.Errorf("empty Local State user_name was marked present: %#v", profileFour)
	}
}
