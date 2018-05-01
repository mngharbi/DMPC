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
	requestType int
	signers     *core.VerifiedSigners
	ticket      status.Ticket
	request     []byte
}

/*
	Utilities
*/
func isValidRequestType(requestType int) bool {
	return UsersRequest <= requestType && requestType <= UsersRequest
}
