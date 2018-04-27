package main

import (
	"github.com/mngharbi/DMPC/core"
)

func main() {
	// Set initial log level for startup
	core.SetLogLevel(initialLogLevel)

	// Check DMPC was configured
	core.Debugf("Checking DMPC install configuration")
	checkSetup()

	// Get configuration structure
	core.Debugf("Parsing configuration")
	config := getConfig()

	// Set log level from configuration
	core.SetLogLevel(config.LogLevel)

	// Get root user object
	core.Debugf("Parsing root user object")
	userObject := config.getRootUserObject()

	return
}
