// This is the main entry for go-wx. It loads the configuration file, sets up logging and starts up the
// web server. It also initializes the data collector ('interceptor') and the database connection.
// My preference is to make this a lightweight extensible application that can be used as a base for
// more complex weather station hardware and data processing applications while still providing a
// robust web interface for configuration and monitoring and reporting. Weather applications are
// frequently built by hobbyists who let them evolve into complex monoliths filled with features and
// legacy code. I want to build this as a modern application that can be easily extended and modified
// while still respecting the principles of reliability engineering. Specifically, I want to prioritize
// testability, observability, and maintainability.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ask-23/go-wx/pkg/config"
	"github.com/ask-23/go-wx/pkg/server"
)

func main() {
	// Define command line flags
	configFile := flag.String("config", "config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Print welcome message
	fmt.Println("go-wx: A modern Weather Station")
	fmt.Println("===============================")

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully")

	// Setup logging
	if err := setupLogging(cfg.Logging); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}
	log.Println("Logging configured")

	// TEMPORARY: Skip database connection for testing
	log.Println("Skipping database connection for testing")
	fmt.Println("NOTE: Database connection temporarily skipped")

	// Initialize the interceptor for data collection
	log.Println("Skipping data collector initialization for testing")
	fmt.Println("NOTE: Data collector temporarily skipped")

	// Skip publishers too
	log.Println("Skipping publishers for testing")
	fmt.Println("NOTE: Publishers temporarily skipped")

	// Initialize and start the web server
	fmt.Println("Initializing web server...")
	srv, err := server.NewServer(cfg.Server, cfg.Station, nil) // Pass nil for db
	if err != nil {
		log.Fatalf("Failed to initialize web server: %v", err)
	}
	fmt.Println("Web server initialized")

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Web server error: %v", err)
		}
	}()
	defer srv.Stop()

	fmt.Println("go-wx is running in TEST MODE (no database). Press Ctrl+C to stop.")

	// Wait for termination signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Shutting down go-wx...")
}

// setupLogging configures the application logging based on configuration
func setupLogging(cfg config.LoggingConfig) error {
	// Basic logging setup for now - can be enhanced with more sophisticated logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// If file logging is configured, set up file logging
	if cfg.File != "" {
		// Create logs directory if it doesn't exist
		dir := "logs"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				log.Printf("Warning: Could not create logs directory: %v", err)
				// Continue with console logging instead of returning error
			}
		}

		// Open log file
		f, err := os.OpenFile(cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Warning: Could not open log file %s: %v", cfg.File, err)
			// Continue with console logging instead of returning error
		} else {
			log.SetOutput(f)
		}
	}

	return nil
}
