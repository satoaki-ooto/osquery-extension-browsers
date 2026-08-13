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

	"osquery-extension-browsers/internal/osquerytables"
)

var debugMode bool

func main() {
	// Setup logging to both stdout and file
	logFile, err := os.OpenFile(
		"/tmp/browser_extend_extension.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
		log.Printf(
			"Configuration: socket=%s, timeout=%d, interval=%d, retry=%d, "+
				"retry-delay=%d, verbose=%v, debug=%v",
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

	debugLog("Registering browser table plugins...")
	server.RegisterPlugin(osquerytables.ProfilesPlugin())
	server.RegisterPlugin(osquerytables.VisitObservationsPlugin())
	server.RegisterPlugin(osquerytables.HistoryPagesPlugin())
	debugLog("✓ Plugins registered successfully")

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
		debugLog(
			"Socket not found (attempt %d/%d), waiting %d seconds...",
			attempt,
			maxAttempts,
			delaySeconds,
		)
		if attempt < maxAttempts {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}
	return fmt.Errorf("socket %s not found after %d attempts", socketPath, maxAttempts)
}
