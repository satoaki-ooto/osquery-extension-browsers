package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	profileIdentityVersion     = "browser-profile-v1"
	observationIdentityVersion = "browser-visit-observation-v1"
)

// CanonicalProfilePath returns the inspectable path used as browser profile identity input.
func CanonicalProfilePath(profilePath string) (string, error) {
	if profilePath == "" {
		return "", fmt.Errorf("canonicalize profile path: path is empty")
	}

	absolutePath, err := filepath.Abs(profilePath)
	if err != nil {
		return "", fmt.Errorf("canonicalize profile path %q: %w", profilePath, err)
	}

	canonicalPath := filepath.ToSlash(filepath.Clean(absolutePath))
	if runtime.GOOS == "windows" {
		canonicalPath = strings.ToLower(canonicalPath)
	}

	return canonicalPath, nil
}

// GenerateProfileID returns a deterministic identifier for a canonical browser profile.
func GenerateProfileID(browserVariant, canonicalProfilePath string) string {
	return hashIdentity(profileIdentityVersion, browserVariant, canonicalProfilePath)
}

// GenerateObservationID returns a deterministic identifier for a native browser visit.
func GenerateObservationID(
	browserVariant string,
	canonicalProfilePath string,
	nativeVisitID int64,
) string {
	return hashIdentity(
		observationIdentityVersion,
		browserVariant,
		canonicalProfilePath,
		strconv.FormatInt(nativeVisitID, 10),
	)
}

func hashIdentity(components ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(components, "\x00")))
	return hex.EncodeToString(digest[:])
}
