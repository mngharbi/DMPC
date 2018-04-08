package executor

import (
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

func createDummyResposeReporterFunctor() (ResponseReporter, *dummyLoggerRegistry) {
	reg := dummyLoggerRegistry{
		ticketLogs: map[int][]dummyLoggerEntry{},
		lock:       &sync.Mutex{},
	}
	reporter := func(ticketNb int, status int, result []byte, errs []error) error {
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

func TestStartShutdownWorker(t *testing.T) {
	usersRequester, _ := createDummyUsersRequesterFunctor(1)
	usersRequesterUnverified, _ := createDummyUsersRequesterFunctor(2)
	responseReporter, _ := createDummyResposeReporterFunctor()
	ticketGenerator := createDummyTicketGeneratorFunctor()
	if !resetAndStartServer(t, multipleWorkersConfig(), usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator) {
		return
	}
	ShutdownServer()
}
