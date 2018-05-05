package daemon

/*
	Startup utilities
*/

import (
	"github.com/mngharbi/DMPC/config"
	"github.com/mngharbi/DMPC/core"
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
	if !config.IsFunctional() {
		log.Fatalf(setupError)
	}
}

// Main function for startup
func doSetup() (conf *config.Config) {
	// Initialize logging
	log = core.InitializeLogging()
	log.SetLogLevel(core.DEBUG)

	// Check DMPC was configured
	log.Debugf("Checking DMPC install configuration")
	checkSetup()

	// Get configuration structure
	log.Debugf("Parsing configuration")
	conf = config.GetConfig()

	// Set log level from configuration
	log.SetLogLevel(conf.LogLevel)

	return
}
