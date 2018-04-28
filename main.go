package main

func main() {
	go shutdownWhenSignaled()

	// Parse configuration and setup logging
	config := doSetup()

	// Build user object from configuration files
	_ = buildRootUserObject(config)

	// Start all subsystems
	startDaemons(config)

	// Sleep forever (program is terminated by shutdown goroutine)
	select {}
}
