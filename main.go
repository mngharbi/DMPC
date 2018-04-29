package main

func main() {
	// Setup listening on shutdown signals
	terminationChannel, _ := setupShutdown()
	go shutdownWhenSignaled(terminationChannel)

	// Parse configuration and setup logging
	config := doSetup()

	// Build user object from configuration files
	_ = buildRootUserObject(config)

	// Start all subsystems
	startDaemons(config)

	// Sleep forever (program is terminated by shutdown goroutine)
	select {}
}
