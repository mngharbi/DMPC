package status

import (
	"testing"
)

func TestStartShutdownSingleWorker(t *testing.T) {
	if !resetAndStartStatusServer(t, multipleWorkersStatusConfig()) {
		return
	}
	ShutdownStatusServer()
}
