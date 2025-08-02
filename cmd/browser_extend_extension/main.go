package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/table"

	"osquery-extension-browsers/internal/browsers/chromium"
	"osquery-extension-browsers/internal/browsers/firefox"
)

func main() {
	// フラグの定義（osqueryが自動的に渡してくれる）
	socket := flag.String("socket", "", "Path to osquery socket file")
	timeout := flag.Int("timeout", 3, "Seconds to wait for autoloaded extensions")
	interval := flag.Int("interval", 3, "Seconds delay between connectivity checks")
	flag.Parse()

	if *socket == "" {
		log.Fatalf("Missing required --socket argument")
	}

	// Create extension manager with proper timeout settings
	serverTimeout := osquery.ServerTimeout(time.Duration(*timeout) * time.Second)
	serverPingInterval := osquery.ServerPingInterval(time.Duration(*interval) * time.Second)

	server, err := osquery.NewExtensionManagerServer(
		"browser_extend_extension",
		*socket,
		serverTimeout,
		serverPingInterval,
		osquery.ExtensionVersion("1.0.0"),
	)
	if err != nil {
		log.Fatalf("Failed to create extension: %v", err)
	}

	// Create the table plugin
	browserHistoryTable := browserHistoryTablePlugin()
	server.RegisterPlugin(browserHistoryTable)

	// Run the server
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run extension: %v", err)
	}
}

// browserHistoryTablePlugin creates a table plugin for browser history
func browserHistoryTablePlugin() *table.Plugin {
	columns := []table.ColumnDefinition{
		table.TextColumn("time"),
		table.TextColumn("title"),
		table.TextColumn("url"),
		table.TextColumn("profile"),
		table.TextColumn("browser_type"),
	}

	return table.NewPlugin("browser_history", columns, generateBrowserHistory)
}

// generateBrowserHistory generates the browser history data for the table
func generateBrowserHistory(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	var results []map[string]string

	// Find Chromium profiles
	chromiumProfiles, err := chromium.FindProfiles()
	if err != nil {
		log.Printf("Failed to find Chromium profiles: %v", err)
	} else {
		// Get history for each Chromium profile
		for _, profile := range chromiumProfiles {
			historyEntries, err := chromium.FindHistory(profile)
			if err != nil {
				log.Printf("Failed to find Chromium history for profile %s: %v", profile.ID, err)
				continue
			}

			// Add history entries to results
			for _, entry := range historyEntries {
				results = append(results, map[string]string{
					"time":            entry.VisitTime.Format("2006-01-02 15:04:05"),
					"url":             entry.URL,
					"title":           entry.Title,
					"visit_count":     string(rune(entry.VisitCount)),
					"profile":         entry.ProfileID,
					"browser_type":    entry.BrowserType,
					"browser_variant": entry.BrowserVariant,
				})
			}
		}
	}

	// Find Firefox profiles
	firefoxProfiles, err := firefox.FindProfiles()
	if err != nil {
		log.Printf("Failed to find Firefox profiles: %v", err)
	} else {
		// Get history for each Firefox profile
		for _, profile := range firefoxProfiles {
			historyEntries, err := firefox.FindHistory(profile)
			if err != nil {
				log.Printf("Failed to find Firefox history for profile %s: %v", profile.ID, err)
				continue
			}

			// Add history entries to results
			for _, entry := range historyEntries {
				results = append(results, map[string]string{
					"time":            entry.VisitTime.Format("2006-01-02 15:04:05"),
					"url":             entry.URL,
					"title":           entry.Title,
					"visit_count":     string(rune(entry.VisitCount)),
					"profile":         entry.ProfileID,
					"browser_type":    entry.BrowserType,
					"browser_variant": entry.BrowserVariant,
				})
			}
		}
	}

	return results, nil
}
