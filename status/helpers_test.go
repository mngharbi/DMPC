/*
	Test helpers
*/

package status

import (
	"sync"
	"testing"
)

func resetAndStartStatusServer(t *testing.T, conf StatusServerConfig) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	serversStartWaitGroup.Add(1)
	statusServerSingleton = statusServer{}
	err := StartStatusServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func multipleWorkersStatusConfig() StatusServerConfig {
	return StatusServerConfig{
		NumWorkers: 6,
	}
}

func resetAndStartListenersServer(t *testing.T, conf ListenersServerConfig) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	serversStartWaitGroup.Add(1)
	listenersServerSingleton = listenersServer{}
	err := StartListenersServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func multipleWorkersListenersConfig() ListenersServerConfig {
	return ListenersServerConfig{
		NumWorkers: 6,
	}
}

func resetAndStartBothServers(t *testing.T, statusConf StatusServerConfig, listenersConf ListenersServerConfig) bool {
	statusServerSingleton = statusServer{}
	listenersServerSingleton = listenersServer{}
	err := StartServers(statusConf, listenersConf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}
