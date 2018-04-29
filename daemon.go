package main

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
)

func startDaemons(config *Config, shutdownLambda core.ShutdownLambda) {
	// Start users subsystem
	log.Debugf("Starting users subsystem")
	usersSubsystemConfig := config.getUsersSubsystemConfig()
	users.StartServer(usersSubsystemConfig, log, shutdownLambda)
}

func shutdownDaemons() {
	log.Debugf("Shutting down users subsystem")
	users.ShutdownServer()
}
