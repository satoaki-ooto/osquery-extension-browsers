package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/osquery/osquery-go"
)

func main() {
	socket := flag.String("socket", "", "Path to the osquery extension socket")
	flag.Parse()

	// Create the extension server
	server, err := osquery.NewExtensionManagerServer("browser_extend_extension", *socket)
	if err != nil {
		log.Fatalf("Error creating extension server: %v", err)
	}

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	// Wait for signal to stop
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	// Stop the server
	server.Shutdown(context.Background())
}
