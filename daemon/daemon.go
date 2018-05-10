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
	log.Debugf(startingUsersSubsystemLogMsg)
	usersSubsystemConfig := conf.GetUsersSubsystemConfig()
	users.StartServer(usersSubsystemConfig, log, shutdownLambda)

	// Start status systems (status update and listeners servers)
	log.Debugf(startingStatusSubsystemLogMsg)
	statusUpdateConfig, statusListenersConfig := conf.GetStatusSubsystemConfig()
	status.StartServers(statusUpdateConfig, statusListenersConfig, log, shutdownLambda)

	// Start executor subsystem
	log.Debugf(startingExecutorSubsystemLogMsg)
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
	log.Debugf(startingDecryptorSubsystemLogMsg)
	privateEncryptionKey, err := conf.GetPrivateEncryptionKey()
	if err != nil {
		log.Fatalf(inaccessiblePrivateEncryptionKeyErrorMsg, err.Error())
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
	log.Debugf(startingPipelineSubsystemLogMsg)
	pipelineSubsystemConfig := conf.GetPipelineSubsystemConfig()
	pipeline.StartServer(pipelineSubsystemConfig, decryptor.MakeTransactionRequest, log)
}

func shutdownDaemons() {
	log.Debugf(shutdownPipelineSubsystemLogMsg)
	pipeline.ShutdownServer()

	log.Debugf(shutdownDecryptorSubsystemLogMsg)
	decryptor.ShutdownServer()

	log.Debugf(shutdownUsersSubsystemLogMsg)
	users.ShutdownServer()

	log.Debugf(shutdownExecutorSubsystemLogMsg)
	executor.ShutdownServer()

	log.Debugf(shutdownStatusSubsystemLogMsg)
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
	log.Infof(startingUpSubsystemsInfoMsg)
	startDaemons(conf, shutdownLambda)

	// Make root user request
	log.Infof(createRootUserInfoMsg)
	createRootUser(rootUserOperation)

	// Sleep forever (program is terminated by shutdown goroutine)
	select {}
}
