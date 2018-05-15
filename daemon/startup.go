package daemon

/*
	Startup utilities
*/

import (
	"github.com/mngharbi/DMPC/startup"
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
	setupErrorMsg string = "DMPC not properly configured"
)

/*
	Startup constants
*/
const (
	initialLogLevel core.LogLevel = core.INFO
)

// Checks if DMPC was set up
func checkInstall() {
	if !startup.IsFunctional() {
		log.Fatalf(setupErrorMsg)
	}
}

// Main function for startup
func doSetup() (conf *startup.Config) {
	// Initialize logging
	log = core.InitializeLogging()
	log.SetLogLevel(core.INFO)

	// Check DMPC was configured
	log.Debugf(checkingInstallLogMsg)
	checkInstall()

	// Get configuration structure
	log.Debugf(parsingConfigurationLogMsg)
	conf = startup.GetConfig()

	// Set log level from configuration
	log.SetLogLevel(conf.LogLevel)

	return
}
