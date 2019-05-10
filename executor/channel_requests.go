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
	unverifiedChannelOpenError             error = errors.New("Channel open request cannot be unverified.")
	channelOpenUnauthorizedError           error = errors.New("Channel open request is not authorized.")
	channelOpenNilChannelError             error = errors.New("Channel open request must have channel object.")
	unverifiedChannelSubscribeError        error = errors.New("Channel subscribe request cannot be unverified.")
	channelReadUnauthorizedError           error = errors.New("Channel read request is not authorized.")
	channelEncryptUnauthorizedError        error = errors.New("Channel encrypt request is not authorized.")
	channelEncryptOperationFormatError     error = errors.New("Channel encrypt requires a valid operation as payload.")
	transactionEncryptUnauthorizedError    error = errors.New("Transaction encryption request is not authorized.")
	transactionEncryptOperationFormatError error = errors.New("Transaction encryption requires a valid transaction as payload.")
)

/*
	Helpers
*/

func (sv *server) makeChannelActionAndWait(wrappedRequest *executorRequest, request interface{}) *channels.ChannelsResponse {
	channelResponseChannel, err := sv.channelActionRequester(request)
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{err})
		return nil
	}
	channelResponsePtr, ok := <-channelResponseChannel
	if !ok ||
		channelResponsePtr.Result != channels.ChannelsSuccess {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{subsystemChannelClosed})
		return nil
	} else {
		return channelResponsePtr
	}
}

func (sv *server) channelActionPassthrough(wrappedRequest *executorRequest, request interface{}) {
	channelResponsePtr := sv.makeChannelActionAndWait(wrappedRequest, request)
	if channelResponsePtr != nil {
		channelResponseEncoded, _ := channelResponsePtr.Encode()
		sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, channelResponseEncoded, nil)
	}
}

func generateLockChannelRequest(channelId string, lockingType core.LockingType, lockType core.LockType) *locker.LockerRequest {
	return &locker.LockerRequest{
		Type:        locker.ChannelLock,
		LockingType: lockingType,
		Needs: []core.LockNeed{
			{
				LockType: lockType,
				Id:       channelId,
			},
		},
	}
}

func (sv *server) makeLockChannelRequest(wrappedRequest *executorRequest, channelId string, lockingType core.LockingType, lockType core.LockType) bool {
	// Make lock request
	lockRequest := generateLockChannelRequest(channelId, lockingType, lockType)
	lockChannel, errs := sv.lockerRequester(lockRequest)
	if len(errs) != 0 {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, errs)
		return false
	}

	// Wait for lock
	if lockResult := <-lockChannel; !lockResult {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{requestRejectedError})
		return false
	}
	return true
}

func (sv *server) lockChannel(wrappedRequest *executorRequest, channelId string) bool {
	return sv.makeLockChannelRequest(wrappedRequest, channelId, core.Locking, core.WriteLockType)
}

func (sv *server) unlockChannel(wrappedRequest *executorRequest, channelId string) bool {
	return sv.makeLockChannelRequest(wrappedRequest, channelId, core.Unlocking, core.WriteLockType)
}

func (sv *server) rlockChannel(wrappedRequest *executorRequest, channelId string) bool {
	return sv.makeLockChannelRequest(wrappedRequest, channelId, core.Locking, core.ReadLockType)
}

func (sv *server) runlockChannel(wrappedRequest *executorRequest, channelId string) bool {
	return sv.makeLockChannelRequest(wrappedRequest, channelId, core.Unlocking, core.ReadLockType)
}

/*
	Read channel
*/

func (sv *server) doReadChannel(wrappedRequest *executorRequest) {
	// Parse request
	request := &channels.ReadChannelRequest{}
	err := request.Decode(wrappedRequest.request)
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{err})
		return
	}

	// Read lock/Unlock channel
	channelId := wrappedRequest.metaFields.ChannelId
	if !sv.rlockChannel(wrappedRequest, channelId) {
		return
	}
	defer func() {
		if !sv.runlockChannel(wrappedRequest, channelId) {
			return
		}
	}()

	// Read Lock/Unlock certifier user object
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
	defer func() {
		usersSubsystemResponse, _ = sv.usersRequesterUnverified(nil, false, true, encodedUsersRequest)
		_ = <-usersSubsystemResponse
	}()

	// Check read channels permission
	certifierCheckSuccess := len(userResponsePtr.Data) == 1 && userResponsePtr.Data[0].Permissions.Channel.Read
	if !certifierCheckSuccess {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{channelReadUnauthorizedError})
		return
	}

	// Set channel id
	request.Id = channelId

	// Pass request through to channels subsystem
	sv.channelActionPassthrough(wrappedRequest, request)
}

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

	// Make sure channel is defined
	if request.Channel == nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{channelOpenNilChannelError})
		return
	}

	// Set channel id from operation meta fields
	channelId := wrappedRequest.metaFields.ChannelId
	request.Channel.Id = channelId

	// Set signers from decryptor
	if wrappedRequest.signers == nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{unverifiedChannelOpenError})
		return
	}
	request.Signers = wrappedRequest.signers

	// Lock/Unlock channel
	if !sv.lockChannel(wrappedRequest, channelId) {
		return
	}
	defer func() {
		if !sv.unlockChannel(wrappedRequest, channelId) {
			return
		}
	}()

	// Read Lock/Unlock certifier user object
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
	defer func() {
		usersSubsystemResponse, _ = sv.usersRequesterUnverified(nil, false, true, encodedUsersRequest)
		<-usersSubsystemResponse
	}()

	certifierCheckSuccess := len(userResponsePtr.Data) == 1 && userResponsePtr.Data[0].Permissions.Channel.Add
	if !certifierCheckSuccess {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{channelOpenUnauthorizedError})
		return
	}

	// Add key to keys subsystems
	if keyAddError := sv.keyAdder(request.Channel.KeyId, request.Key); keyAddError != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{keyAddError})
		return
	}

	// Send request through to channels subsystem
	sv.channelActionPassthrough(wrappedRequest, request)
}

