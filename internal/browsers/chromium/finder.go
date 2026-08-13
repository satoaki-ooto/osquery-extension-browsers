package chromium

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"osquery-extension-browsers/internal/browsers/common"
)

const chromiumFamily = "chromium"

type rootCandidate struct {
	variant string
	path    string
}

type rootScanResult struct {
	roots []common.BrowserRoot
	err   error
}

// FindRoots returns existing Chromium-family data roots with their owner and variant.
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

	roots, err := scanUsersWithWorkerPool(accessibleUsers, findChromiumRootsForUser)
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

// FindChromiumPaths preserves the path-only discovery API for existing callers.
func FindChromiumPaths() []string {
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

func findChromiumRootsForUser(user common.UserInfo) ([]common.BrowserRoot, error) {
	if !user.IsAccessible {
		return []common.BrowserRoot{}, nil
	}

	candidates := chromiumRootCandidates(user)
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
			BrowserFamily:  chromiumFamily,
			BrowserVariant: candidate.variant,
		})
	}

	return roots, nil
}

func findChromiumPathsForUser(user common.UserInfo) []string {
	roots, err := findChromiumRootsForUser(user)
	if err != nil {
		return []string{}
	}
	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		paths = append(paths, root.Path)
	}
	return paths
}

func chromiumRootCandidates(user common.UserInfo) []rootCandidate {
	switch runtime.GOOS {
	case "windows":
		localAppData := filepath.Join(user.HomeDir, "AppData", "Local")
		return []rootCandidate{
			{variant: "chrome", path: filepath.Join(localAppData, "Google", "Chrome", "User Data")},
			{variant: "edge", path: filepath.Join(localAppData, "Microsoft", "Edge", "User Data")},
			{variant: "chromium", path: filepath.Join(localAppData, "Chromium", "User Data")},
			{
				variant: "brave",
				path: filepath.Join(
					localAppData,
					"BraveSoftware",
					"Brave-Browser",
					"User Data",
				),
			},
			{variant: "vivaldi", path: filepath.Join(localAppData, "Vivaldi", "User Data")},
		}
	case "darwin":
		applicationSupport := filepath.Join(user.HomeDir, "Library", "Application Support")
		return []rootCandidate{
			{variant: "chrome", path: filepath.Join(applicationSupport, "Google", "Chrome")},
			{variant: "edge", path: filepath.Join(applicationSupport, "Microsoft Edge")},
			{variant: "chromium", path: filepath.Join(applicationSupport, "Chromium")},
			{variant: "brave", path: filepath.Join(applicationSupport, "BraveSoftware", "Brave-Browser")},
			{variant: "vivaldi", path: filepath.Join(applicationSupport, "Vivaldi")},
			{variant: "comet", path: filepath.Join(applicationSupport, "Comet")},
		}
	default:
		configDirectory := filepath.Join(user.HomeDir, ".config")
		return []rootCandidate{
			{variant: "chrome", path: filepath.Join(configDirectory, "google-chrome")},
			{variant: "edge", path: filepath.Join(configDirectory, "microsoft-edge")},
			{variant: "chromium", path: filepath.Join(configDirectory, "chromium")},
			{variant: "brave", path: filepath.Join(configDirectory, "BraveSoftware", "Brave-Browser")},
			{variant: "vivaldi", path: filepath.Join(configDirectory, "vivaldi")},
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
