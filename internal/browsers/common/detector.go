package common

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// UserInfo represents information about a system user
type UserInfo struct {
	Username     string
	HomeDir      string
	UID          string
	IsAccessible bool
}

// usersFromContext returns a list of all system users
func UsersFromContext() ([]UserInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return getUsersWindows()
	case "darwin":
		return getUsersMacOS()
	default:
		return getUsersLinux()
	}
}

// getUsersLinux enumerates users on Linux systems
func getUsersLinux() ([]UserInfo, error) {
	var users []UserInfo

	file, err := os.Open("/etc/passwd")
	if err != nil {
		log.Printf("Warning: Failed to open /etc/passwd for user enumeration: %v", err)
		return users, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")

		if len(fields) < 6 {
			continue
		}

		username := fields[0]
		uid := fields[2]
		homeDir := fields[5]

		// Skip system users (UID < 1000) and users without valid home directories
		if uidInt, err := strconv.Atoi(uid); err == nil && uidInt >= 1000 {
			if strings.HasPrefix(homeDir, "/home/") || strings.HasPrefix(homeDir, "/Users/") {
				user := UserInfo{
					Username: username,
					HomeDir:  homeDir,
					UID:      uid,
				}

				// Check if home directory is accessible
				if _, err := os.Stat(homeDir); err == nil {
					user.IsAccessible = true
				} else {
					log.Printf("Debug: User %s home directory not accessible: %v", username, err)
				}

				users = append(users, user)
			}
		}
	}

	return users, scanner.Err()
}

// getUsersMacOS enumerates users on macOS systems
func getUsersMacOS() ([]UserInfo, error) {
	var users []UserInfo

	// Use dscl to get user list
	cmd := exec.Command("dscl", ".", "list", "/Users")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to enumerate macOS users with dscl: %v", err)
		return users, err
	}

	lines := strings.Split(string(output), "\n")
	for _, username := range lines {
		username = strings.TrimSpace(username)
		if username == "" || strings.HasPrefix(username, "_") || username == "daemon" || username == "nobody" {
			continue
		}

		// Get user info
		cmd := exec.Command("dscl", ".", "read", "/Users/"+username)
		userOutput, err := cmd.Output()
		if err != nil {
			log.Printf("Debug: Failed to get user info for %s: %v", username, err)
			continue
		}

		userInfo := string(userOutput)
		var homeDir, uid string

		// Parse dscl output
		lines := strings.Split(userInfo, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NFSHomeDirectory:") {
				homeDir = strings.TrimSpace(strings.TrimPrefix(line, "NFSHomeDirectory:"))
			} else if strings.HasPrefix(line, "UniqueID:") {
				uid = strings.TrimSpace(strings.TrimPrefix(line, "UniqueID:"))
			}
		}

		if homeDir != "" && strings.HasPrefix(homeDir, "/Users/") {
			user := UserInfo{
				Username: username,
				HomeDir:  homeDir,
				UID:      uid,
			}

			// Check if home directory is accessible
			if _, err := os.Stat(homeDir); err == nil {
				user.IsAccessible = true
			} else {
				log.Printf("Debug: User %s home directory not accessible: %v", username, err)
			}

			users = append(users, user)
		}
	}

	return users, nil
}

// getUsersWindows enumerates users on Windows systems
func getUsersWindows() ([]UserInfo, error) {
	var users []UserInfo

	// Get list of user directories from C:\Users
	usersDir := filepath.Join("C:", "Users")
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		log.Printf("Warning: Failed to read Windows users directory: %v", err)
		return users, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		username := entry.Name()
		// Skip system directories
		if username == "Public" || username == "Default" || username == "All Users" {
			continue
		}

		homeDir := filepath.Join(usersDir, username)
		user := UserInfo{
			Username: username,
			HomeDir:  homeDir,
			UID:      "", // Windows doesn't use numeric UIDs in the same way
		}

		// Check if home directory is accessible
		if _, err := os.Stat(homeDir); err == nil {
			user.IsAccessible = true
		} else {
			log.Printf("Debug: User %s home directory not accessible: %v", username, err)
		}

		users = append(users, user)
	}

	return users, nil
}
