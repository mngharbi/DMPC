package main

/*
	Shutdown utilities
*/

import (
	"github.com/mngharbi/DMPC/core"
	"os"
	"os/signal"
	"syscall"
)

/*
	Termination messages
*/
var terminationCauseMessageMapping map[TerminationCause]string = map[TerminationCause]string{
	FatalError:       "Fatal runtime error occured",
	UserInterrupted:  "Detected user interruption",
	SystemTerminated: "Detected system termination",
}

/*
	Termination causes specification
*/
type TerminationCause int

const (
	NoTermination TerminationCause = iota
	FatalError
	UserInterrupted
	SystemTerminated
)

/*
	List of fatal system signals handled and their mapping
*/
var signalMapping map[os.Signal]TerminationCause = map[os.Signal]TerminationCause{
	os.Interrupt:    UserInterrupted,
	syscall.SIGHUP:  SystemTerminated,
	syscall.SIGINT:  UserInterrupted,
	syscall.SIGTERM: SystemTerminated,
	syscall.SIGQUIT: SystemTerminated,
}

func mapSystemSignal(sig os.Signal) TerminationCause {
	return signalMapping[sig]
}
func isTerminal(terminationCause TerminationCause) bool {
	return terminationCause != NoTermination
}

func listenForSystemTermination(terminationChannel chan TerminationCause) {
	// Put everything from system into signalChannel
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel)

	// Keep waiting on signals and push them to termination channel
	for {
		systemSignal := <-signalChannel
		mappedTerminationCause := mapSystemSignal(systemSignal)
		terminationChannel <- mappedTerminationCause
	}
}

func listenForTermination(terminationChannel chan TerminationCause) {
	// Setup system termination listening
	go listenForSystemTermination(terminationChannel)

	// Keep waiting on causes until a terminal cause is sent
	for {
		terminationCause := <-terminationChannel
		if isTerminal(terminationCause) {
			log.Errorf(terminationCauseMessageMapping[terminationCause])
			return
		}
	}
}

/*
	Generates shutdown functor to be passed to subsystems
*/
func shutdownFunctor(terminationChannel chan TerminationCause) core.ShutdownLambda {
	return func() {
		terminationChannel <- FatalError
	}
}

func setupShutdown() (chan TerminationCause, core.ShutdownLambda) {
	terminationChannel := make(chan TerminationCause)
	return terminationChannel, shutdownFunctor(terminationChannel)
}

func shutdownWhenSignaled(terminationChannel chan TerminationCause) {
	// Wait until signal to terminate is received
	listenForTermination(terminationChannel)

	// Soft shutdown all subsystems
	shutdownDaemons()

	// Terminate program
	os.Exit(1)
}
