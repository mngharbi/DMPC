package daemon

/*
	Debug messages
*/
const (
	// Starting up subsystems
	startingUsersSubsystemLogMsg     string = "Starting users subsystem"
	startingChannelsSubsystemLogMsg  string = "Starting channels subsystem"
	startingStatusSubsystemLogMsg    string = "Starting status subsystem"
	startingKeysSubsystemLogMsg      string = "Starting keys subsystem"
	startingExecutorSubsystemLogMsg  string = "Starting executor subsystem"
	startingDecryptorSubsystemLogMsg string = "Starting decryptor subsystem"
	startingPipelineSubsystemLogMsg  string = "Starting pipeline subsystem"

	// Shutting down subsystems
	shutdownUsersSubsystemLogMsg     string = "Shutting down users subsystem"
	shutdownChannelsSubsystemLogMsg  string = "Shutting down channels subsystem"
	shutdownStatusSubsystemLogMsg    string = "Shutting down status subsystem"
	shutdownKeysSubsystemLogMsg      string = "Shutting down keys subsystem"
	shutdownExecutorSubsystemLogMsg  string = "Shutting down executor subsystem"
	shutdownDecryptorSubsystemLogMsg string = "Shutting down decryptor subsystem"
	shutdownPipelineSubsystemLogMsg  string = "Shutting down pipeline subsystem"

	checkingInstallLogMsg      string = "Checking DMPC install configuration"
	parsingConfigurationLogMsg string = "Parsing configuration"
)

/*
	Info messages
*/
const (
	startingUpSubsystemsInfoMsg string = "Starting up subsystems"
	createRootUserInfoMsg       string = "Initializing root user"
)

/*
	Failure messages
*/
const (
	inaccessiblePrivateEncryptionKeyErrorMsg string = "Unable to access private encryption key. Error: %v"
)
