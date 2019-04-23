package executor

import (
	"errors"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/locker"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
)

/*
	Errors
*/

var (
	unverifiedChannelCreationError   error = errors.New("Channel creation cannot be unverified.")
	channelCreationUnauthorizedError error = errors.New("Channel creation is not authorized.")
)

/*
	Add channel
*/

func (sv *server) doAddChannel(wrappedRequest *executorRequest) {
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
}

/*
	Add message
*/

func (sv *server) doAddMessage(wrappedRequest *executorRequest) {
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
