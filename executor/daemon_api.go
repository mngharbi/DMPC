package executor

import (
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/locker"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"github.com/mngharbi/gofarm"
)

/*
	Function to send in a decrypted request into the executor and returns a ticket
*/
type Requester func(bool, *core.OperationMetaFields, *core.VerifiedSigners, []byte, *core.Operation) (status.Ticket, error)

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

var (
	serverSingleton server
	serverHandler   *gofarm.ServerHandler
)

type server struct {
	// Requester lambdas
	usersRequester            users.Requester
	usersRequesterUnverified  users.Requester
	messageAdder              channels.MessageAdder
	operationBufferer         channels.OperationBufferer
	channelActionRequester    channels.ChannelActionRequester
	channelListenersRequester channels.ListenersRequester
	lockerRequester           locker.Requester
	keyAdder                  core.KeyAdder
	responseReporter          status.Reporter
	ticketGenerator           status.TicketGenerator
}

func InitializeServer(
	usersRequester users.Requester,
	usersRequesterUnverified users.Requester,
	messageAdder channels.MessageAdder,
	operationBufferer channels.OperationBufferer,
	channelActionRequester channels.ChannelActionRequester,
	channelListenersRequester channels.ListenersRequester,
	lockerRequester locker.Requester,
	keyAdder core.KeyAdder,
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
	serverSingleton.channelActionRequester = channelActionRequester
	serverSingleton.channelListenersRequester = channelListenersRequester
	serverSingleton.lockerRequester = lockerRequester
	serverSingleton.keyAdder = keyAdder
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

func provisionServerOnce() {
	if serverHandler == nil {
		serverHandler = gofarm.ProvisionServer()
	}
}

func MakeRequest(
	isVerified bool,
	metaFields *core.OperationMetaFields,
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
