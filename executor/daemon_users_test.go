package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

/*
	User generic requests
*/

func doUserRequestTesting(t *testing.T, isVerified bool) {
	// Set up context needed
	usersRequesterDummy, _ := createDummyUsersRequesterFunctor(users.Success, nil, nil, false)
	operationBufferer, _ := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
	messageAdder, _ := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
	channelActionRequester, _ := createDummyChannelActionFunctor(channels.ChannelsSuccess, nil, false)
	channelListenersRequester, _ := createDummyListenersRequesterFunctor(channels.ListenersSuccess, nil, false)
	lockerRequester, _ := createDummyLockerFunctor(true, nil, false)
	keyAdder, _ := createDummyKeyAdderFunctor(nil)
	keyEncryptor, _ := createDummyKeyEncryptorFunctor(nil)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	var usersRequester, usersRequesterVerified users.Requester

	// Test with request rejected directly from users requester
	requestError := errors.New("Request Failed.")
	usersRequesterFailing, _ := createDummyUsersRequesterFunctor(users.Success, nil, []error{requestError}, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterFailing, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterFailing, usersRequesterDummy
	}

	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(isVerified, &core.OperationMetaFields{RequestType: core.UsersRequestType, Timestamp: nowTime}, generateGenericSigners(), []byte{}, nil)
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

	// Test with channel closed from users requester
	usersRequesterSuccess, _ := createDummyUsersRequesterFunctor(users.Success, nil, nil, true)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterSuccess, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterSuccess, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}
	ticketId, err = MakeRequest(isVerified, &core.OperationMetaFields{RequestType: core.UsersRequestType, Timestamp: nowTime}, generateGenericSigners(), []byte{}, nil)
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

	// Test with failed requests
	usersRequesterUnsuccessfulResponse, _ := createDummyUsersRequesterFunctor(1+users.Success, nil, nil, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterUnsuccessfulResponse, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterUnsuccessfulResponse, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}
	ticketId, err = MakeRequest(isVerified, &core.OperationMetaFields{RequestType: core.UsersRequestType, Timestamp: nowTime}, generateGenericSigners(), []byte{}, nil)
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

	// Test with one successful request
	usersRequesterSuccessfulResponse, _ := createDummyUsersRequesterFunctor(users.Success, nil, nil, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterSuccessfulResponse, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterSuccessfulResponse, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}
	ticketId, err = MakeRequest(isVerified, &core.OperationMetaFields{RequestType: core.UsersRequestType, Timestamp: nowTime}, generateGenericSigners(), []byte{}, nil)
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

	// Test with concurrent successful requests and check calls made to users subsystem
	usersRequesterSuccessfulResponseMultiple, callsChannel := createDummyUsersRequesterFunctor(users.Success, nil, nil, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterSuccessfulResponseMultiple, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterSuccessfulResponseMultiple, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	var wg sync.WaitGroup
	wg.Add(10)
	checksumExpected := 0
	for i := 1; i <= 10; i++ {
		checksumExpected += i
		copyI := i
		go (func() {
			waitForRandomDuration()
			payload := []byte(strconv.Itoa(copyI))
			_, _ = MakeRequest(isVerified, &core.OperationMetaFields{RequestType: core.UsersRequestType, Timestamp: nowTime}, generateGenericSigners(), payload, nil)
			wg.Done()
		})()
	}
	wg.Wait()

	ShutdownServer()

	checksum := 0
	for i := 1; i <= 10; i++ {
		callLog := <-callsChannel
		if callLog.signers.IssuerId != genericIssuerId ||
			callLog.signers.CertifierId != genericCertifierId {
			t.Error("Unexpected call made to users subsystem.")
			return
		} else {
			nb, _ := strconv.Atoi(string(callLog.request))
			checksum += nb
		}
	}
	if checksum != checksumExpected {
		t.Errorf("Payload didn't make it through as expected. checksum=%v, checksumExpected=%v", checksum, checksumExpected)
	}
}

func TestUnverifiedUserRequest(t *testing.T) {
	doUserRequestTesting(t, false)
}

func TestVerifiedUserRequest(t *testing.T) {
	doUserRequestTesting(t, true)
}
