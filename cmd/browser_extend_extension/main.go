package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/table"

	"osquery-extension-browsers/internal/browsers/chromium"
	"osquery-extension-browsers/internal/browsers/firefox"
	"osquery-extension-browsers/internal/diff_table"
)

var debugMode bool

func main() {
	// Setup logging to both stdout and file
	logFile, err := os.OpenFile("/tmp/browser_extend_extension.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("=== Starting browser extension with periodic diff support ===")

	socket := flag.String("socket", "", "Path to osquery socket file")
	timeout := flag.Int("timeout", 60, "Seconds to wait for autoloaded extensions")
	interval := flag.Int("interval", 5, "Seconds delay between connectivity checks")
	retryAttempts := flag.Int("retry", 5, "Number of retry attempts for connection")
	retryDelay := flag.Int("retry-delay", 2, "Delay in seconds between retry attempts")
	verbose := flag.Bool("verbose", false, "Enable verbose logging (osquery compatibility)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	debugMode = *debug

	if debugMode {
		log.Println("=== Extension Starting (Debug Mode) ===")
		log.Printf("Configuration: socket=%s, timeout=%d, interval=%d, retry=%d, retry-delay=%d, verbose=%v, debug=%v",
			*socket, *timeout, *interval, *retryAttempts, *retryDelay, *verbose, *debug)
	}

	if *socket == "" {
		log.Fatalf("Missing required --socket argument")
	}

	// Wait for socket to be available with retry logic
	if err := waitForSocket(*socket, *retryAttempts, *retryDelay); err != nil {
		log.Fatalf("Socket not available after retries: %v", err)
	}

	serverTimeout := osquery.ServerTimeout(time.Duration(*timeout) * time.Second)
	serverPingInterval := osquery.ServerPingInterval(time.Duration(*interval) * time.Second)

	// Retry extension server creation
	var server *osquery.ExtensionManagerServer
	for attempt := 1; attempt <= *retryAttempts; attempt++ {
		debugLog("Attempt %d/%d: Creating extension manager server...", attempt, *retryAttempts)
		server, err = osquery.NewExtensionManagerServer(
			"browser_extend_extension",
			*socket,
			serverTimeout,
			serverPingInterval,
			osquery.ExtensionVersion("1.0.0"),
		)
		if err == nil {
			debugLog("✓ Extension manager server created successfully")
			break
		}
		log.Printf("Failed to create extension (attempt %d/%d): %v", attempt, *retryAttempts, err)
		if attempt < *retryAttempts {
			time.Sleep(time.Duration(*retryDelay) * time.Second)
		}
	}
	if err != nil {
		log.Fatalf("Failed to create extension after %d attempts: %v", *retryAttempts, err)
	}

	debugLog("Registering browser history table plugins...")
	
	// Register the standard browser history table
	browserHistoryTable := browserHistoryTablePlugin()
	server.RegisterPlugin(browserHistoryTable)
	debugLog("✓ Standard browser_history table registered")

	// Register the periodic diff table
	diffTable, err := diff_table.New()
	if err != nil {
		log.Printf("Failed to create diff table: %v", err)
	} else {
		// Register the diff table plugin
		diffTablePlugin := table.NewPlugin(
			"browser_history_diff",
			diffTable.Columns(),
			diffTable.Generate,
		)
		server.RegisterPlugin(diffTablePlugin)
		debugLog("✓ browser_history_diff table registered")

		// Ensure the diff table is properly closed on shutdown
		defer diffTable.Close()
	}

	// Setup signal handling
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigc
		log.Println("Received shutdown signal, cleaning up...")
		server.Shutdown(context.Background())
		os.Exit(0)
	}()

	debugLog("Starting extension server (this will block)...")
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run extension: %v", err)
	}
	debugLog("Extension server stopped")
}

// debugLog logs a message only when debug mode is enabled
func debugLog(format string, v ...interface{}) {
	if debugMode {
		log.Printf(format, v...)
	}
}

// waitForSocket waits for the osquery socket to be available
func waitForSocket(socketPath string, maxAttempts, delaySeconds int) error {
	debugLog("Waiting for socket: %s", socketPath)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check if socket file exists
		if _, err := os.Stat(socketPath); err == nil {
			debugLog("✓ Socket found on attempt %d/%d", attempt, maxAttempts)
			return nil
		}
		debugLog("Socket not found (attempt %d/%d), waiting %d seconds...", attempt, maxAttempts, delaySeconds)
		if attempt < maxAttempts {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}
	return fmt.Errorf("socket %s not found after %d attempts", socketPath, maxAttempts)
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
