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
	if !config.IsFunctional() {
		log.Fatalf(setupErrorMsg)
	}
}

// Main function for startup
func doSetup() (conf *config.Config) {
	// Initialize logging
	log = core.InitializeLogging()
	log.SetLogLevel(core.DEBUG)

	// Check DMPC was configured
	log.Debugf(checkingInstallLogMsg)
	checkInstall()

	// Get configuration structure
	log.Debugf(parsingConfigurationLogMsg)
	conf = config.GetConfig()

	// Set log level from configuration
	log.SetLogLevel(conf.LogLevel)

	return
}