/*
	Close channel
*/

func (sv *server) doCloseChannel(wrappedRequest *executorRequest) {
	// Parse request
	request := &channels.CloseChannelRequest{}
	err := request.Decode(wrappedRequest.request)
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{err})
		return
	}

	// Set channel id from operation meta fields
	request.Id = wrappedRequest.metaFields.ChannelId

	// Set signers from decryptor
	if wrappedRequest.signers == nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{unverifiedChannelOpenError})
		return
	}
	request.Signers = wrappedRequest.signers

	// Lock/Unlock channel
	if !sv.lockChannel(wrappedRequest, request.Id) {
		return
	}
	defer func() {
		if !sv.unlockChannel(wrappedRequest, request.Id) {
			return
		}
	}()

	// Send request through to channels subsystem
	sv.channelActionPassthrough(wrappedRequest, request)
}

/*
	Add message
*/

func (sv *server) doAddMessage(wrappedRequest *executorRequest) {
	// Lock/Unlock channel
	channelId := wrappedRequest.metaFields.ChannelId
	if !sv.lockChannel(wrappedRequest, channelId) {
		return
	}
	defer func() {
		if !sv.unlockChannel(wrappedRequest, channelId) {
			return
		}
	}()

	// Send request to channels subsystem based on type (operation buffering/ add message)
	var messageChannel chan *channels.MessagesResponse
	var requestErr error
	if wrappedRequest.failedOperation == nil {
		messageChannel, requestErr = sv.messageAdder(&channels.AddMessageRequest{
			ChannelId: wrappedRequest.metaFields.ChannelId,
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
		sv.responseReporter(wrappedRequest.ticket, status.FailedStatus, status.FailedReason, messageResponse, nil)
	} else {
		sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, messageResponse, nil)
	}
}

/*
	Subscribe to channel
*/

func (sv *server) doSubscribeChannel(wrappedRequest *executorRequest) {
	request := &channels.SubscribeRequest{}

	// Set channel id from operation meta fields
	channelId := wrappedRequest.metaFields.ChannelId
	request.ChannelId = channelId

	// Set signers from decryptor
	if wrappedRequest.signers == nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{unverifiedChannelSubscribeError})
		return
	}
	request.Signers = wrappedRequest.signers

	// Read Lock/Unlock channel
	if !sv.rlockChannel(wrappedRequest, channelId) {
		return
	}
	defer func() {
		if !sv.runlockChannel(wrappedRequest, channelId) {
			return
		}
	}()

	// Make request to channels subsystem
	listenersResponseChannel, err := sv.channelListenersRequester(request)
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{requestRejectedError})
		return
	}
	listenersResponsePtr, ok := <-listenersResponseChannel
	if !ok {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{subsystemChannelClosed})
	} else {
		sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, listenersResponsePtr, nil)
	}
}

/*
	Encrypt operation based on channel key
*/

func (sv *server) doChannelEncrypt(wrappedRequest *executorRequest) {
	// Interpret payload as operation json string
	op := &core.Operation{}
	err := op.Decode([]byte(wrappedRequest.request))
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{channelEncryptOperationFormatError})
		return
	}

	// Decode operation payload
	opPayloadBytes, err := op.DecodePayload()
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{channelEncryptOperationFormatError})
		return
	}

	// Read Lock/Unlock channel
	channelId := wrappedRequest.metaFields.ChannelId
	if !sv.rlockChannel(wrappedRequest, channelId) {
		return
	}
	defer func() {
		if !sv.runlockChannel(wrappedRequest, channelId) {
			return
		}
	}()

	// Read channel
	request := &channels.ReadChannelRequest{
		Id: channelId,
	}
	channelResponse := sv.makeChannelActionAndWait(wrappedRequest, request)
	if channelResponse == nil {
		return
	}

	// Check that channel was opened and certifier has write permissions
	channelOpened := channelResponse.Channel.State == channels.ChannelObjectOpenState || channelResponse.Channel.State == channels.ChannelObjectClosedState
	certifierChannelPermisisons, isMemberOfChannel := channelResponse.Channel.Permissions.Users[wrappedRequest.signers.CertifierId]
	authorized := channelOpened && isMemberOfChannel && certifierChannelPermisisons.Write
	if !authorized {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{channelEncryptUnauthorizedError})
		return
	}

	// Encrypt using keys subsystem
	encrypted, nonce, err := sv.keyEncryptor(channelResponse.Channel.KeyId, opPayloadBytes)
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{err})
		return
	}

	// Fill in operation
	op.Encryption = core.OperationEncryptionFields{
		Encrypted: true,
		KeyId:     channelResponse.Channel.KeyId,
		Nonce:     core.Base64EncodeToString(nonce),
	}
	op.Meta.ChannelId = channelResponse.Channel.Id
	op.Payload = core.CiphertextEncode(encrypted)

	sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, op, nil)
}
