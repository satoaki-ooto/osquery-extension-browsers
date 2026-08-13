package osquerytables

import (
	"context"
	"strconv"

	osquerygo "github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/table"
)

// HistoryPagesPlugin returns the current browser URL/page state table plugin.
func HistoryPagesPlugin() osquerygo.OsqueryPlugin {
	return historyPagesPlugin(browserDataSource{})
}

func historyPagesPlugin(source dataSource) osquerygo.OsqueryPlugin {
	columns := []table.ColumnDefinition{
		table.TextColumn("profile_id"),
		table.BigIntColumn("native_url_id"),
		table.TextColumn("browser_family"),
		table.TextColumn("browser_variant"),
		table.TextColumn("url"),
		table.TextColumn("title"),
		table.IntegerColumn("title_present"),
		table.IntegerColumn("visit_count"),
		table.BigIntColumn("last_visit_time"),
		table.IntegerColumn("hidden"),
	}

	return newIndexedTablePlugin(
		"browser_history_pages",
		columns,
		func(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
			return generateHistoryPages(ctx, queryContext, source)
		},
		"profile_id",
		"native_url_id",
	)
}

func generateHistoryPages(
	ctx context.Context,
	queryContext table.QueryContext,
	source dataSource,
) ([]map[string]string, error) {
	profiles, err := source.Profiles(ctx)
	if err != nil {
		return nil, err
	}

	profileIDs := equalityStrings(queryContext, "profile_id")
	nativeURLIDs := equalityInt64s(queryContext, "native_url_id")

	var rows []map[string]string
	for _, profile := range profiles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !matchesEquality(profile.ProfileID, profileIDs) {
			continue
		}

		pages, err := source.HistoryPages(ctx, profile, nativeURLIDs)
		if err != nil {
			return nil, err
		}
		for _, page := range pages {
			row := map[string]string{
				"profile_id":      page.ProfileID,
				"native_url_id":   strconv.FormatInt(page.NativeURLID, 10),
				"browser_family":  page.BrowserFamily,
				"browser_variant": page.BrowserVariant,
				"url":             page.URL,
				"title_present":   boolString(page.TitlePresent),
				"visit_count":     strconv.FormatInt(page.VisitCount, 10),
			}
			if page.TitlePresent {
				row["title"] = page.Title
			}
			if page.LastVisitTimePresent {
				row["last_visit_time"] = strconv.FormatInt(page.LastVisitTime, 10)
			}
			if page.HiddenPresent {
				row["hidden"] = strconv.FormatInt(page.Hidden, 10)
			}
			rows = append(rows, row)
		}
	}
	return rows, nil
}
