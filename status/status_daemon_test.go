package status

import (
	"testing"
)

func TestStatusStartShutdown(t *testing.T) {
	if !resetAndStartStatusServer(t, multipleWorkersStatusConfig()) {
		return
	}
	ShutdownStatusServer()
}

func TestInvalidStatusUpdate(t *testing.T) {
	err := UpdateStatus(RequestNewTicket(), FailedStatus+1, NoReason, nil, nil)
	if err != statusRangeError {
		t.Errorf("Request with invalid status code should fail. err=%v", err)
	}

	err = UpdateStatus(RequestNewTicket(), FailedStatus, FailedReason+1, nil, nil)
	if err != failedRangeError {
		t.Errorf("Request with invalid failure code should fail. err=%v", err)
	}
}
