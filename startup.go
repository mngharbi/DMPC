package main

/*
	Startup utilities
*/

import (
	"github.com/mngharbi/DMPC/core"
	"os"
)

/*
	Package wide logging handler
*/
var (
	log *core.LoggingHandler
)

/*
	Error messages
*/
const (
	setupError string = "DMPC not properly configured"
)

/*
	Startup constants
*/
const (
	initialLogLevel core.LogLevel = core.INFO
)

// Checks if DMPC was set up
func checkSetup() {
	if _, err := os.Stat(getSetupFilename()); err != nil {
		log.Fatalf(setupError)
	}
}

// Main function for startup
func doSetup() (config *Config) {
	// Initialize logging
	log = core.InitializeLogging()
	log.SetLogLevel(core.DEBUG)

	// Check DMPC was configured
	log.Debugf("Checking DMPC install configuration")
	checkSetup()

	// Get configuration structure
	log.Debugf("Parsing configuration")
	config = getConfig()

	// Set log level from configuration
	log.SetLogLevel(config.LogLevel)

	return
}
