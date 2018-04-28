package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"github.com/mngharbi/gofarm"
)

/*
	Errors
*/

var invalidRequestTypeError error = errors.New("Invalid request type.")
var subsystemChannelClosed error = errors.New("Corresponding subsystem shutdown during the request.")

/*
	Daemon configuration
*/

type Config struct {
	NumWorkers int
}

/*
	Types of lambdas to call other subsystems
*/
type UsersRequester func(string, string, []byte) (chan *users.UserResponse, []error)
type ResponseReporter func(status.Ticket, status.StatusCode, status.FailReasonCode, []byte, []error) error
type TicketGenerator func() status.Ticket

/*
	Logging
*/
var (
	log *core.LoggingHandler
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
	usersRequester UsersRequester,
	usersRequesterUnverified UsersRequester,
	responseReporter ResponseReporter,
	ticketGenerator TicketGenerator,
	loggingHandler *core.LoggingHandler,
) {
	provisionServerOnce()
	log = loggingHandler
	serverSingleton.usersRequester = usersRequester
	serverSingleton.usersRequesterUnverified = usersRequesterUnverified
	serverSingleton.responseReporter = responseReporter
	serverSingleton.ticketGenerator = ticketGenerator
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
	requestType int,
	issuerId string,
	certifierId string,
	request []byte,
) (status.Ticket, error) {
	// Check type
	if !isValidRequestType(requestType) {
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
		isVerified:  isVerified,
		requestType: requestType,
		issuerId:    issuerId,
		certifierId: certifierId,
		ticket:      ticketId,
		request:     request,
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
	usersRequester           UsersRequester
	usersRequesterUnverified UsersRequester
	responseReporter         ResponseReporter
	ticketGenerator          TicketGenerator
}

func (sv *server) Start(_ gofarm.Config, _ bool) error { return nil }

func (sv *server) Shutdown() error { return nil }

func (sv *server) Work(nativeRequest *gofarm.Request) (dummyResponsePtr *gofarm.Response) {
	dummyResponsePtr = nil

	wrappedRequest := (*nativeRequest).(*executorRequest)

	switch wrappedRequest.requestType {
	case UsersRequest:
		sv.responseReporter(wrappedRequest.ticket, status.RunningStatus, status.NoReason, nil, nil)

		// Determine lambda to use based on whether the request is verified or not
		var usersRequester UsersRequester
		if wrappedRequest.isVerified {
			usersRequester = sv.usersRequester
		} else {
			usersRequester = sv.usersRequesterUnverified
		}

		// Make the request to users subsytem
		channel, errs := usersRequester(wrappedRequest.issuerId, wrappedRequest.certifierId, wrappedRequest.request)
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
