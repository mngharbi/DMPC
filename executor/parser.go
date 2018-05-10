package executor

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
)

/*
	Request types
*/
const (
	UsersRequest = iota
)

/*
	Internal request structure
*/
type executorRequest struct {
	isVerified  bool
	requestType core.RequestType
	signers     *core.VerifiedSigners
	ticket      status.Ticket
	request     []byte
}

/*
	Utilities
*/
func isValidRequestType(requestType core.RequestType) bool {
	return core.UsersRequestType <= requestType && requestType <= core.UsersRequestType
}
