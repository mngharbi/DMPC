package main

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
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

	// Get root user object from configuration
	log.Debugf("Parsing root user object from configuration")
	_ = config.getRootUserObject()

	// Start users subsystem
	log.Debugf("Starting users subsystem")
	usersSubsystemConfig := config.getUsersSubsystemConfig()
	users.StartServer(usersSubsystemConfig)

}
