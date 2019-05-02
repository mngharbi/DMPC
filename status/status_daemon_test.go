package status

import (
	"testing"
)

func TestStatusStartShutdown(t *testing.T) {
	if !resetAndStartStatusServer(t, multipleWorkersStatusConfig()) {
		return
	}
	shutdownStatusServer()
}

func TestInvalidStatusUpdate(t *testing.T) {
	err := UpdateStatus(RequestNewTicket(), StatusCode("random_status"), NoReason, nil, nil)
	if err != statusRangeError {
		t.Errorf("Request with invalid status code should fail. err=%v", err)
	}

	err = UpdateStatus(RequestNewTicket(), FailedStatus, FailReasonCode("random_status"), nil, nil)
	if err != failedRangeError {
		t.Errorf("Request with invalid failure code should fail. err=%v", err)
	}
}

func TestStatusUpdateServerDown(t *testing.T) {
	err := UpdateStatus(RequestNewTicket(), QueuedStatus, NoReason, nil, nil)
	if err == nil {
		t.Errorf("Request while server is down fails.")
	}
}
