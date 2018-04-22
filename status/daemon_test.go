package status

import (
	"sync"
	"testing"
)

func TestStartShutdown(t *testing.T) {
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	ShutdownServers()
}

func TestStartServersFailure(t *testing.T) {
	serversStartWaitGroup = sync.WaitGroup{}
	statusServerSingleton = statusServer{}
	listenersServerSingleton = listenersServer{}
	statusConf := multipleWorkersStatusConfig()
	listenersConf := multipleWorkersListenersConfig()

	// Test status server failure
	if !startStatusServerAndTest(t, statusConf) {
		return
	}
	serversStartWaitGroup.Wait()
	if startBothServersAndTest(t, statusConf, listenersConf, true) {
		t.Errorf("Servers start should fail if status server was up.")
		return
	}
	shutdownStatusServer()

	// Test listeners server failure
	if !startListenersServerAndTest(t, listenersConf) {
		return
	}
	serversStartWaitGroup.Wait()
	if startBothServersAndTest(t, statusConf, listenersConf, true) {
		t.Errorf("Servers start should fail if listeners server was up.")
		return
	}
	shutdownListenersServer()
}
