package main

import (
	"github.com/mngharbi/DMPC/users"
)

func startDaemons(config *Config) {
	// Start users subsystem
	log.Debugf("Starting users subsystem")
	usersSubsystemConfig := config.getUsersSubsystemConfig()
	users.StartServer(log, usersSubsystemConfig)
}

func shutdownDaemons() {
	log.Debugf("Shutting down users subsystem")
	users.ShutdownServer()
}
