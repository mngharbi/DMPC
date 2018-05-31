package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"github.com/mngharbi/gofarm"
)

/*
	Function to send in a decrypted request into the executor and returns a ticket
*/
type Requester func(bool, *core.OperationMetaFields, string, *core.VerifiedSigners, []byte, *core.Operation) (status.Ticket, error)

/*
	Errors
*/

var (
	invalidRequestTypeError error = errors.New("Invalid request type.")
	subsystemChannelClosed  error = errors.New("Corresponding subsystem shutdown during the request.")
	requestRejectedError    error = errors.New("Corresponding subsystem rejected the request.")
)

/*
	Daemon configuration
*/

type Config struct {
	NumWorkers int
}

/*
	Logging
*/
var (
	log             *core.LoggingHandler
	shutdownProgram core.ShutdownLambda
)

/*
	Server API
*/

func provisionServerOnce() {
	if serverHandler == nil {
		serverHandler = gofarm.ProvisionServer()
	}
}

func InitializeServer(
	usersRequester users.Requester,
	usersRequesterUnverified users.Requester,
	messageAdder channels.MessageAdder,
	operationBufferer channels.OperationBufferer,
	responseReporter status.Reporter,
	ticketGenerator status.TicketGenerator,
	loggingHandler *core.LoggingHandler,
	shutdownLambda core.ShutdownLambda,
) {
	provisionServerOnce()
	serverSingleton.usersRequester = usersRequester
	serverSingleton.usersRequesterUnverified = usersRequesterUnverified
	serverSingleton.messageAdder = messageAdder
	serverSingleton.operationBufferer = operationBufferer
	serverSingleton.responseReporter = responseReporter
	serverSingleton.ticketGenerator = ticketGenerator
	log = loggingHandler
	shutdownProgram = shutdownLambda
	serverHandler.InitServer(&serverSingleton)
}

func StartServer(conf Config) error {
	provisionServerOnce()
	return serverHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownServer() {
	provisionServerOnce()
	serverHandler.ShutdownServer()
}

func (sv *server) reportRejection(ticketId status.Ticket, reason status.FailReasonCode, errs []error) {
	sv.responseReporter(ticketId, status.FailedStatus, reason, nil, errs)
}

func MakeRequest(
	isVerified bool,
	metaFields *core.OperationMetaFields,
	keyId string,
	signers *core.VerifiedSigners,
	request []byte,
	failedOperation *core.Operation,
) (status.Ticket, error) {
	log.Debugf(receivedRequestLogMsg)

	// Check type
	if !isValidRequestType(metaFields.RequestType) {
		return "", invalidRequestTypeError
	}

	// Generate ticket
	ticketId := serverSingleton.ticketGenerator()
	err := serverSingleton.responseReporter(ticketId, status.QueuedStatus, status.NoReason, nil, nil)
	if err != nil {
		return ticketId, err
	}

	// Make request
	_, err = serverHandler.MakeRequest(&executorRequest{
		isVerified:      isVerified,
		metaFields:      metaFields,
		keyId:           keyId,
		signers:         signers,
		ticket:          ticketId,
		request:         request,
		failedOperation: failedOperation,
	})
	if err != nil {
		serverSingleton.reportRejection(ticketId, status.RejectedReason, []error{err})
		return ticketId, err
	}

	return ticketId, nil
}

/*
	Server implementation
*/

var (
	serverSingleton server
	serverHandler   *gofarm.ServerHandler
)

type server struct {
	// Requester lambdas
	usersRequester           users.Requester
	usersRequesterUnverified users.Requester
	messageAdder             channels.MessageAdder
	operationBufferer        channels.OperationBufferer
	responseReporter         status.Reporter
	ticketGenerator          status.TicketGenerator
}

func (sv *server) Start(_ gofarm.Config, _ bool) error {
	log.Debugf(daemonStartLogMsg)
	return nil
}

func (sv *server) Shutdown() error {
	log.Debugf(daemonShutdownLogMsg)
	return nil
}

func (sv *server) Work(nativeRequest *gofarm.Request) (dummyResponsePtr *gofarm.Response) {
	log.Debugf(runningRequestLogMsg)
	dummyResponsePtr = nil

	wrappedRequest := (*nativeRequest).(*executorRequest)

	switch wrappedRequest.metaFields.RequestType {
	case core.UsersRequestType:
		sv.responseReporter(wrappedRequest.ticket, status.RunningStatus, status.NoReason, nil, nil)

		// Determine lambda to use based on whether the request is verified or not
		var usersRequester users.Requester
		if wrappedRequest.isVerified {
			usersRequester = sv.usersRequester
		} else {
			usersRequester = sv.usersRequesterUnverified
		}

		// Make the request to users subsystem
		channel, errs := usersRequester(wrappedRequest.signers, wrappedRequest.request)
		if errs != nil {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, errs)
			return
		}

		// Wait for response from users subsystem
		userResponsePtr, ok := <-channel
		if !ok {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{subsystemChannelClosed})
			return
		}

		// Handle failure after running the request
		userReponseEncoded, _ := userResponsePtr.Encode()
		if userResponsePtr.Result != users.Success {
			sv.responseReporter(wrappedRequest.ticket, status.FailedStatus, status.FailedReason, userReponseEncoded, nil)
		} else {
			sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, userReponseEncoded, nil)
		}
	}

	return
}
