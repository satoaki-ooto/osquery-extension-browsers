package common

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIdentityDeterminismAndBoundaries(t *testing.T) {
	canonicalPath, err := CanonicalProfilePath(filepath.Join(t.TempDir(), "Default"))
	if err != nil {
		t.Fatalf("CanonicalProfilePath() error = %v", err)
	}

	profileID := GenerateProfileID("chrome", canonicalPath)
	observationID := GenerateObservationID("chrome", canonicalPath, 42)
	if profileID != GenerateProfileID("chrome", canonicalPath) {
		t.Fatal("GenerateProfileID() was not deterministic")
	}
	if observationID != GenerateObservationID("chrome", canonicalPath, 42) {
		t.Fatal("GenerateObservationID() was not deterministic")
	}

	for name, identity := range map[string]string{
		"profile":     profileID,
		"observation": observationID,
	} {
		if len(identity) != 64 || identity != strings.ToLower(identity) {
			t.Errorf("%s identity = %q, want lowercase 64-character SHA-256", name, identity)
		}
	}

	if observationID == GenerateObservationID("chrome", canonicalPath, 43) {
		t.Error("changing native visit ID did not change observation identity")
	}
	if observationID == GenerateObservationID("edge", canonicalPath, 42) {
		t.Error("changing browser variant did not change observation identity")
	}
	if profileID == GenerateProfileID("chrome", canonicalPath+"-other") {
		t.Error("changing canonical profile path did not change profile identity")
	}
}

func TestCanonicalProfilePathDoesNotResolveSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation commonly requires elevated Windows privileges")
	}

	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}
	link := filepath.Join(root, "profile-link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("os.Symlink() error = %v", err)
	}

	canonicalPath, err := CanonicalProfilePath(link)
	if err != nil {
		t.Fatalf("CanonicalProfilePath() error = %v", err)
	}
	if !strings.HasSuffix(canonicalPath, "/profile-link") {
		t.Errorf("CanonicalProfilePath() = %q, symlink was unexpectedly resolved", canonicalPath)
	}
}

func TestIdentityComponentsAreBoundarySafe(t *testing.T) {
	first := hashIdentity("ab", "c")
	second := hashIdentity("a", "bc")
	if first == second {
		t.Error("NUL-delimited identity components produced a boundary collision")
	}
}
