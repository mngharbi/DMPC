/*
	Test helpers
*/

package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/locker"
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
	genericChannelId   string = "CHANNEL_ID"
	genericKeyId       string = "KEY_ID"
	genericIssuerId    string = "ISSUER_ID"
	genericCertifierId string = "CERTIFIER_ID"
)

var (
	genericKey []byte = generateRandomBytes(core.SymmetricKeySize)
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
	signers    *core.VerifiedSigners
	readLock   bool
	readUnlock bool
	request    []byte
}

var (
	userObjectsWithPermissions []users.UserObject = []users.UserObject{
		{
			Permissions: users.PermissionsObject{
				Channel: users.ChannelPermissionsObject{
					Add: true,
				},
			},
		},
	}
)

func sendUserResponseAfterRandomDelay(channel chan *users.UserResponse, responseCode int, data []users.UserObject) {
	waitForRandomDuration()
	UserResponsePtr := &users.UserResponse{
		Result: responseCode,
		Data:   data,
	}
	channel <- UserResponsePtr
}

func createDummyUsersRequesterFunctor(responseCodeReturned int, data []users.UserObject, errsReturned []error, closeChannel bool) (users.Requester, chan userRequesterCall) {
	callsChannel := make(chan userRequesterCall, 0)
	requester := func(signers *core.VerifiedSigners, readLock bool, readUnlock bool, request []byte) (chan *users.UserResponse, []error) {
		go (func() {
			callsChannel <- userRequesterCall{
				signers:    signers,
				readLock:   readLock,
				readUnlock: readUnlock,
				request:    request,
			}
		})()
		if errsReturned != nil {
			return nil, errsReturned
		}
		responseChannel := make(chan *users.UserResponse)
		if closeChannel {
			close(responseChannel)
		} else {
			go sendUserResponseAfterRandomDelay(responseChannel, responseCodeReturned, data)
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
	Channels dummies
*/

func sendChannelsResponseAfterRandomDelay(channel chan *channels.ChannelsResponse, responseCode channels.ChannelsStatusCode) {
	waitForRandomDuration()
	channelsReponsePtr := &channels.ChannelsResponse{
		Result: responseCode,
	}
	channel <- channelsReponsePtr
}

func createDummyChannelActionFunctor(responseCodeReturned channels.ChannelsStatusCode, errReturned error, closeChannel bool) (channels.ChannelActionRequester, chan interface{}) {
	callsChannel := make(chan interface{}, 0)
	requester := func(request interface{}) (chan *channels.ChannelsResponse, error) {
		go (func() {
			callsChannel <- request
		})()
		if errReturned != nil {
			return nil, errReturned
		}
		responseChannel := make(chan *channels.ChannelsResponse)
		if closeChannel {
			close(responseChannel)
		} else {
			go sendChannelsResponseAfterRandomDelay(responseChannel, responseCodeReturned)
		}
		return responseChannel, nil
	}
	return requester, callsChannel
}

/*
	Locker dummies
*/

func sendLockerResponseAfterRandomDelay(channel chan bool, response bool) {
	waitForRandomDuration()
	channel <- response
}

func createDummyLockerFunctor(responseReturned bool, errsReturned []error, closeChannel bool) (locker.Requester, chan interface{}) {
	callsChannel := make(chan interface{}, 0)
	requester := func(request *locker.LockerRequest) (chan bool, []error) {
		go (func() {
			callsChannel <- request
		})()
		if errsReturned != nil {
			return nil, errsReturned
		}
		responseChannel := make(chan bool)
		if closeChannel {
			close(responseChannel)
		} else {
			go sendLockerResponseAfterRandomDelay(responseChannel, responseReturned)
		}
		return responseChannel, nil
	}
	return requester, callsChannel
}

/*
	Key adder dummies
*/

type keyAdderCall struct {
	keyId string
	key   []byte
}

func createDummyKeyAdderFunctor(response error) (core.KeyAdder, chan interface{}) {
	callsChannel := make(chan interface{}, 0)
	requester := func(keyId string, key []byte) error {
		go (func() {
			callsChannel <- keyAdderCall{
				keyId: keyId,
				key:   key,
			}
		})()

		return response
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
	channelActionRequester channels.ChannelActionRequester,
	lockerRequester locker.Requester,
	keyAdder core.KeyAdder,
	responseReporter status.Reporter,
	ticketGenerator status.TicketGenerator,
) bool {
	serverSingleton = server{}
	InitializeServer(usersRequester, usersRequesterUnverified, messageAdder, operationBufferer, channelActionRequester, lockerRequester, keyAdder, responseReporter, ticketGenerator, log, shutdownProgram)
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

/*
	Utilities
*/

func generateRandomBytes(nbBytes int) (bytes []byte) {
	bytes = make([]byte, nbBytes)
	rand.Read(bytes)
	return
}
