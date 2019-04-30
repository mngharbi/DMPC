package daemon

import (
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/cli"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/keys"
	"github.com/mngharbi/DMPC/locker"
	"github.com/mngharbi/DMPC/pipeline"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
)

func startDaemons(conf *cli.Config, shutdownLambda core.ShutdownLambda) {
	// Start locker subsystem
	log.Debugf(startingLockerSubsystemLogMsg)
	lockerSubsystemConfig := conf.GetLockerSubsystemConfig()
	locker.StartServer(lockerSubsystemConfig, log, shutdownLambda)

	// Start users subsystem
	log.Debugf(startingUsersSubsystemLogMsg)
	usersSubsystemConfig := conf.GetUsersSubsystemConfig()
	users.StartServer(usersSubsystemConfig, log, shutdownLambda)

	// Start channels subsystem
	log.Debugf(startingChannelsSubsystemLogMsg)
	channelsMainSubsystemConfig, channelsMessagesSubsystemConfig, channelsListenersSubsystemConfig := conf.GetChannelsSubsystemConfig()
	channels.StartServers(channelsMainSubsystemConfig, channelsMessagesSubsystemConfig, channelsListenersSubsystemConfig, decryptor.MakeOperationRequest, log, shutdownLambda)

	// Start status systems (status update and listeners servers)
	log.Debugf(startingStatusSubsystemLogMsg)
	statusUpdateConfig, statusListenersConfig := conf.GetStatusSubsystemConfig()
	status.StartServers(statusUpdateConfig, statusListenersConfig, log, shutdownLambda)

	// Start keys subsystem
	log.Debugf(startingKeysSubsystemLogMsg)
	keysSubsystemConfig := conf.GetKeysSubsystemConfig()
	keys.StartServer(keysSubsystemConfig, log, shutdownLambda)

	// Start executor subsystem
	log.Debugf(startingExecutorSubsystemLogMsg)
	executor.InitializeServer(
		users.MakeRequest,
		users.MakeUnverifiedRequest,
		channels.AddMessage,
		channels.BufferOperation,
		channels.ChannelAction,
		channels.ListenerAction,
		locker.RequestLock,
		keys.AddKey,
		keys.Encrypt,
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
		keys.Decrypt,
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

	log.Debugf(shutdownKeysSubsystemLogMsg)
	keys.ShutdownServer()

	log.Debugf(shutdownUsersSubsystemLogMsg)
	users.ShutdownServer()

	log.Debugf(shutdownChannelsSubsystemLogMsg)
	channels.ShutdownServers()

	log.Debugf(shutdownExecutorSubsystemLogMsg)
	executor.ShutdownServer()

	log.Debugf(shutdownStatusSubsystemLogMsg)
	status.ShutdownServers()

	log.Debugf(shutdownLockerSubsystemLogMsg)
	locker.ShutdownServer()
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
