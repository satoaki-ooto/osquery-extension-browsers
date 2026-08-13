package osquerytables

import (
	"context"
	"reflect"
	"testing"

	osquerygo "github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/table"

	"osquery-extension-browsers/internal/browsers/common"
)

type sourceCall struct {
	profileID    string
	nativeURLIDs []int64
}

type fakeDataSource struct {
	profiles         []common.Profile
	observations     map[string][]common.VisitObservation
	pages            map[string][]common.HistoryPage
	observationCalls []sourceCall
	pageCalls        []sourceCall
}

func (source *fakeDataSource) Profiles(context.Context) ([]common.Profile, error) {
	return source.profiles, nil
}

func (source *fakeDataSource) VisitObservations(
	_ context.Context,
	profile common.Profile,
	nativeURLIDs []int64,
) ([]common.VisitObservation, error) {
	source.observationCalls = append(source.observationCalls, sourceCall{
		profileID:    profile.ProfileID,
		nativeURLIDs: nativeURLIDs,
	})
	return source.observations[profile.ProfileID], nil
}

func (source *fakeDataSource) HistoryPages(
	_ context.Context,
	profile common.Profile,
	nativeURLIDs []int64,
) ([]common.HistoryPage, error) {
	source.pageCalls = append(source.pageCalls, sourceCall{
		profileID:    profile.ProfileID,
		nativeURLIDs: nativeURLIDs,
	})
	return source.pages[profile.ProfileID], nil
}

func TestTableSchemasExposeIndexedJoinColumns(t *testing.T) {
	tests := []struct {
		name           string
		plugin         osquerygo.OsqueryPlugin
		indexedColumns []string
	}{
		{
			name:           "browser_profiles",
			plugin:         ProfilesPlugin(),
			indexedColumns: []string{"profile_id"},
		},
		{
			name:           "browser_visit_observations",
			plugin:         VisitObservationsPlugin(),
			indexedColumns: []string{"obs_id", "profile_id", "native_url_id"},
		},
		{
			name:           "browser_history_pages",
			plugin:         HistoryPagesPlugin(),
			indexedColumns: []string{"profile_id", "native_url_id"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.plugin.Name() != test.name {
				t.Errorf("plugin name = %q, want %q", test.plugin.Name(), test.name)
			}
			options := make(map[string]string)
			for _, column := range test.plugin.Routes() {
				options[column["name"]] = column["op"]
			}
			for _, column := range test.indexedColumns {
				if options[column] != indexColumnOption {
					t.Errorf("column %q options = %q, want %q", column, options[column], indexColumnOption)
				}
			}
		})
	}
}

func TestVisitObservationConstraintsAvoidUnrelatedProfileReads(t *testing.T) {
	source := &fakeDataSource{
		profiles: []common.Profile{
			{ProfileID: "p1", BrowserFamily: "chromium"},
			{ProfileID: "p2", BrowserFamily: "firefox"},
		},
		observations: map[string][]common.VisitObservation{
			"p2": {
				{
					ObservationID:  "keep",
					ProfileID:      "p2",
					BrowserFamily:  "firefox",
					BrowserVariant: "firefox",
					NativeVisitID:  12,
					NativeURLID:    7,
					VisitTime:      1710000000,
					VisitTimeUS:    1710000000123456,
				},
				{ObservationID: "drop", ProfileID: "p2", NativeURLID: 7},
			},
		},
	}
	queryContext := table.QueryContext{Constraints: map[string]table.ConstraintList{
		"profile_id": {
			Constraints: []table.Constraint{{Operator: table.OperatorEquals, Expression: "p2"}},
		},
		"obs_id": {
			Constraints: []table.Constraint{{Operator: table.OperatorEquals, Expression: "keep"}},
		},
		"native_url_id": {
			Constraints: []table.Constraint{{Operator: table.OperatorEquals, Expression: "7"}},
		},
	}}

	rows, err := generateVisitObservations(context.Background(), queryContext, source)
	if err != nil {
		t.Fatalf("generateVisitObservations() error = %v", err)
	}
	if len(rows) != 1 || rows[0]["obs_id"] != "keep" {
		t.Fatalf("generateVisitObservations() rows = %#v, want only keep", rows)
	}
	wantCalls := []sourceCall{{profileID: "p2", nativeURLIDs: []int64{7}}}
	if !reflect.DeepEqual(source.observationCalls, wantCalls) {
		t.Errorf("observation calls = %#v, want %#v", source.observationCalls, wantCalls)
	}
}

func TestNullableColumnsAreOmittedAndPresenceIsExplicit(t *testing.T) {
	source := &fakeDataSource{
		profiles: []common.Profile{
			{
				ProfileID:          "p1",
				BrowserFamily:      "firefox",
				BrowserVariant:     "firefox",
				OSUserName:         "os-user",
				Directory:          "abc.default",
				Path:               "/profiles/abc.default",
				DisplayNamePresent: false,
				AccountPresent:     false,
			},
		},
		pages: map[string][]common.HistoryPage{
			"p1": {
				{
					ProfileID:      "p1",
					NativeURLID:    9,
					BrowserFamily:  "firefox",
					BrowserVariant: "firefox",
					URL:            "https://example.com/",
					TitlePresent:   false,
					VisitCount:     1,
				},
			},
		},
	}

	profileRows, err := generateProfiles(context.Background(), table.QueryContext{}, source)
	if err != nil {
		t.Fatalf("generateProfiles() error = %v", err)
	}
	if _, present := profileRows[0]["profile_display_name"]; present {
		t.Error("unavailable profile_display_name key is present")
	}
	if _, present := profileRows[0]["profile_account"]; present {
		t.Error("unavailable profile_account key is present")
	}
	if profileRows[0]["profile_display_name_present"] != "0" ||
		profileRows[0]["profile_account_present"] != "0" {
		t.Errorf("profile presence columns = %#v", profileRows[0])
	}

	pageRows, err := generateHistoryPages(context.Background(), table.QueryContext{}, source)
	if err != nil {
		t.Fatalf("generateHistoryPages() error = %v", err)
	}
	for _, column := range []string{"title", "last_visit_time", "hidden"} {
		if _, present := pageRows[0][column]; present {
			t.Errorf("unavailable %s key is present", column)
		}
	}
	if pageRows[0]["title_present"] != "0" {
		t.Errorf("title_present = %q, want 0", pageRows[0]["title_present"])
	}
}
