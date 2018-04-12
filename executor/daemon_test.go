package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/users"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

/*
	Dummy subsystem lambdas
*/

func sendUserResponseAfterRandomDelay(channel chan *users.UserResponse, responseCode int) {
	duration := time.Duration(rand.Intn(1000)) * time.Millisecond
	timer := time.NewTimer(duration)
	<-timer.C
	UserResponsePtr := &users.UserResponse{
		Result: responseCode,
	}
	channel <- UserResponsePtr
}

func createDummyUsersRequesterFunctor(responseCodeReturned int, errsReturned []error, closeChannel bool) (UsersRequester, chan userRequesterCall) {
	var callsChannel chan userRequesterCall
	requester := func(issuerId string, certifierId string, request []byte) (chan *users.UserResponse, []error) {
		go (func() {
			callsChannel <- userRequesterCall{
				issuerId:    issuerId,
				certifierId: certifierId,
				request:     request,
			}
		})()
		if errsReturned != nil {
			return nil, errsReturned
		}
		responseChannel := make(chan *users.UserResponse, 1)
		if closeChannel {
			close(responseChannel)
		} else {
			go sendUserResponseAfterRandomDelay(responseChannel, responseCodeReturned)
		}
		return responseChannel, nil
	}
	return requester, callsChannel
}

func createDummyTicketGeneratorFunctor() TicketGenerator {
	ticketNum := 0
	lock := &sync.Mutex{}
	generator := func() int {
		lock.Lock()
		ticketCopy := ticketNum
		ticketNum += 1
		lock.Unlock()
		return ticketCopy
	}
	return generator
}

type dummyStatusEntry struct {
	status        int
	failureReason int
	result        []byte
	errors        []error
}

type dummyStatusRegistry struct {
	ticketLogs map[int][]dummyStatusEntry
	lock       *sync.Mutex
}

type userRequesterCall struct {
	issuerId    string
	certifierId string
	request     []byte
}

var responseReporterError error = errors.New("Response reporter error")

func createDummyResposeReporterFunctor(success bool) (ResponseReporter, *dummyStatusRegistry) {
	reg := dummyStatusRegistry{
		ticketLogs: map[int][]dummyStatusEntry{},
		lock:       &sync.Mutex{},
	}
	reporter := func(ticketNb int, status int, failureReason int, result []byte, errs []error) error {
		if !success {
			return responseReporterError
		}
		reg.lock.Lock()
		reg.ticketLogs[ticketNb] = append(reg.ticketLogs[ticketNb], dummyStatusEntry{
			status:        status,
			failureReason: failureReason,
			result:        result,
			errors:        errs,
		})
		reg.lock.Unlock()
		return nil
	}
	return reporter, &reg
}

/*
	General tests
*/

func TestStartShutdownServer(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(1, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2, nil, false)
	responseReporter, _ := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}
	ShutdownServer()
}

func TestInvalidRequestType(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(1, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2, nil, false)
	responseReporter, _ := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}

	_, err := MakeRequest(false, UsersRequest-1, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != invalidRequestTypeError {
		t.Error("Request with invalid type should be rejected.")
	}

	ShutdownServer()
}

func TestReponseReporterQueueError(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(1, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2, nil, false)
	responseReporter, reg := createDummyResposeReporterFunctor(false)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}

	ticketNb, err := MakeRequest(false, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != responseReporterError {
		t.Error("Request should fail with response reporter error while queueing.")
	}

	if len(reg.ticketLogs[ticketNb]) != 0 {
		t.Error("Status for ticket number should be empty if queueing failed.")
	}

	ShutdownServer()
}

func TestRequestWhileNotRunning(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(1, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2, nil, false)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}

	ShutdownServer()

	ticketNb, err := MakeRequest(false, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err == nil {
		t.Error("Request should fail if made while server is down.")
	}

	if len(reg.ticketLogs[ticketNb]) != 2 ||
		reg.ticketLogs[ticketNb][0].status != QueuedStatus ||
		reg.ticketLogs[ticketNb][1].status != FailedStatus ||
		reg.ticketLogs[ticketNb][1].failureReason != RejectedReason {
		t.Error("Status for ticket number should be updated if failing when server is down.")
	}
}

func TestVerifiedUserRequest(t *testing.T) {
	// Set up context needed
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2, nil, false)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()

	// Test with request rejected directly from users requester
	requestError := errors.New("Request Failed.")
	usersRequesterFailing, _ := createDummyUsersRequesterFunctor(1, []error{requestError}, false)
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterFailing, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}

	ticketNb, err := MakeRequest(true, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketNb]) != 3 ||
		reg.ticketLogs[ticketNb][0].status != QueuedStatus ||
		reg.ticketLogs[ticketNb][1].status != RunningStatus ||
		reg.ticketLogs[ticketNb][2].status != FailedStatus ||
		reg.ticketLogs[ticketNb][2].failureReason != RejectedReason ||
		!reflect.DeepEqual(reg.ticketLogs[ticketNb][2].errors, []error{requestError}) {
		t.Error("Request should run but fail, and statuses should be reported correctly.")
	}

	// Test with channel closed from users requester
	usersRequesterSuccess, _ := createDummyUsersRequesterFunctor(3, nil, true)
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequesterSuccess, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}
	ticketNb, err = MakeRequest(true, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketNb]) != 3 ||
		reg.ticketLogs[ticketNb][0].status != QueuedStatus ||
		reg.ticketLogs[ticketNb][1].status != RunningStatus ||
		reg.ticketLogs[ticketNb][2].status != FailedStatus ||
		reg.ticketLogs[ticketNb][2].failureReason != RejectedReason ||
		!reflect.DeepEqual(reg.ticketLogs[ticketNb][2].errors, []error{subsystemChannelClosed}) {
		t.Error("Request should run but fail, and statuses should be reported correctly.")
	}


}
