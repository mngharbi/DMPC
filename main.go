package main

import (
	"github.com/mngharbi/DMPC/core"
)

var (
	log *core.LoggingHandler
)

func main() {
	// Initialize logging
	log = core.InitializeLogging()
	log.SetLogLevel(core.DEBUG)

	// Check DMPC was configured
	log.Debugf("Checking DMPC install configuration")
	checkSetup()

	// Get configuration structure
	log.Debugf("Parsing configuration")
	config := getConfig()

	// Set log level from configuration
	log.SetLogLevel(config.LogLevel)

	// Get root user object
	log.Debugf("Parsing root user object")
	_ = config.getRootUserObject()

	return
}
