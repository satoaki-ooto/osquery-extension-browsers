package osquerytables

import (
	"context"
	"strconv"

	osquerygo "github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/table"
)

// VisitObservationsPlugin returns the stable native browser visit table plugin.
func VisitObservationsPlugin() osquerygo.OsqueryPlugin {
	return visitObservationsPlugin(browserDataSource{})
}

func visitObservationsPlugin(source dataSource) osquerygo.OsqueryPlugin {
	columns := []table.ColumnDefinition{
		table.TextColumn("obs_id"),
		table.TextColumn("profile_id"),
		table.TextColumn("browser_family"),
		table.TextColumn("browser_variant"),
		table.BigIntColumn("native_visit_id"),
		table.BigIntColumn("native_url_id"),
		table.BigIntColumn("visit_time"),
		table.BigIntColumn("visit_time_us"),
	}

	return newIndexedTablePlugin(
		"browser_visit_observations",
		columns,
		func(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
			return generateVisitObservations(ctx, queryContext, source)
		},
		"obs_id",
		"profile_id",
		"native_url_id",
	)
}

func generateVisitObservations(
	ctx context.Context,
	queryContext table.QueryContext,
	source dataSource,
) ([]map[string]string, error) {
	profiles, err := source.Profiles(ctx)
	if err != nil {
		return nil, err
	}

	profileIDs := equalityStrings(queryContext, "profile_id")
	observationIDs := equalityStrings(queryContext, "obs_id")
	nativeURLIDs := equalityInt64s(queryContext, "native_url_id")

	var rows []map[string]string
	for _, profile := range profiles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !matchesEquality(profile.ProfileID, profileIDs) {
			continue
		}

		observations, err := source.VisitObservations(ctx, profile, nativeURLIDs)
		if err != nil {
			return nil, err
		}
		for _, observation := range observations {
			if !matchesEquality(observation.ObservationID, observationIDs) {
				continue
			}
			rows = append(rows, map[string]string{
				"obs_id":          observation.ObservationID,
				"profile_id":      observation.ProfileID,
				"browser_family":  observation.BrowserFamily,
				"browser_variant": observation.BrowserVariant,
				"native_visit_id": strconv.FormatInt(observation.NativeVisitID, 10),
				"native_url_id":   strconv.FormatInt(observation.NativeURLID, 10),
				"visit_time":      strconv.FormatInt(observation.VisitTime, 10),
				"visit_time_us":   strconv.FormatInt(observation.VisitTimeUS, 10),
			})
		}
	}
	return rows, nil
}
