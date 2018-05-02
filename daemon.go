package main

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
)

func startDaemons(config *Config, shutdownLambda core.ShutdownLambda) {
	// Start users subsystem
	log.Debugf("Starting users subsystem")
	usersSubsystemConfig := config.getUsersSubsystemConfig()
	users.StartServer(usersSubsystemConfig, log, shutdownLambda)

	// Start status systems (status update and listeners servers)
	log.Debugf("Starting status subsystem")
	statusUpdateConfig, statusListenersConfig := config.getStatusSubsystemConfig()
	status.StartServers(statusUpdateConfig, statusListenersConfig, log, shutdownLambda)

	// Start executor subsystem
	log.Debugf("Starting executor subsystem")
	executor.InitializeServer(
		users.MakeRequest,
		users.MakeUnverifiedRequest,
		status.UpdateStatus,
		status.RequestNewTicket,
		log,
		shutdownLambda,
	)
	executorSubsystemConfig := config.getExecutorSubsystemConfig()
	executor.StartServer(executorSubsystemConfig)
}

func shutdownDaemons() {
	log.Debugf("Shutting down users subsystem")
	users.ShutdownServer()

	log.Debugf("Shutting down executor subsystem")
	executor.ShutdownServer()

	log.Debugf("Shutting down status subsystem")
	status.ShutdownServers()
}
