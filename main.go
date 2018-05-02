package main

func main() {
	// Setup listening on shutdown signals
	terminationChannel, shutdownLambda := setupShutdown()
	go shutdownWhenSignaled(terminationChannel)

	// Parse configuration and setup logging
	config := doSetup()

	// Build user object from configuration files
	rootUserOperation := buildRootUserOperation(config)

	// Start all subsystems
	startDaemons(config, shutdownLambda)

	// Make root user request
	_ = createRootUser(rootUserOperation)

	// Sleep forever (program is terminated by shutdown goroutine)
	select {}
}
