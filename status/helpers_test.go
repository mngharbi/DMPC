/*
	Test helpers
*/

package status

import (
	"testing"
)

func resetAndStartStatusServer(t *testing.T, conf StatusServerConfig) bool {
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
