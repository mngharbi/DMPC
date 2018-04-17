package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

/*
	Dummy subsystem lambdas
*/

func waitForRandomDuration() {
	duration := time.Duration(rand.Intn(100)) * time.Millisecond
	timer := time.NewTimer(duration)
	<-timer.C
}

func sendUserResponseAfterRandomDelay(channel chan *users.UserResponse, responseCode int) {
	waitForRandomDuration()
	UserResponsePtr := &users.UserResponse{
		Result: responseCode,
	}
	channel <- UserResponsePtr
}

func createDummyUsersRequesterFunctor(responseCodeReturned int, errsReturned []error, closeChannel bool) (UsersRequester, chan userRequesterCall) {
	callsChannel := make(chan userRequesterCall, 0)
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
		responseChannel := make(chan *users.UserResponse)
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
	lock := &sync.Mutex{}
	generator := func() string {
		lock.Lock()
		ticket := core.GenerateUniqueId()
		lock.Unlock()
		return ticket
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
	ticketLogs map[string][]dummyStatusEntry
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
		ticketLogs: map[string][]dummyStatusEntry{},
		lock:       &sync.Mutex{},
	}
	reporter := func(ticketId string, status int, failureReason int, result []byte, errs []error) error {
		if !success {
			return responseReporterError
		}
		reg.lock.Lock()
		reg.ticketLogs[ticketId] = append(reg.ticketLogs[ticketId], dummyStatusEntry{
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
	usersRequester, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	responseReporter, _ := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}
	ShutdownServer()
}

func TestInvalidRequestType(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
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
	usersRequester, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	responseReporter, reg := createDummyResposeReporterFunctor(false)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(false, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != responseReporterError {
		t.Error("Request should fail with response reporter error while queueing.")
	}

	if len(reg.ticketLogs[ticketId]) != 0 {
		t.Error("Status for ticket number should be empty if queueing failed.")
	}

	ShutdownServer()
}

func TestRequestWhileNotRunning(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}

	ShutdownServer()

	ticketId, err := MakeRequest(false, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err == nil {
		t.Error("Request should fail if made while server is down.")
	}

	if len(reg.ticketLogs[ticketId]) != 2 ||
		reg.ticketLogs[ticketId][0].status != QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != FailedStatus ||
		reg.ticketLogs[ticketId][1].failureReason != RejectedReason {
		t.Error("Status for ticket number should be updated if failing when server is down.")
	}
}

func doUserRequestTesting(t *testing.T, isVerified bool) {
	// Set up context needed
	usersRequesterDummy, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	responseReporter, reg := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	var usersRequester, usersRequesterVerified UsersRequester

	// Test with request rejected directly from users requester
	requestError := errors.New("Request Failed.")
	usersRequesterFailing, _ := createDummyUsersRequesterFunctor(users.Success, []error{requestError}, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterFailing, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterFailing, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, responseReporter, ticketGenerator) {
		return
	}

	ticketId, err := MakeRequest(isVerified, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != RunningStatus ||
		reg.ticketLogs[ticketId][2].status != FailedStatus ||
		reg.ticketLogs[ticketId][2].failureReason != RejectedReason ||
		!reflect.DeepEqual(reg.ticketLogs[ticketId][2].errors, []error{requestError}) {
		t.Error("Request should run but fail, and statuses should be reported correctly when request is rejected.")
	}

	// Test with channel closed from users requester
	usersRequesterSuccess, _ := createDummyUsersRequesterFunctor(users.Success, nil, true)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterSuccess, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterSuccess, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, responseReporter, ticketGenerator) {
		return
	}
	ticketId, err = MakeRequest(isVerified, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != RunningStatus ||
		reg.ticketLogs[ticketId][2].status != FailedStatus ||
		reg.ticketLogs[ticketId][2].failureReason != RejectedReason ||
		!reflect.DeepEqual(reg.ticketLogs[ticketId][2].errors, []error{subsystemChannelClosed}) {
		t.Error("Request should run but fail, and statuses should be reported correctly when channel closes.")
	}

	// Test with failed requests
	usersRequesterUnsuccessfulResponse, _ := createDummyUsersRequesterFunctor(1+users.Success, nil, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterUnsuccessfulResponse, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterUnsuccessfulResponse, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, responseReporter, ticketGenerator) {
		return
	}
	ticketId, err = MakeRequest(isVerified, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != RunningStatus ||
		reg.ticketLogs[ticketId][2].status != FailedStatus ||
		reg.ticketLogs[ticketId][2].failureReason != FailedReason ||
		reg.ticketLogs[ticketId][2].errors != nil {
		t.Error("Request should run but fail, and statuses should be reported correctly when the request failed.")
	}

	// Test with one successful request
	usersRequesterSuccessfulResponse, _ := createDummyUsersRequesterFunctor(users.Success, nil, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterSuccessfulResponse, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterSuccessfulResponse, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, responseReporter, ticketGenerator) {
		return
	}
	ticketId, err = MakeRequest(isVerified, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", []byte{})
	if err != nil {
		t.Error("Request should not fail.")
		return
	}

	ShutdownServer()

	if len(reg.ticketLogs[ticketId]) != 3 ||
		reg.ticketLogs[ticketId][0].status != QueuedStatus ||
		reg.ticketLogs[ticketId][1].status != RunningStatus ||
		reg.ticketLogs[ticketId][2].status != SuccessStatus {
		t.Error("Request should succeed.")
	}

	// Test with concurrent successful requests and check calls made to users subsystem
	usersRequesterSuccessfulResponseMultiple, callsChannel := createDummyUsersRequesterFunctor(users.Success, nil, false)
	if isVerified {
		usersRequester, usersRequesterVerified = usersRequesterSuccessfulResponseMultiple, usersRequesterDummy
	} else {
		usersRequesterVerified, usersRequester = usersRequesterSuccessfulResponseMultiple, usersRequesterDummy
	}
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterVerified, responseReporter, ticketGenerator) {
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
			_, err = MakeRequest(isVerified, UsersRequest, "ISSUER_ID", "CERTIFIER_ID", payload)
			wg.Done()
		})()
	}
	wg.Wait()

	ShutdownServer()

	checksum := 0
	for i := 1; i <= 10; i++ {
		callLog := <-callsChannel
		if callLog.issuerId != "ISSUER_ID" ||
			callLog.certifierId != "CERTIFIER_ID" {
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
