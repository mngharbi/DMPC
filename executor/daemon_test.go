package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/users"
	"math/rand"
	"sync"
	"testing"
	"time"
)

/*
	Dummy subsystem lambdas
*/

func sendUserResponseAfterRandomDelay(channel chan *users.UserResponse, responseCode int) {
	timer := time.NewTimer(time.Duration(rand.Intn(1000)) * time.Millisecond)
	<-timer.C
	go (func() {
		UserResponsePtr := &users.UserResponse{
			Result: responseCode,
		}
		channel <- UserResponsePtr
	})()
}

func createDummyUsersRequesterFunctor(responseCodeReturned int) (UsersRequester, chan userRequesterCall) {
	var callsChannel chan userRequesterCall
	requester := func(issuerId string, certifierId string, request []byte) (chan *users.UserResponse, []error) {
		callsChannel <- userRequesterCall{
			issuerId:    issuerId,
			certifierId: certifierId,
			request:     request,
		}
		responseChannel := make(chan *users.UserResponse)
		sendUserResponseAfterRandomDelay(responseChannel, responseCodeReturned)
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

type dummyLoggerEntry struct {
	status int
	result []byte
	errors []error
}

type dummyLoggerRegistry struct {
	ticketLogs map[int][]dummyLoggerEntry
	lock       *sync.Mutex
}

type userRequesterCall struct {
	issuerId    string
	certifierId string
	request     []byte
}

var responseReporterError error = errors.New("Response reporter error")

func createDummyResposeReporterFunctor(success bool) (ResponseReporter, *dummyLoggerRegistry) {
	reg := dummyLoggerRegistry{
		ticketLogs: map[int][]dummyLoggerEntry{},
		lock:       &sync.Mutex{},
	}
	reporter := func(ticketNb int, status int, result []byte, errs []error) error {
		if !success {
			return responseReporterError
		}
		reg.lock.Lock()
		reg.ticketLogs[ticketNb] = append(reg.ticketLogs[ticketNb], dummyLoggerEntry{
			status: status,
			result: result,
			errors: errs,
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
	usersRequester, _ := createDummyUsersRequesterFunctor(1)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2)
	responseReporter, _ := createDummyResposeReporterFunctor(true)
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}
	ShutdownServer()
}

func TestInvalidRequestType(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(1)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2)
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
	usersRequester, _ := createDummyUsersRequesterFunctor(1)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2)
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
	usersRequester, _ := createDummyUsersRequesterFunctor(1)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2)
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

	if len(reg.ticketLogs[ticketNb]) != 2 || reg.ticketLogs[ticketNb][1].status != 3 {
		t.Error("Status for ticket number should be updated if failing when server is down.")
	}
}
