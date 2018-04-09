package executor

/*
	Request types
*/
const (
	UsersRequest = iota
)

func isValidRequestType(requestType int) bool {
	return UsersRequest <= requestType && requestType <= UsersRequest
}

type executorRequest struct {
	isVerified  bool
	requestType int
	issuerId    string
	certifierId string
	ticket      int
	request     []byte
}
