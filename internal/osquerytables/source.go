package osquerytables

import (
	"context"
	"fmt"
	"sort"

	"osquery-extension-browsers/internal/browsers/chromium"
	"osquery-extension-browsers/internal/browsers/common"
	"osquery-extension-browsers/internal/browsers/firefox"
)

type dataSource interface {
	Profiles(context.Context) ([]common.Profile, error)
	VisitObservations(context.Context, common.Profile, []int64) ([]common.VisitObservation, error)
	HistoryPages(context.Context, common.Profile, []int64) ([]common.HistoryPage, error)
}

type browserDataSource struct{}

func (browserDataSource) Profiles(ctx context.Context) ([]common.Profile, error) {
	chromiumProfiles, err := chromium.FindProfiles()
	if err != nil {
		return nil, fmt.Errorf("discover Chromium profiles: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	firefoxProfiles, err := firefox.FindProfiles()
	if err != nil {
		return nil, fmt.Errorf("discover Firefox profiles: %w", err)
	}

	allProfiles := append(chromiumProfiles, firefoxProfiles...)
	profilesByID := make(map[string]common.Profile, len(allProfiles))
	for _, profile := range allProfiles {
		existingProfile, present := profilesByID[profile.ProfileID]
		if present && existingProfile != profile {
			return nil, fmt.Errorf(
				"profile ID %q resolves to conflicting profile metadata",
				profile.ProfileID,
			)
		}
		profilesByID[profile.ProfileID] = profile
	}

	profiles := make([]common.Profile, 0, len(profilesByID))
	for _, profile := range profilesByID {
		profiles = append(profiles, profile)
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].ProfileID < profiles[j].ProfileID
	})
	return profiles, nil
}

func (browserDataSource) VisitObservations(
	ctx context.Context,
	profile common.Profile,
	nativeURLIDs []int64,
) ([]common.VisitObservation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	switch profile.BrowserFamily {
	case "chromium":
		return chromium.FindVisitObservations(profile, nativeURLIDs)
	case "firefox":
		return firefox.FindVisitObservations(profile, nativeURLIDs)
	default:
		return nil, fmt.Errorf(
			"read visit observations for profile %q: unsupported browser family %q",
			profile.Path,
			profile.BrowserFamily,
		)
	}
}

func (browserDataSource) HistoryPages(
	ctx context.Context,
	profile common.Profile,
	nativeURLIDs []int64,
) ([]common.HistoryPage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	switch profile.BrowserFamily {
	case "chromium":
		return chromium.FindHistoryPages(profile, nativeURLIDs)
	case "firefox":
		return firefox.FindHistoryPages(profile, nativeURLIDs)
	default:
		return nil, fmt.Errorf(
			"read history pages for profile %q: unsupported browser family %q",
			profile.Path,
			profile.BrowserFamily,
		)
	}
}
