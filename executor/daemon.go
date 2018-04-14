package executor

import (
	"errors"
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
type ResponseReporter func(int, int, int, []byte, []error) error
type TicketGenerator func() int

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
) {
	provisionServerOnce()
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

func (sv *server) reportRejection(ticketNb int, reason int, errs []error) {
	sv.responseReporter(ticketNb, FailedStatus, reason, nil, errs)
}

func MakeRequest(
	isVerified bool,
	requestType int,
	issuerId string,
	certifierId string,
	request []byte,
) (int, error) {
	// Check type
	if !isValidRequestType(requestType) {
		return 0, invalidRequestTypeError
	}

	// Generate ticket
	ticketNb := serverSingleton.ticketGenerator()
	err := serverSingleton.responseReporter(ticketNb, QueuedStatus, NoReason, nil, nil)
	if err != nil {
		return ticketNb, err
	}

	// Make request
	_, err = serverHandler.MakeRequest(&executorRequest{
		isVerified:  isVerified,
		requestType: requestType,
		issuerId:    issuerId,
		certifierId: certifierId,
		ticket:      ticketNb,
		request:     request,
	})
	if err != nil {
		serverSingleton.reportRejection(ticketNb, RejectedReason, []error{err})
		return ticketNb, err
	}

	return ticketNb, nil
}

/*
	Server implementation
*/

var (
	serverSingleton server
	serverHandler *gofarm.ServerHandler
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
		sv.responseReporter(wrappedRequest.ticket, RunningStatus, NoReason, nil, nil)

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
			sv.reportRejection(wrappedRequest.ticket, RejectedReason, errs)
			return
		}

		// Wait for response from users subsystem
		userResponsePtr, ok := <-channel
		if !ok {
			sv.reportRejection(wrappedRequest.ticket, RejectedReason, []error{subsystemChannelClosed})
			return
		}

		// Handle failure after running the request
		userReponseEncoded, _ := userResponsePtr.Encode()
		if userResponsePtr.Result != users.Success {
			sv.responseReporter(wrappedRequest.ticket, FailedStatus, FailedReason, userReponseEncoded, nil)
		} else {
			sv.responseReporter(wrappedRequest.ticket, SuccessStatus, NoReason, userReponseEncoded, nil)
		}
	}

	return
}
