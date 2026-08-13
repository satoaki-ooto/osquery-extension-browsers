package firefox

import (
	"os"
	"path/filepath"
	"testing"

	"osquery-extension-browsers/internal/browsers/common"
)

func TestFindProfilesForRootKeepsFirefoxMetadataDistinct(t *testing.T) {
	productRoot := t.TempDir()
	profilesRoot := filepath.Join(productRoot, "Profiles")
	profilePath := filepath.Join(profilesRoot, "abc.default-release")
	if err := os.MkdirAll(profilePath, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	profilesINI := "[Profile0]\nName=Visible Name\nIsRelative=1\nPath=Profiles/abc.default-release\n"
	profilesINIPath := filepath.Join(productRoot, "profiles.ini")
	if err := os.WriteFile(profilesINIPath, []byte(profilesINI), 0o600); err != nil {
		t.Fatalf("os.WriteFile(profiles.ini) error = %v", err)
	}

	profiles, err := FindProfilesForRoot(common.BrowserRoot{
		Path:           profilesRoot,
		OSUserName:     "os-user",
		BrowserFamily:  "firefox",
		BrowserVariant: "firefox",
	})
	if err != nil {
		t.Fatalf("FindProfilesForRoot() error = %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("FindProfilesForRoot() returned %d profiles, want 1", len(profiles))
	}

	profile := profiles[0]
	if profile.Directory != "abc.default-release" || profile.DisplayName != "Visible Name" {
		t.Errorf("directory/display name were conflated: %#v", profile)
	}
	if !profile.DisplayNamePresent || profile.AccountPresent || profile.Account != "" {
		t.Errorf("unexpected Firefox presence metadata: %#v", profile)
	}
	if profile.OSUserName != "os-user" || profile.BrowserVariant != "firefox" {
		t.Errorf("unexpected Firefox owner/variant: %#v", profile)
	}
}
