package executor

import (
	"crypto/rsa"
	"github.com/mngharbi/DMPC/core"
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

/*
	Encrypt transaction
*/

func (sv *server) doTransactionEncrypt(wrappedRequest *executorRequest) {
	// Interpret payload as transaction json
	ts := &core.Transaction{}
	err := ts.Decode([]byte(wrappedRequest.request))
	if err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{transactionEncryptOperationFormatError})
		return
	}

	// Get and RLock transaction receipients
	receipientIds := []string{}
	for receipientId := range ts.Encryption.Challenges {
		receipientIds = append(receipientIds, receipientId)
	}
	usersRequest := &users.UserRequest{
		Type:      users.ReadRequest,
		Timestamp: wrappedRequest.metaFields.Timestamp,
		Fields:    receipientIds,
	}
	encodedUsersRequest, _ := usersRequest.Encode()
	usersSubsystemResponse, errs := sv.usersRequester(wrappedRequest.signers, true, false, encodedUsersRequest)
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
		usersSubsystemResponse, _ = sv.usersRequester(wrappedRequest.signers, false, true, encodedUsersRequest)
		_ = <-usersSubsystemResponse
	}()

	// Build keys array
	keys := []*rsa.PublicKey{}
	for _, userObject := range userResponsePtr.Data {
		key, err := core.PublicStringToAsymKey(userObject.EncKey)
		if err != nil {
			sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{err})
			return
		}
		keys = append(keys, key)
	}

	// Do transaction encryption
	if err = ts.Encrypt(keys); err != nil {
		sv.reportRejection(wrappedRequest.ticket, status.RejectedReason, []error{err})
		return
	}

	tsEncoded, _ := ts.Encode()
	sv.responseReporter(wrappedRequest.ticket, status.SuccessStatus, status.NoReason, tsEncoded, nil)
}
