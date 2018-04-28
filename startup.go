package main

/*
	Startup utilities
*/

import (
	"github.com/mngharbi/DMPC/core"
	"os"
)

/*
	Error messages
*/
const (
	setupError string = "DMPC not properly configured"
)

/*
	Startup constants
*/
const (
	initialLogLevel core.LogLevel = core.INFO
)

// Checks if DMPC was set up
func checkSetup() {
	if _, err := os.Stat(getSetupFilename()); err != nil {
		log.Fatalf(setupError)
	}
}
