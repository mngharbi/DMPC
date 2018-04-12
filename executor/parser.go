package executor

/*
	Request types
*/

const (
	UsersRequest = iota
)

/*
	Aliases for statuses
*/

const (
	QueuedStatus = iota
	RunningStatus
	SuccessStatus
	FailedStatus
)

const (
	NoReason = iota
	RejectedReason
	FailedReason
)

/*
	Internal request structure
*/

type executorRequest struct {
	isVerified  bool
	requestType int
	issuerId    string
	certifierId string
	ticket      int
	request     []byte
}

/*
	Utilities
*/

func isValidRequestType(requestType int) bool {
	return UsersRequest <= requestType && requestType <= UsersRequest
}
