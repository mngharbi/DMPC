package executor

import (
	"errors"
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
	Errors
*/

var (
	invalidRequestTypeError          error = errors.New("Invalid request type.")
	subsystemChannelClosed           error = errors.New("Corresponding subsystem shutdown during the request.")
	requestRejectedError             error = errors.New("Corresponding subsystem rejected the request.")
	unverifiedChannelCreationError   error = errors.New("Channel creation cannot be unverified.")
	channelCreationUnauthorizedError error = errors.New("Channel creation is not authorized.")
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
	channelActionRequester channels.ChannelActionRequester,
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

func (sv *server) reportRejection(ticketId status.Ticket, reason status.FailReasonCode, errs []error) {
	sv.responseReporter(ticketId, status.FailedStatus, reason, nil, errs)
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
	channelActionRequester   channels.ChannelActionRequester
	lockerRequester          locker.Requester
	keyAdder                 core.KeyAdder
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

	// Report running status
	sv.responseReporter(wrappedRequest.ticket, status.RunningStatus, status.NoReason, nil, nil)

	switch wrappedRequest.metaFields.RequestType {
	case core.UsersRequestType:
		// Determine lambda to use based on whether the request is verified or not
		var usersRequester users.Requester
		if wrappedRequest.isVerified {
			usersRequester = sv.usersRequester
		} else {
			usersRequester = sv.usersRequesterUnverified
		}

		// Make the request to users subsystem (not leaving it locked)
		channel, errs := usersRequester(wrappedRequest.signers, true, true, wrappedRequest.request)
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
	case core.AddChannelType:
		// Parse request
		request := &channels.OpenChannelRequest{}
		err := request.Decode(wrappedRequest.request)
		if err != nil {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{err})
			return
		}

		// Get and RLock certifier user object
		if wrappedRequest.signers == nil {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{unverifiedChannelCreationError})
			return
		}
		usersRequest := &users.UserRequest{
			Type:      users.ReadRequest,
			Timestamp: wrappedRequest.metaFields.Timestamp,
			Fields:    []string{wrappedRequest.signers.CertifierId},
		}
		encodedUsersRequest, _ := usersRequest.Encode()
		usersSubsystemResponse, errs := sv.usersRequesterUnverified(nil, true, false, encodedUsersRequest)
		if len(errs) != 0 {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{requestRejectedError})
			return
		}
		userResponsePtr, ok := <-usersSubsystemResponse
		if !ok {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{subsystemChannelClosed})
			return
		}
		if userResponsePtr.Result != users.Success {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{requestRejectedError})
			return
		}

		// RUnlock at the end
		defer func() {
			usersSubsystemResponse, _ = sv.usersRequesterUnverified(nil, false, true, encodedUsersRequest)
			_ = <-usersSubsystemResponse
		}()

		certifierCheckSuccess := len(userResponsePtr.Data) == 1 && userResponsePtr.Data[0].Permissions.Channel.Add
		if !certifierCheckSuccess {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{channelCreationUnauthorizedError})
			return
		}

		// Lock channel
		lockRequest := &locker.LockerRequest{
			Type: locker.ChannelLock,
			Needs: []core.LockNeed{
				{
					LockType: core.WriteLockType,
					Id:       request.Id,
				},
			},
		}
		lockRequest.LockingType = core.Locking
		lockChannel, errs := sv.lockerRequester(lockRequest)
		if len(errs) != 0 {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, errs)
			return
		}
		defer func() {
			lockRequest.LockingType = core.Unlocking
			lockChannel, _ = sv.lockerRequester(lockRequest)
			_ = <-lockChannel
		}()

		// Add key to keys subsystems
		if keyAddError := sv.keyAdder(request.KeyId, request.Key); keyAddError != nil {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{keyAddError})
			return
		}

		// Send request through to channels subsystem
		channelResponseChannel, err := sv.channelActionRequester(request)
		if err != nil {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{requestRejectedError})
			return
		}
		channelResponsePtr, ok := <-channelResponseChannel
		if !ok {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{subsystemChannelClosed})
		} else {
			channelResponseEncoded, _ := channelResponsePtr.Encode()
			sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, channelResponseEncoded, nil)
		}

	case core.AddMessageType:
		// Read Lock channel
		lockRequest := &locker.LockerRequest{
			Type: locker.ChannelLock,
			Needs: []core.LockNeed{
				{
					LockType: core.ReadLockType,
					Id:       wrappedRequest.metaFields.ChannelId,
				},
			},
		}
		lockRequest.LockingType = core.Locking
		lockChannel, errs := sv.lockerRequester(lockRequest)
		if len(errs) != 0 {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, errs)
			return
		}
		defer func() {
			lockRequest.LockingType = core.Unlocking
			lockChannel, _ = sv.lockerRequester(lockRequest)
			_ = <-lockChannel
		}()

		// Send request to channels subsystem based on type (operation buffering/ add message)
		var messageChannel chan *channels.MessagesResponse
		var requestErr error
		if wrappedRequest.failedOperation == nil {
			messageChannel, requestErr = sv.messageAdder(&channels.AddMessageRequest{
				Timestamp: wrappedRequest.metaFields.Timestamp,
				Signers:   wrappedRequest.signers,
				Message:   wrappedRequest.request,
			})
		} else {
			messageChannel, requestErr = sv.operationBufferer(&channels.BufferOperationRequest{
				Operation: wrappedRequest.failedOperation,
			})
		}

		// Handle request rejection
		if requestErr != nil {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{requestErr})
			return
		}

		// Wait for response and handle premature channel closure
		messageResponse, ok := <-messageChannel
		if !ok {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{subsystemChannelClosed})
			return
		}

		// Handle response
		if messageResponse.Result != channels.MessagesSuccess {
			sv.responseReporter(wrappedRequest.ticket, status.FailedStatus, status.FailedReason, nil, nil)
		} else {
			sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, nil, nil)
		}
	}

	return
}
