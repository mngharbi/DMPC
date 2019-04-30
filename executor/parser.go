package executor

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
)

/*
	Internal request structure
*/
type executorRequest struct {
	isVerified      bool
	metaFields      *core.OperationMetaFields
	signers         *core.VerifiedSigners
	ticket          status.Ticket
	request         []byte
	failedOperation *core.Operation
}

/*
	Utilities
*/
func isValidRequestType(requestType core.RequestType) bool {
	return core.UsersRequestType <= requestType && requestType <= core.ChannelEncryptType
}
