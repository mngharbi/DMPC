package pipeline

/*
	Logging messages
*/
const (
	startLogMsg                string = "Starting up pipeline server"
	shutdownLogMsg             string = "Shutting down pipeline server"
	connectionRequestedLogMsg  string = "Got connection request to pipeline server"
	readTransactionLogMsg      string = "Read transaction in pipeline server"
	invalidTransactionLogMsg   string = "Received invalid transaction in pipeline server."
	transactionRejected        string = "Transaction rejected in decryptor. err=%+v"
	invalidDecryptorResponse   string = "Decryptor has invalid response. response=%+v"
	updateChannelFailureLogMsg string = "Ticket[%v]: Failed to get status update channel. err=%+v"
	unsubscribeFailedLogMsg    string = "Ticket[%v]: Failed to unsubscribe. err=%+v"
)

/*
	Info messages
*/
const (
	startListeningInfoMsg string = "Pipeline server started listening on port %v"
	shutdownInfoMsg       string = "Server was shutdown"
)

/*
	Error messages
*/
const (
	serverCannotListenErrorMsg string = "Pipeline server could not start listening on %v. Error: %v"
)
