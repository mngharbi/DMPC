package locker

import (
	"github.com/mngharbi/DMPC/core"
	"sync"
	"testing"
)

/*
	General tests
*/

func TestStartShutdownSingleWorker(t *testing.T) {
	if !resetAndStartServer(t, singleWorkerConfig()) {
		return
	}
	ShutdownServer()
}

func TestStartShutdown(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}
	ShutdownServer()
}

/*
	Test requests
*/

const (
	channelId1 string = "CHANNEL_ID_1"
	channelId2 string = "CHANNEL_ID_2"
)

var (
	validLockerRequest LockerRequest = LockerRequest{
		Type: ChannelLock,
		Needs: []core.LockNeed{
			{
				LockType: core.WriteLockType,
				Id:       channelId1,
			},
		},
		LockingType: core.Locking,
	}
)

func TestInvalidRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Invalid resource type
	invalidResourceTypeRequest := validLockerRequest
	invalidResourceTypeRequest.Type = ChannelLock + 1
	_, errs := RequestLock(&invalidResourceTypeRequest)
	if len(errs) != 1 || errs[0].Error() != unknownResourceTypeErrorMsg {
		t.Errorf("Request with invalid resource type must be rejected. errs=%+v", errs[0].Error())
	}

	// nil needs list
	nilNeedsRequest := validLockerRequest
	nilNeedsRequest.Needs = nil
	_, errs = RequestLock(&nilNeedsRequest)
	if len(errs) != 1 || errs[0].Error() != noNeedsErrorMsg {
		t.Errorf("Request with nil needs list must be rejected. errs=%+v", errs[0].Error())
	}

	// empty needs list
	emptyNeedsRequest := validLockerRequest
	emptyNeedsRequest.Needs = []core.LockNeed{}
	_, errs = RequestLock(&emptyNeedsRequest)
	if len(errs) != 1 || errs[0].Error() != noNeedsErrorMsg {
		t.Errorf("Request with empty needs list must be rejected. errs=%+v", errs[0].Error())
	}

	ShutdownServer()
}

func TestOneLockUnlock(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Build lock/unlock requests for same resource
	lockRequest := validLockerRequest
	lockRequest.LockingType = core.Locking
	unlockRequest := validLockerRequest
	unlockRequest.LockingType = core.Unlocking

	// Request to lock
	resChannel, errs := RequestLock(&lockRequest)
	if len(errs) != 0 {
		t.Errorf("Valid lock request should not be rejected. errs=%+v", errs)
	}
	lockResult := <-resChannel
	if lockResult != true {
		t.Errorf("Valid lock request should not fail.")
	}

	// Request to unlock
	resChannel, errs = RequestLock(&unlockRequest)
	if len(errs) != 0 {
		t.Errorf("Valid unlock request should not be rejected. errs=%+v", errs)
	}
	unlockResult := <-resChannel
	if unlockResult != true {
		t.Errorf("Valid unlock request should not fail.")
	}

	ShutdownServer()
}

func testValidRequest(t *testing.T, request *LockerRequest, locking bool, expectation bool, msg string) {
	lockingString := "lock"
	if !locking {
		lockingString = "unlock"
	}

	resChannel, errs := RequestLock(request)
	if len(errs) != 0 {
		t.Errorf("Valid %v request should not be rejected. errs=%+v", lockingString, errs)
	}
	result := <-resChannel
	if result != expectation {
		t.Errorf(msg)
	}
}

func TestMultipleResourceLockUnlock(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Build lock/unlock requests for two resources
	lockRequest := validLockerRequest
	lockRequest.LockingType = core.Locking
	addLockingNeed(&lockRequest, core.WriteLockType, channelId2)
	unlockRequest := validLockerRequest
	unlockRequest.LockingType = core.Unlocking
	addLockingNeed(&unlockRequest, core.WriteLockType, channelId2)

	/*
		Sequential locking/unlocking
	*/

	// Request to unlock before lock
	testValidRequest(t, &unlockRequest, false, false, "Unlock before lock request should fail.")

	// Request to lock
	testValidRequest(t, &lockRequest, true, true, "Valid lock request should not fail.")

	// Request to unlock
	testValidRequest(t, &unlockRequest, false, true, "Valid unlock request should not fail.")

	/*
		Concurrent locking
	*/

	group := &sync.WaitGroup{}
	group.Add(1)

	go func() {
		testValidRequest(t, &lockRequest, true, true, "Valid lock request should not fail.")
		group.Done()
	}()
	go func() {
		// Reversed needs
		reverseLockRequest := lockRequest
		reverseLockRequest.Needs = []core.LockNeed{lockRequest.Needs[1], lockRequest.Needs[0]}
		testValidRequest(t, &reverseLockRequest, true, true, "Reverse order lock request should not fail.")
		group.Done()
	}()

	group.Wait()

	// Unlock once and wait for second goroutine to take lock
	group.Add(2)
	go func() {
		testValidRequest(t, &unlockRequest, false, true, "Valid unlock request should not fail.")
		group.Done()
	}()
	group.Wait()

	// Unlock one last time
	group.Add(1)
	go func() {
		testValidRequest(t, &unlockRequest, false, true, "Valid unlock request should not fail.")
		group.Done()
	}()
	group.Wait()

	ShutdownServer()
}
