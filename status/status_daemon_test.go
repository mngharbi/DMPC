package status

import (
	"testing"
)

func TestStatusStartShutdownMultipleWorker(t *testing.T) {
	if !resetAndStartStatusServer(t, multipleWorkersStatusConfig()) {
		return
	}
	ShutdownStatusServer()
}
