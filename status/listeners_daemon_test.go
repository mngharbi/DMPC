package status

import (
	"testing"
)

func TestListenersStartShutdownMultipleWorker(t *testing.T) {
	if !resetAndStartListenersServer(t, multipleWorkersListenersConfig()) {
		return
	}
	ShutdownStatusServer()
}
