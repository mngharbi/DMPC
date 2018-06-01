/*
	Test helpers
*/

package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"math/rand"
	"sync"
	"testing"
	"time"
)

/*
	General
*/

const (
	genericKeyId       string = "KEY_ID"
	genericIssuerId    string = "ISSUER_ID"
	genericCertifierId string = "CERTIFIER_ID"
)

func generateSigners(issuerId string, certifierId string) *core.VerifiedSigners {
	return &core.VerifiedSigners{
		IssuerId:    issuerId,
		CertifierId: certifierId,
	}
}

func generateGenericSigners() *core.VerifiedSigners {
	return generateSigners(genericIssuerId, genericCertifierId)
}

func waitForRandomDuration() {
	duration := time.Duration(rand.Intn(100)) * time.Millisecond
	timer := time.NewTimer(duration)
	<-timer.C
}

/*
	User dummies
*/

type userRequesterCall struct {
	signers *core.VerifiedSigners
	request []byte
}

func sendUserResponseAfterRandomDelay(channel chan *users.UserResponse, responseCode int) {
	waitForRandomDuration()
	UserResponsePtr := &users.UserResponse{
		Result: responseCode,
	}
	channel <- UserResponsePtr
}

func createDummyUsersRequesterFunctor(responseCodeReturned int, errsReturned []error, closeChannel bool) (users.Requester, chan userRequesterCall) {
	callsChannel := make(chan userRequesterCall, 0)
	requester := func(signers *core.VerifiedSigners, request []byte) (chan *users.UserResponse, []error) {
		go (func() {
			callsChannel <- userRequesterCall{
				signers: signers,
				request: request,
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

/*
	Ticket dummy
*/

func createDummyTicketGeneratorFunctor() status.TicketGenerator {
	lock := &sync.Mutex{}
	generator := func() status.Ticket {
		lock.Lock()
		ticket := status.RequestNewTicket()
		lock.Unlock()
		return ticket
	}
	return generator
}

/*
	Status dummies
*/
type dummyStatusEntry struct {
	status        status.StatusCode
	failureReason status.FailReasonCode
	result        []byte
	errors        []error
}

type dummyStatusRegistry struct {
	ticketLogs map[status.Ticket][]dummyStatusEntry
	lock       *sync.Mutex
}

var responseReporterError error = errors.New("Response reporter error")

func createDummyResposeReporterFunctor(success bool) (status.Reporter, *dummyStatusRegistry) {
	reg := dummyStatusRegistry{
		ticketLogs: map[status.Ticket][]dummyStatusEntry{},
		lock:       &sync.Mutex{},
	}
	reporter := func(ticketId status.Ticket, status status.StatusCode, failureReason status.FailReasonCode, result []byte, errs []error) error {
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
	Messages dummies
*/

func sendMessageResponseAfterRandomDelay(channel chan *channels.MessagesResponse, responseCode channels.MessagesStatusCode) {
	waitForRandomDuration()
	messageReponsePtr := &channels.MessagesResponse{
		Result: responseCode,
	}
	channel <- messageReponsePtr
}

func createDummyMessageAdderFunctor(responseCodeReturned channels.MessagesStatusCode, errReturned error, closeChannel bool) (channels.MessageAdder, chan interface{}) {
	callsChannel := make(chan interface{}, 0)
	requester := func(addMessageRequest *channels.AddMessageRequest) (chan *channels.MessagesResponse, error) {
		go (func() {
			callsChannel <- addMessageRequest
		})()
		if errReturned != nil {
			return nil, errReturned
		}
		responseChannel := make(chan *channels.MessagesResponse)
		if closeChannel {
			close(responseChannel)
		} else {
			go sendMessageResponseAfterRandomDelay(responseChannel, responseCodeReturned)
		}
		return responseChannel, nil
	}
	return requester, callsChannel
}

/*
	Operation bufferer
*/

func createDummyOperationBuffererFunctor(responseCodeReturned channels.MessagesStatusCode, errReturned error, closeChannel bool) (channels.OperationBufferer, chan interface{}) {
	callsChannel := make(chan interface{}, 0)
	requester := func(bufferOperationRequest *channels.BufferOperationRequest) (chan *channels.MessagesResponse, error) {
		go (func() {
			callsChannel <- bufferOperationRequest
		})()
		if errReturned != nil {
			return nil, errReturned
		}
		responseChannel := make(chan *channels.MessagesResponse)
		if closeChannel {
			close(responseChannel)
		} else {
			go sendMessageResponseAfterRandomDelay(responseChannel, responseCodeReturned)
		}
		return responseChannel, nil
	}
	return requester, callsChannel
}

/*
	Server
*/

func resetAndStartServer(
	t *testing.T,
	conf Config,
	usersRequester users.Requester,
	usersRequesterUnverified users.Requester,
	messageAdder channels.MessageAdder,
	operationBufferer channels.OperationBufferer,
	responseReporter status.Reporter,
	ticketGenerator status.TicketGenerator,
) bool {
	serverSingleton = server{}
	InitializeServer(usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, responseReporter, ticketGenerator, log, shutdownProgram)
	err := StartServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func multipleWorkersConfig() Config {
	return Config{
		NumWorkers: 6,
	}
}
