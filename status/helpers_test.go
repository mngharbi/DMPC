/*
	Test helpers
*/

package status

import (
	"sync"
	"testing"
)

func startStatusServerAndTest(t *testing.T, conf StatusServerConfig) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	serversStartWaitGroup.Add(1)
	if err := startStatusServer(conf); err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func resetAndStartStatusServer(t *testing.T, conf StatusServerConfig) bool {
	statusServerSingleton = statusServer{}
	return startStatusServerAndTest(t, conf)
}

func multipleWorkersStatusConfig() StatusServerConfig {
	return StatusServerConfig{
		NumWorkers: 6,
	}
}

func startListenersServerAndTest(t *testing.T, conf ListenersServerConfig) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	serversStartWaitGroup.Add(1)
	if err := startListenersServer(conf); err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func resetAndStartListenersServer(t *testing.T, conf ListenersServerConfig) bool {
	listenersServerSingleton = listenersServer{}
	return startListenersServerAndTest(t, conf)
}

func multipleWorkersListenersConfig() ListenersServerConfig {
	return ListenersServerConfig{
		NumWorkers: 6,
	}
}

func startBothServersAndTest(t *testing.T, statusConf StatusServerConfig, listenersConf ListenersServerConfig, ignoreError bool) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	if err := StartServers(statusConf, listenersConf); err != nil {
		if !ignoreError {
			t.Errorf(err.Error())
		}
		return false
	}
	return true
}

func resetAndStartBothServers(t *testing.T, statusConf StatusServerConfig, listenersConf ListenersServerConfig, ignoreError bool) bool {
	statusServerSingleton = statusServer{}
	listenersServerSingleton = listenersServer{}
	return startBothServersAndTest(t, statusConf, listenersConf, ignoreError)
}
