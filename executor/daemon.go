package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/gofarm"
)

/*
	Errors
*/

var (
	invalidRequestTypeError error = errors.New("Invalid request type.")
	subsystemChannelClosed  error = errors.New("Corresponding subsystem shutdown during the request.")
	requestRejectedError    error = errors.New("Corresponding subsystem rejected the request.")
)

/*
	Server implementation
*/

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
		sv.doGenericUsersRequest(wrappedRequest)
	case core.ReadChannelType:
		sv.doReadChannel(wrappedRequest)
	case core.AddChannelType:
		sv.doAddChannel(wrappedRequest)
	case core.CloseChannelType:
		sv.doCloseChannel(wrappedRequest)
	case core.AddMessageType:
		sv.doAddMessage(wrappedRequest)
	}

	return
}

func (sv *server) reportRejection(ticketId status.Ticket, reason status.FailReasonCode, errs []error) {
	sv.responseReporter(ticketId, status.FailedStatus, reason, nil, errs)
}
