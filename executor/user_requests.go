package executor

import (
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
)

/*
	User request
*/

func (sv *server) doGenericUsersRequest(wrappedRequest *executorRequest) {
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
}
