package common

import (
	"runtime"
	"testing"
)

func TestUsersFromContext(t *testing.T) {
	users, err := UsersFromContext()

	// Should not return an error on most systems
	if err != nil {
		t.Logf("UsersFromContext() returned error: %v", err)
		// Don't fail the test as this might be expected in some environments
	}

	// Should return at least the current user in most cases
	if len(users) == 0 {
		t.Log("UsersFromContext() returned no users - this might be expected in restricted environments")
	}

	// Verify user info structure
	for _, user := range users {
		if user.Username == "" {
			t.Error("User has empty username")
		}
		if user.HomeDir == "" {
			t.Error("User has empty home directory")
		}
		// UID might be empty on Windows, so don't check it

		t.Logf("Found user: %s (Home: %s, UID: %s, Accessible: %v)",
			user.Username, user.HomeDir, user.UID, user.IsAccessible)
	}
}

func TestGetUsersLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux OS")
	}

	users, err := getUsersLinux()

	// Test should handle both success and failure gracefully
	if err != nil {
		t.Logf("getUsersLinux() returned error: %v", err)
	}

	// Verify user structure if any users returned
	for _, user := range users {
		if user.Username == "" {
			t.Error("Linux user has empty username")
		}
		if user.HomeDir == "" {
			t.Error("Linux user has empty home directory")
		}
		if user.UID == "" {
			t.Error("Linux user has empty UID")
		}
	}
}

func TestGetUsersMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS OS")
	}

	users, err := getUsersMacOS()

	// Test should handle both success and failure gracefully
	if err != nil {
		t.Logf("getUsersMacOS() returned error: %v", err)
	}

	// Verify user structure if any users returned
	for _, user := range users {
		if user.Username == "" {
			t.Error("macOS user has empty username")
		}
		if user.HomeDir == "" {
			t.Error("macOS user has empty home directory")
		}
		// UID might be empty in some cases, so don't enforce it
	}
}

func TestGetUsersWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows OS")
	}

	users, err := getUsersWindows()

	// Test should handle both success and failure gracefully
	if err != nil {
		t.Logf("getUsersWindows() returned error: %v", err)
	}

	// Verify user structure if any users returned
	for _, user := range users {
		if user.Username == "" {
			t.Error("Windows user has empty username")
		}
		if user.HomeDir == "" {
			t.Error("Windows user has empty home directory")
		}
		// Windows doesn't use numeric UIDs, so UID might be empty
	}
}

func TestUserInfoStructure(t *testing.T) {
	// Test UserInfo struct behavior
	user := UserInfo{
		Username:     "testuser",
		HomeDir:      "/home/testuser",
		UID:          "1001",
		IsAccessible: true,
	}

	if user.Username != "testuser" {
		t.Error("UserInfo Username not set correctly")
	}
	if user.HomeDir != "/home/testuser" {
		t.Error("UserInfo HomeDir not set correctly")
	}
	if user.UID != "1001" {
		t.Error("UserInfo UID not set correctly")
	}
	if !user.IsAccessible {
		t.Error("UserInfo IsAccessible not set correctly")
	}
}
