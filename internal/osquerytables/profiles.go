package osquerytables

import (
	"context"

	osquerygo "github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/table"
)

// ProfilesPlugin returns the browser profile inventory table plugin.
func ProfilesPlugin() osquerygo.OsqueryPlugin {
	return profilesPlugin(browserDataSource{})
}

func profilesPlugin(source dataSource) osquerygo.OsqueryPlugin {
	columns := []table.ColumnDefinition{
		table.TextColumn("profile_id"),
		table.TextColumn("browser_family"),
		table.TextColumn("browser_variant"),
		table.TextColumn("os_user_name"),
		table.TextColumn("profile_directory"),
		table.TextColumn("profile_path"),
		table.TextColumn("profile_display_name"),
		table.IntegerColumn("profile_display_name_present"),
		table.TextColumn("profile_account"),
		table.IntegerColumn("profile_account_present"),
	}

	return newIndexedTablePlugin(
		"browser_profiles",
		columns,
		func(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
			return generateProfiles(ctx, queryContext, source)
		},
		"profile_id",
	)
}

func generateProfiles(
	ctx context.Context,
	queryContext table.QueryContext,
	source dataSource,
) ([]map[string]string, error) {
	profiles, err := source.Profiles(ctx)
	if err != nil {
		return nil, err
	}

	profileIDs := equalityStrings(queryContext, "profile_id")
	rows := make([]map[string]string, 0, len(profiles))
	for _, profile := range profiles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !matchesEquality(profile.ProfileID, profileIDs) {
			continue
		}

		row := map[string]string{
			"profile_id":                   profile.ProfileID,
			"browser_family":               profile.BrowserFamily,
			"browser_variant":              profile.BrowserVariant,
			"os_user_name":                 profile.OSUserName,
			"profile_directory":            profile.Directory,
			"profile_path":                 profile.Path,
			"profile_display_name_present": boolString(profile.DisplayNamePresent),
			"profile_account_present":      boolString(profile.AccountPresent),
		}
		if profile.DisplayNamePresent {
			row["profile_display_name"] = profile.DisplayName
		}
		if profile.AccountPresent {
			row["profile_account"] = profile.Account
		}
		rows = append(rows, row)
	}
	return rows, nil
}
