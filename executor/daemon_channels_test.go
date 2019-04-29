package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/locker"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

/*
	Helpers
*/

func checkChannelLocking(t *testing.T, lockerCalls chan interface{}, expectedLockType core.LockType) bool {
	for i := 0; i < 2; i++ {
		lockCall := (<-lockerCalls).(*locker.LockerRequest)
		if lockCall.Type != locker.ChannelLock ||
			len(lockCall.Needs) != 1 ||
			lockCall.Needs[0].Id != genericChannelId ||
			lockCall.Needs[0].LockType != expectedLockType {
			t.Error("Channel add request should lock/unlock channel properly.")
			return false
		}
	}
	return true
}

/*
	Add channel request
*/

func TestAddChannelRequest(t *testing.T) {
	// Set up context needed
	usersRequesterDummy, userCalls := createDummyUsersRequesterFunctor(users.Success, userObjectsWithPermissions, nil, false)
	operationBufferer, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
	messageAdder, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
	channelActionRequester, channelActionCalls := createDummyChannelActionFunctor(channels.ChannelsSuccess, nil, false)
	lockerRequester, lockerCalls := createDummyLockerFunctor(true, nil, false)
	keyAdder, keyAdderCalls := createDummyKeyAdderFunctor(nil)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()

	rq := &channels.OpenChannelRequest{
		Channel: &channels.ChannelObject{
			KeyId: genericKeyId,
		},
		Key:       genericKey,
		Timestamp: nowTime,
	}
	meta := &core.OperationMetaFields{
		RequestType: core.AddChannelType,
		ChannelId:   genericChannelId,
		Timestamp:   nowTime,
	}
	rqEncoded, _ := rq.Encode()

	// Test valid request
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(true, meta, generateGenericSigners(), rqEncoded, nil)
	if err != nil {
		t.Error("Request should not fail.")
		ShutdownServer()
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.SuccessStatus {
		t.Errorf("Request should succeed and statuses should be reported correctly.")
	}

	// Check user read lock/unlock
	for i := 0; i < 2; i++ {
		userCall := <-userCalls
		userCallRq := &users.UserRequest{}
		userCallRq.Decode(userCall.request)
		if userCallRq.Type != users.ReadRequest ||
			len(userCallRq.Fields) != 1 ||
			userCallRq.Fields[0] != genericCertifierId ||
			userCall.readLock && userCall.readUnlock ||
			!userCall.readLock && !userCall.readUnlock {
			t.Error("Channel add request should read lock/unlock user.")
		}
	}

	// Check channel write lock/unlock
	checkChannelLocking(t, lockerCalls, core.WriteLockType)

	keyAdderCall := (<-keyAdderCalls).(keyAdderCall)
	if keyAdderCall.keyId != genericKeyId ||
		!reflect.DeepEqual(keyAdderCall.key, genericKey) {
		t.Error("Channel add request should add key.")
	}

	channelActionCall := (<-channelActionCalls).(*channels.OpenChannelRequest)
	expectedRq := &channels.OpenChannelRequest{}
	expectedRq.Decode(rqEncoded)
	expectedRq.Channel.Id = genericChannelId
	expectedRq.Signers = generateGenericSigners()
	if !reflect.DeepEqual(channelActionCall, expectedRq) {
		t.Errorf("Channel add request should be forwarded to channel action subsystem. expected=%+v, found=%+v", expectedRq, channelActionCall)
	}
}

func TestInvalidAddChannelRequest(t *testing.T) {
	usersRequesterDummy, _ := createDummyUsersRequesterFunctor(users.Success, userObjectsWithPermissions, nil, false)
	operationBufferer, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
	messageAdder, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
	channelActionRequester, _ := createDummyChannelActionFunctor(channels.ChannelsSuccess, nil, false)
	lockerRequester, _ := createDummyLockerFunctor(true, nil, false)
	keyAdder, _ := createDummyKeyAdderFunctor(nil)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()

	// Test request with nil channel object
	rq := &channels.OpenChannelRequest{
		Channel:   nil,
		Key:       genericKey,
		Timestamp: nowTime,
	}
	meta := &core.OperationMetaFields{
		RequestType: core.AddChannelType,
		ChannelId:   genericChannelId,
		Timestamp:   nowTime,
	}
	rqEncoded, _ := rq.Encode()

	// Make request
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
		return
	}
	ticketId, err := MakeRequest(true, meta, generateGenericSigners(), rqEncoded, nil)
	if err != nil {
		t.Error("Request should not be rejected.")
		ShutdownServer()
		return
	}
	ShutdownServer()

	// Check status
	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.FailedStatus ||
		reg.ticketLogs[ticketId][2].failureReason != status.RejectedReason ||
		!reflect.DeepEqual(reg.ticketLogs[ticketId][2].errors, []error{channelOpenNilChannelError}) {
		t.Errorf("Request should fail and statuses should be reported correctly.")
	}
}

