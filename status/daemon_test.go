package status

import (
	"testing"
)

func TestStartShutdown(t *testing.T) {
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig()) {
		return
	}
	ShutdownServers()
}
