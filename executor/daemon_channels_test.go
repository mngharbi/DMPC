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

func checkUserLocking(t *testing.T, userCalls chan userRequesterCall) bool {
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
	return true
}

/*
	Read channel request
*/

func TestReadChannelRequest(t *testing.T) {
	// Set up context needed
	usersRequester, _, usersRequesterUnverified, userCalls, messageAdder, _, operationBufferer, _, channelActionRequester, channelActionCalls, channelListenersRequester, _, lockerRequester, _, keyAdder, _, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(true)

	rq := &channels.ReadChannelRequest{}
	meta := &core.OperationMetaFields{
		RequestType: core.ReadChannelType,
		ChannelId:   genericChannelId,
		Timestamp:   nowTime,
	}
	rqEncoded, _ := rq.Encode()

	// Test valid request
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(true, meta, generateGenericSigners(), rqEncoded, nil)
	if err != nil {
		t.Error("Request should not fail.")
		ShutdownServer()
		return
	}

	ShutdownServer()

	// Check status
	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.SuccessStatus {
		t.Errorf("Request should succeed and statuses should be reported correctly.")
	}

	// Expect user read locking
	checkUserLocking(t, userCalls)

	// Check channel subsystem call
	channelActionCall := (<-channelActionCalls).(*channels.ReadChannelRequest)
	expectedRq := &channels.ReadChannelRequest{
		Id: genericChannelId,
	}
	if !reflect.DeepEqual(channelActionCall, expectedRq) {
		t.Errorf("Channel read request should be forwarded to channel action subsystem. expected=%+v, found=%+v", expectedRq, channelActionCall)
	}
}

/*
	Add channel request
*/

func TestAddChannelRequest(t *testing.T) {
	// Set up context needed
	usersRequester, _, usersRequesterUnverified, userCalls, messageAdder, _, operationBufferer, _, channelActionRequester, channelActionCalls, channelListenersRequester, _, lockerRequester, lockerCalls, keyAdder, keyAdderCalls, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(true)

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
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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
	checkUserLocking(t, userCalls)

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
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, _, channelListenersRequester, _, lockerRequester, _, keyAdder, _, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(true)

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
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, channelActionCalls, channelListenersRequester, _, lockerRequester, lockerCalls, keyAdder, _, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(true)

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
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, _, channelListenersRequester, _, lockerRequester, lockerCalls, keyAdder, _, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(true)

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
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBuffererFailing, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, requestError, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdderFailing, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBuffererFailing, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, true)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdderFailing, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBuffererFailing, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(1+channels.MessagesSuccess, nil, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdderFailing, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBuffererFailing, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
			return
		}
	} else {
		messageAdderFailing, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdderFailing, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBuffererFailing, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
			return
		}
	} else {
		var messageAdderFailing channels.MessageAdder
		messageAdderFailing, callsChannel = createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
		if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdderFailing, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
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

func TestSubscribeRequest(t *testing.T) {
	// Set up context needed
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, _, channelListenersRequester, subscribeCalls, lockerRequester, _, keyAdder, _, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(true)

	meta := &core.OperationMetaFields{
		RequestType: core.SubscribeChannelType,
		ChannelId:   genericChannelId,
		Timestamp:   nowTime,
	}

	// Test valid request
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(true, meta, generateGenericSigners(), nil, nil)
	if err != nil {
		t.Error("Request should not fail.")
		ShutdownServer()
		return
	}

	ShutdownServer()

	// Check status
	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.RunningStatus ||
		reg.ticketLogs[ticketId][2].status != status.SuccessStatus {
		t.Errorf("Request should succeed and statuses should be reported correctly.")
	}

	// Check response in status
	channelResp := reg.ticketLogs[ticketId][2].result.(*channels.ListenersResponse)
	if channelResp.Result != channels.ListenersSuccess {
		t.Errorf("Status should include full response")
	}

	// Check channel subsystem call
	subscribeCall := (<-subscribeCalls).(*channels.SubscribeRequest)
	expectedRq := &channels.SubscribeRequest{
		ChannelId: genericChannelId,
		Signers:   generateGenericSigners(),
	}
	if !reflect.DeepEqual(subscribeCall, expectedRq) {
		t.Errorf("Channel subscribe request should be forwarded to channel listeners subsystem. expected=%+v, found=%+v", expectedRq, subscribeCall)
	}
}

func TestChannelEncryptRequest(t *testing.T) {
	// Set up context needed
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, channelActionCalls, channelListenersRequester, _, lockerRequester, lockerCalls, keyAdder, _, keyEncryptor, keyEncryptorCalls, responseReporter, reg, ticketGenerator := createDummies(true)

	innerPlaintextBytes := []byte{25}
	innerMeta := core.OperationMetaFields{
		RequestType: core.AddMessageType,
		ChannelId:   genericChannelId2,
		Timestamp:   nowTime,
	}
	op := &core.Operation{
		Meta:    innerMeta,
		Payload: core.Base64EncodeToString(innerPlaintextBytes),
	}
	opEncoded, _ := op.Encode()

	meta := &core.OperationMetaFields{
		RequestType: core.ChannelEncryptType,
		ChannelId:   genericChannelId,
		Timestamp:   nowTime,
	}

	// Test valid request
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(true, meta, generateGenericSigners(), opEncoded, nil)
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

	// Check channel read lock/unlock
	checkChannelLocking(t, lockerCalls, core.ReadLockType)

	// Check channel read operation
	channelActionCall := (<-channelActionCalls).(*channels.ReadChannelRequest)
	expectedRq := &channels.ReadChannelRequest{
		Id: genericChannelId,
	}
	if !reflect.DeepEqual(channelActionCall, expectedRq) {
		t.Errorf("Channel encrypt request should read channel from channels subsystem. expected=%+v, found=%+v", expectedRq, channelActionCall)
	}

	// Check encryption calls
	keyEncryptorCall := (<-keyEncryptorCalls).(keyEncryptorCall)
	if keyEncryptorCall.keyId != genericKeyId ||
		!reflect.DeepEqual(keyEncryptorCall.payload, innerPlaintextBytes) {
		t.Error("Channel encrypt request should call keys subsystem with inner decoded payload to encrypt.")
	}

	// Check final operation
	resultOp := reg.ticketLogs[ticketId][2].result.(core.Operation)
	if !resultOp.Encryption.Encrypted ||
		resultOp.Encryption.KeyId != genericKeyId ||
		resultOp.Meta.ChannelId != genericChannelId {
		t.Errorf("Channel encrypt should set fields correctly in resulting operation. resultOp=%+v", resultOp)
	}
}
