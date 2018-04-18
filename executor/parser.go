package executor

import (
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
	issuerId    string
	certifierId string
	ticket      status.Ticket
	request     []byte
}

/*
	Utilities
*/
func isValidRequestType(requestType int) bool {
	return UsersRequest <= requestType && requestType <= UsersRequest
}
