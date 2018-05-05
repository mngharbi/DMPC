package daemon

import (
	"github.com/mngharbi/DMPC/config"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/pipeline"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
)

func startDaemons(conf *config.Config, shutdownLambda core.ShutdownLambda) {
	// Start users subsystem
	log.Debugf("Starting users subsystem")
	usersSubsystemConfig := conf.GetUsersSubsystemConfig()
	users.StartServer(usersSubsystemConfig, log, shutdownLambda)

	// Start status systems (status update and listeners servers)
	log.Debugf("Starting status subsystem")
	statusUpdateConfig, statusListenersConfig := conf.GetStatusSubsystemConfig()
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
	executorSubsystemConfig := conf.GetExecutorSubsystemConfig()
	executor.StartServer(executorSubsystemConfig)

	// Start decryptor subsystem
	log.Debugf("Starting decryptor subsystem")
	privateEncryptionKey, err := conf.GetPrivateEncryptionKey()
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
	decryptorSubsystemConfig := conf.GetDecryptorSubsystemConfig()
	decryptor.StartServer(decryptorSubsystemConfig)

	// Start pipeline subsystem (websocket server)
	log.Debugf("Starting pipeline subsystem")
	pipelineSubsystemConfig := conf.GetPipelineSubsystemConfig()
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

func Start() {
	// Setup listening on shutdown signals
	terminationChannel, shutdownLambda := setupShutdown()
	go shutdownWhenSignaled(terminationChannel)

	// Parse confuration and setup logging
	conf := doSetup()

	// Build user object from confuration files
	rootUserOperation := buildRootUserOperation(conf)

	// Start all subsystems
	startDaemons(conf, shutdownLambda)

	// Make root user request
	createRootUser(rootUserOperation)

	// Sleep forever (program is terminated by shutdown goroutine)
	select {}
}
