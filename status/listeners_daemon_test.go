package status

import (
	"testing"
)

func TestListenersStartShutdownMultipleWorker(t *testing.T) {
	if !resetAndStartListenersServer(t, multipleWorkersListenersConfig()) {
		return
	}
	ShutdownListenersServer()
}

func TestAddListenerServerDown(t *testing.T) {
	ch, err := AddListener(RequestNewTicket())
	if err == nil {
		t.Errorf("Add listener while listeners server is down should fail.")
		return
	}

	_, isOpen := <- ch
	if isOpen {
		t.Errorf("Add listener while listeners server is down should close channel.")
	}
}
