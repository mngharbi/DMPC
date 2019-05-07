package executor

import (
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/keys"
	"github.com/mngharbi/DMPC/locker"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"testing"
)

/*
	General tests
*/

func createDummies(
	responseReporterSuccess bool,
) (
	users.Requester,
	chan userRequesterCall,
	users.Requester,
	chan userRequesterCall,
	channels.MessageAdder,
	chan interface{},
	channels.OperationBufferer,
	chan interface{},
	channels.ChannelActionRequester,
	chan interface{},
	channels.ListenersRequester,
	chan interface{},
	locker.Requester,
	chan interface{},
	core.KeyAdder,
	chan interface{},
	keys.Encryptor,
	chan interface{},
	status.Reporter,
	*dummyStatusRegistry,
	status.TicketGenerator,
) {
	usersRequester, usersRequesterCh := createDummyUsersRequesterFunctor(users.Success, userObjectsWithPermissions, nil, false)
	usersRequesterUnverified, usersRequesterUnverifiedCh := createDummyUsersRequesterFunctor(users.Success, userObjectsWithPermissions, nil, false)
	messageAdder, messageAdderCalls := createDummyMessageAdderFunctor(channels.MessagesSuccess, nil, false)
	operationBufferer, operationBuffererCalls := createDummyOperationBuffererFunctor(channels.MessagesSuccess, nil, false)
	channelActionRequester, channelActionCalls := createDummyChannelActionFunctor(channels.ChannelsSuccess, nil, false)
	channelListenersRequester, channelListenersCalls := createDummyListenersRequesterFunctor(channels.ListenersSuccess, nil, false)
	lockerRequester, lockerCalls := createDummyLockerFunctor(true, nil, false)
	keyAdder, keyAdderCalls := createDummyKeyAdderFunctor(nil)
	keyEncryptor, keyEncryptorCalls := createDummyKeyEncryptorFunctor(nil)
	responseReporter, statusReg := createDummyResposeReporterFunctor(responseReporterSuccess)
	ticketGenerator := createDummyTicketGeneratorFunctor()

	return usersRequester, usersRequesterCh, usersRequesterUnverified, usersRequesterUnverifiedCh, messageAdder, messageAdderCalls, operationBufferer, operationBuffererCalls, channelActionRequester, channelActionCalls, channelListenersRequester, channelListenersCalls, lockerRequester, lockerCalls, keyAdder, keyAdderCalls, keyEncryptor, keyEncryptorCalls, responseReporter, statusReg, ticketGenerator
}

func TestStartShutdownServer(t *testing.T) {
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, _, channelListenersRequester, _, lockerRequester, _, keyAdder, _, keyEncryptor, _, responseReporter, _, ticketGenerator := createDummies(true)
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}
	ShutdownServer()
}

func TestInvalidRequestType(t *testing.T) {
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, _, channelListenersRequester, _, lockerRequester, _, keyAdder, _, keyEncryptor, _, responseReporter, _, ticketGenerator := createDummies(true)
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	_, err := MakeRequest(false, &core.OperationMetaFields{RequestType: core.UsersRequestType - 1, Timestamp: nowTime}, generateGenericSigners(), []byte{}, nil)
	if err != invalidRequestTypeError {
		t.Error("Request with invalid type should be rejected.")
	}

	ShutdownServer()
}

func TestReponseReporterQueueError(t *testing.T) {
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, _, channelListenersRequester, _, lockerRequester, _, keyAdder, _, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(false)
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(false, &core.OperationMetaFields{RequestType: core.UsersRequestType, Timestamp: nowTime}, generateGenericSigners(), []byte{}, nil)
	if err != responseReporterError {
		t.Error("Request should fail with response reporter error while queueing.")
	}

	if len(reg.ticketLogs[ticketId]) != 0 {
		t.Error("Status for ticket number should be empty if queueing failed.")
	}

	ShutdownServer()
}

func TestRequestWhileNotRunning(t *testing.T) {
	usersRequester, _, usersRequesterUnverified, _, messageAdder, _, operationBufferer, _, channelActionRequester, _, channelListenersRequester, _, lockerRequester, _, keyAdder, _, keyEncryptor, _, responseReporter, reg, ticketGenerator := createDummies(true)
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, channelListenersRequester, lockerRequester, keyAdder, keyEncryptor, responseReporter, ticketGenerator) {
		return
	}

	ShutdownServer()

	ticketId, err := MakeRequest(false, &core.OperationMetaFields{RequestType: core.UsersRequestType, Timestamp: nowTime}, generateGenericSigners(), []byte{}, nil)
	if err == nil {
		t.Error("Request should fail if made while server is down.")
	}

	if len(reg.ticketLogs[ticketId]) != 2 ||
		reg.ticketLogs[ticketId][0].status != status.QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != status.FailedStatus ||
		reg.ticketLogs[ticketId][1].failureReason != status.RejectedReason {
		t.Error("Status for ticket number should be updated if failing when server is down.")
	}
}
