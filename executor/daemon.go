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
func InitializeServer(
	usersRequester UsersRequester,
	usersRequesterUnverified UsersRequester,
	responseReporter ResponseReporter,
	ticketGenerator TicketGenerator,
) {
	serverSingleton.usersRequester = usersRequester
	serverSingleton.usersRequesterUnverified = usersRequesterUnverified
	serverSingleton.responseReporter = responseReporter
	serverSingleton.ticketGenerator = ticketGenerator
	gofarm.InitServer(&serverSingleton)
}

func StartServer(conf Config) error {
	return gofarm.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownServer() {
	gofarm.ShutdownServer()
}

func (sv *server) reportFailure(ticketNb int, reason int, errs []error) {
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
	_, err = gofarm.MakeRequest(&executorRequest{
		isVerified:  isVerified,
		requestType: requestType,
		issuerId:    issuerId,
		certifierId: certifierId,
		ticket:      ticketNb,
		request:     request,
	})
	if err != nil {
		serverSingleton.reportFailure(ticketNb, RejectedReason, []error{err})
		return ticketNb, err
	}

	return ticketNb, nil
}

/*
	Server implementation
*/

type server struct {
	// Requester lambdas
	usersRequester           UsersRequester
	usersRequesterUnverified UsersRequester
	responseReporter         ResponseReporter
	ticketGenerator          TicketGenerator
}

var serverSingleton server

func (sv *server) Start(_ gofarm.Config, _ bool) error { return nil }

func (sv *server) Shutdown() error { return nil }

func (sv *server) Work(nativeRequest *gofarm.Request) *gofarm.Response {
	wrappedRequest := (*nativeRequest).(*executorRequest)

	switch wrappedRequest.requestType {
	case UsersRequest:
		// @TODO: Replace with status codes
		sv.responseReporter(wrappedRequest.ticket, RunningStatus, NoReason, nil, nil)

		// Determine lambda to use based on whether the request is verified or not
		var usersRequester UsersRequester
		if wrappedRequest.isVerified {
			usersRequester = sv.usersRequester
		} else {
			usersRequester = sv.usersRequesterUnverified
		}

		channel, errs := usersRequester(wrappedRequest.issuerId, wrappedRequest.certifierId, wrappedRequest.request)
		if errs != nil {
			sv.reportFailure(wrappedRequest.ticket, RejectedReason, errs)
			break
		}
		userResponsePtr := <-channel
		userReponseEncoded, _ := userResponsePtr.Encode()
		sv.responseReporter(wrappedRequest.ticket, SuccessStatus, NoReason, userReponseEncoded, nil)
	}

	return nil
}