func TestCloseChannelRequest(t *testing.T) {
	// Set up context needed
	usersRequesterDummy, _ := createDummyUsersRequesterFunctor(users.Success, nil, nil, false)
	operationBufferer, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
	messageAdder, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
	channelActionRequester, channelActionCalls := createDummyChannelActionFunctor(channels.ChannelsSuccess, nil, false)
	lockerRequester, lockerCalls := createDummyLockerFunctor(true, nil, false)
	keyAdder, _ := createDummyKeyAdderFunctor(nil)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()

	rq := &channels.CloseChannelRequest{
		Timestamp: nowTime,
	}
	rqEncoded, _ := rq.Encode()

	meta := &core.OperationMetaFields{
		RequestType: core.CloseChannelType,
		ChannelId:   genericChannelId,
		Timestamp:   nowTime,
	}

	// Test valid request
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(true, meta, generateGenericSigners(), rqEncoded, nil)
	if err != nil {
		t.Error("Request should not fail.")
		ShutdownServer()
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.SuccessStatus {
		t.Errorf("Request should succeed and statuses should be reported correctly.")
	}

	// Check channel write lock/unlock
	checkChannelLocking(t, lockerCalls, core.WriteLockType)

	channelActionCall := (<-channelActionCalls).(*channels.CloseChannelRequest)
	expectedRq := &channels.CloseChannelRequest{}
	expectedRq.Decode(rqEncoded)
	expectedRq.Id = genericChannelId
	expectedRq.Signers = generateGenericSigners()
	if !reflect.DeepEqual(channelActionCall, expectedRq) {
		t.Errorf("Channel close request should be forwarded to channel action subsystem. expected=%+v, found=%+v", expectedRq, channelActionCall)
	}
}

/*
	Message requests
*/

func doMessagesTesting(t *testing.T, isVerified bool, isBuffered bool) {
	// Set up context needed
	usersRequesterDummy, _ := createDummyUsersRequesterFunctor(users.Success, nil, nil, false)
	operationBufferer, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
	messageAdder, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
	channelActionRequester, _ := createDummyChannelActionFunctor(channels.ChannelsSuccess, nil, false)
	lockerRequester, lockerCalls := createDummyLockerFunctor(true, nil, false)
	keyAdder, _ := createDummyKeyAdderFunctor(nil)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()

	// Generic meta fields definition
	metaFields := core.OperationMetaFields{
		ChannelId:   genericChannelId,
		RequestType: core.AddMessageType,
		Timestamp:   nowTime,
	}

	// Generic operation defintion with generic meta
	requestOperationDefault := core.Operation{
		Meta: metaFields,
	}

	// Default empty operation (nil if not buffered)
	var requestOperation *core.Operation
	if isBuffered {
		requestOperation = &core.Operation{}
		*requestOperation = requestOperationDefault
	}

	// Test with request rejected directly from channels subsystem closure
	requestError := errors.New("Request Failed.")
	if isBuffered {
		operationBuffererFailing, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, requestError, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBuffererFailing, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, requestError, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdderFailing, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	}
	ticketId, err := MakeRequest(isVerified, &metaFields, generateGenericSigners(), []byte{}, requestOperation)
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.FailedStatus ||
		reg.ticketLogs[ticketId][2].failureReason != status.RejectedReason ||
		!reflect.DeepEqual(reg.ticketLogs[ticketId][2].errors, []error{requestError}) {
		t.Error("Request should run but fail, and statuses should be reported correctly when request is rejected.")
	}

	// Check channel read lock/unlock
	checkChannelLocking(t, lockerCalls, core.WriteLockType)

	// Test with channel closed from channels subsystem closure
	if isBuffered {
		operationBuffererFailing, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, true)
		*requestOperation = requestOperationDefault
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBuffererFailing, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, true)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdderFailing, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	}
	ticketId, err = MakeRequest(isVerified, &metaFields, generateGenericSigners(), []byte{}, requestOperation)
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.FailedStatus ||
		reg.ticketLogs[ticketId][2].failureReason != status.RejectedReason ||
		!reflect.DeepEqual(reg.ticketLogs[ticketId][2].errors, []error{subsystemChannelClosed}) {
		t.Error("Request should run but fail, and statuses should be reported correctly when channel closes.")
	}

	// Check channel read lock/unlock
	checkChannelLocking(t, lockerCalls, core.WriteLockType)

	// Test with failed requests
	if isBuffered {
		operationBuffererFailing, _ := createDummyOperationBuffererFunctor(1+channels.MessagesSuccess, nil, false)
		*requestOperation = requestOperationDefault
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBuffererFailing, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(1+channels.MessagesSuccess, nil, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdderFailing, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	}
	ticketId, err = MakeRequest(isVerified, &metaFields, generateGenericSigners(), []byte{}, requestOperation)
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.FailedStatus ||
		reg.ticketLogs[ticketId][2].failureReason != status.FailedReason ||
		reg.ticketLogs[ticketId][2].errors != nil {
		t.Error("Request should run but fail, and statuses should be reported correctly when the request failed.")
	}

	// Check channel read lock/unlock
	checkChannelLocking(t, lockerCalls, core.WriteLockType)

	// Test with one successful request
	if isBuffered {
		operationBuffererFailing, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
		*requestOperation = requestOperationDefault
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBuffererFailing, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdderFailing, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	}
	ticketId, err = MakeRequest(isVerified, &metaFields, generateGenericSigners(), []byte{}, requestOperation)
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.SuccessStatus {
		t.Error("Request should succeed.")
	}

	// Test with concurrent successful requests and check calls made
	var callsChannel chan interface{}
	if isBuffered {
		var operationBuffererFailing channels.OperationBufferer
		operationBuffererFailing, callsChannel = createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdder, operationBuffererFailing, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	} else {
		var messageAdderFailing channels.MessageAdder
		messageAdderFailing, callsChannel = createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterDummy, usersRequesterDummy, messageAdderFailing, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator) {
			return
		}
	}

	var wg sync.WaitGroup
	wg.Add(10)
	checksumExpected := 0
	for i := 1; i <= 10; i++ {
		checksumExpected += i
		copyI := i
		go (func() {
			waitForRandomDuration()
			if isBuffered {
				payload := []byte(strconv.Itoa(copyI))
				op := requestOperationDefault
				op.Payload = string(payload)
				_, _ = MakeRequest(isVerified, &op.Meta, generateGenericSigners(), []byte{}, &op)
			} else {
				payload := []byte(strconv.Itoa(copyI))
				_, _ = MakeRequest(isVerified, &metaFields, generateGenericSigners(), payload, nil)
			}
			wg.Done()
		})()
	}
	wg.Wait()

	ShutdownServer()

	checksum := 0
	for i := 1; i <= 10; i++ {
		callLog := <-callsChannel
		if isBuffered {
			nb, _ := strconv.Atoi(string(callLog.(*channels.BufferOperationRequest).Operation.Payload))
			checksum += nb
		} else {
			addMessageRequest := callLog.(*channels.AddMessageRequest)
			if addMessageRequest.Timestamp != nowTime ||
				!reflect.DeepEqual(addMessageRequest.Signers, generateGenericSigners()) ||
				addMessageRequest.ChannelId != genericChannelId {
				t.Errorf("Unexpected call made to messages subsystem. Request=%+v", addMessageRequest)
				return
			} else {
				nb, _ := strconv.Atoi(string(addMessageRequest.Message))
				checksum += nb
			}

			// Check channel read lock/unlock
			checkChannelLocking(t, lockerCalls, core.WriteLockType)
		}
	}
	if checksum != checksumExpected {
		t.Errorf("Payload didn't make it through as expected. checksum=%v, checksumExpected=%v", checksum, checksumExpected)
	}
}

func TestUnverifiedAddMessage(t *testing.T) {
	doMessagesTesting(t, false, false)
}

func TestVerifiedAddMessage(t *testing.T) {
	doMessagesTesting(t, true, false)
}

func TestUnverifiedBufferOperation(t *testing.T) {
	doMessagesTesting(t, false, true)
}

func TestVerifiedBufferOperation(t *testing.T) {
	doMessagesTesting(t, true, true)
}
