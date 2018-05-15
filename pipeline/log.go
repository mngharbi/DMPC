package pipeline

/*
	Logging messages
*/
const (
	startLogMsg string = "Starting up pipeline server"
	shutdownLogMsg string = "Shutting down pipeline server"
	connectionRequestedLogMsg string = "Got connection request to pipeline server"
	invalidOperationLogMsg string = "Received invalid operation in pipeline server"
)

/*
	Info messages
*/
const (
	startListeningInfoMsg string = "Pipeline server started listening on port %v"
	shutdownInfoMsg string = "Server was shutdown"
)

/*
	Error messages
*/
const (
	serverCannotListenErrorMsg string = "Pipeline server could not start listening on %v. Error: %v"
)
