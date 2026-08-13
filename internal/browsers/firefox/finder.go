package firefox

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"osquery-extension-browsers/internal/browsers/common"
)

const firefoxFamily = "firefox"

type rootCandidate struct {
	variant string
	path    string
}

type rootScanResult struct {
	roots []common.BrowserRoot
	err   error
}

// FindRoots returns existing Firefox-family profile roots with their owner and variant.
func FindRoots() ([]common.BrowserRoot, error) {
	users, err := common.UsersFromContext()
	if err != nil {
		return nil, err
	}

	accessibleUsers := make([]common.UserInfo, 0, len(users))
	for _, user := range users {
		if user.IsAccessible {
			accessibleUsers = append(accessibleUsers, user)
		}
	}

	roots, err := scanUsersWithWorkerPool(accessibleUsers, findFirefoxRootsForUser)
	if err != nil {
		return nil, err
	}
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].Path == roots[j].Path {
			return roots[i].OSUserName < roots[j].OSUserName
		}
		return roots[i].Path < roots[j].Path
	})
	return roots, nil
}

// FindFirefoxPaths preserves the path-only discovery API for existing callers.
func FindFirefoxPaths() []string {
	roots, err := FindRoots()
	if err != nil {
		return []string{}
	}

	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		paths = append(paths, root.Path)
	}
	return paths
}

func findFirefoxRootsForUser(user common.UserInfo) ([]common.BrowserRoot, error) {
	if !user.IsAccessible {
		return []common.BrowserRoot{}, nil
	}

	candidates := firefoxRootCandidates(user)
	roots := make([]common.BrowserRoot, 0, len(candidates))
	for _, candidate := range candidates {
		info, err := os.Stat(candidate.path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf(
				"inspect %s root %q for OS user %q: %w",
				candidate.variant,
				candidate.path,
				user.Username,
				err,
			)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf(
				"inspect %s root %q for OS user %q: path is not a directory",
				candidate.variant,
				candidate.path,
				user.Username,
			)
		}
		roots = append(roots, common.BrowserRoot{
			Path:           candidate.path,
			OSUserName:     user.Username,
			BrowserFamily:  firefoxFamily,
			BrowserVariant: candidate.variant,
		})
	}
	return roots, nil
}

func findFirefoxPathsForUser(user common.UserInfo) []string {
	roots, err := findFirefoxRootsForUser(user)
	if err != nil {
		return []string{}
	}
	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		paths = append(paths, root.Path)
	}
	return paths
}

func firefoxRootCandidates(user common.UserInfo) []rootCandidate {
	switch runtime.GOOS {
	case "windows":
		roaming := filepath.Join(user.HomeDir, "AppData", "Roaming")
		return []rootCandidate{
			{variant: "firefox", path: filepath.Join(roaming, "Mozilla", "Firefox", "Profiles")},
			{variant: "zen", path: filepath.Join(roaming, "zen", "Profiles")},
			{variant: "floorp", path: filepath.Join(roaming, "Floorp", "Profiles")},
		}
	case "darwin":
		applicationSupport := filepath.Join(user.HomeDir, "Library", "Application Support")
		return []rootCandidate{
			{variant: "firefox", path: filepath.Join(applicationSupport, "Firefox", "Profiles")},
			{variant: "zen", path: filepath.Join(applicationSupport, "zen", "Profiles")},
			{variant: "floorp", path: filepath.Join(applicationSupport, "Floorp", "Profiles")},
		}
	default:
		return []rootCandidate{
			{variant: "firefox", path: filepath.Join(user.HomeDir, ".mozilla", "firefox")},
			{variant: "zen", path: filepath.Join(user.HomeDir, ".zen")},
			{
				variant: "zen",
				path: filepath.Join(
					user.HomeDir,
					".var",
					"app",
					"app.zen_browser.zen",
					".zen",
				),
			},
			{variant: "floorp", path: filepath.Join(user.HomeDir, ".floorp")},
			{
				variant: "floorp",
				path: filepath.Join(
					user.HomeDir,
					".var",
					"app",
					"one.ablaze.floorp",
					".floorp",
				),
			},
		}
	}
}

func scanUsersWithWorkerPool(
	users []common.UserInfo,
	scan func(common.UserInfo) ([]common.BrowserRoot, error),
) ([]common.BrowserRoot, error) {
	if len(users) == 0 {
		return []common.BrowserRoot{}, nil
	}

	workerCount := min(runtime.NumCPU(), len(users), 10)
	userChannel := make(chan common.UserInfo, len(users))
	resultChannel := make(chan rootScanResult, len(users))

	var waitGroup sync.WaitGroup
	for range workerCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for user := range userChannel {
				roots, err := scan(user)
				resultChannel <- rootScanResult{roots: roots, err: err}
			}
		}()
	}

	go func() {
		defer close(userChannel)
		for _, user := range users {
			userChannel <- user
		}
	}()

	go func() {
		waitGroup.Wait()
		close(resultChannel)
	}()

	var roots []common.BrowserRoot
	var firstError error
	for result := range resultChannel {
		if result.err != nil && firstError == nil {
			firstError = result.err
		}
		roots = append(roots, result.roots...)
	}
	if firstError != nil {
		return nil, firstError
	}
	return roots, nil
}
