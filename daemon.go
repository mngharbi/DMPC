package main

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/pipeline"
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

	// Start decryptor subsystem
	log.Debugf("Starting decryptor subsystem")
	privateEncryptionKey, err := config.getPrivateEncryptionKey()
	if err != nil {
		log.Fatalf("Unable to access private encryption key. Error: %v", err.Error())
	}
	decryptor.InitializeServer(
		privateEncryptionKey,
		users.GetSigningKeysById,

		// @TODO: Replace with actual closure when keys subsystem is implemented
		(func(string) []byte { return nil }),

		executor.MakeRequest,
		log,
		shutdownLambda,
	)
	decryptorSubsystemConfig := config.getDecryptorSubsystemConfig()
	decryptor.StartServer(decryptorSubsystemConfig)

	// Start pipeline subsystem (websocket server)
	log.Debugf("Starting pipeline subsystem")
	pipelineSubsystemConfig := config.getPipelineSubsystemConfig()
	pipeline.StartServer(pipelineSubsystemConfig, decryptor.MakeRequest, log)
}

func shutdownDaemons() {
	log.Debugf("Shutting down pipeline subsystem")
	pipeline.ShutdownServer()

	log.Debugf("Shutting down decryptor subsystem")
	decryptor.ShutdownServer()

	log.Debugf("Shutting down users subsystem")
	users.ShutdownServer()

	log.Debugf("Shutting down executor subsystem")
	executor.ShutdownServer()

	log.Debugf("Shutting down status subsystem")
	status.ShutdownServers()
}
