package decryptor

import (
	"crypto/rsa"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/gofarm"
)

/*
	Function to accept a structured request
*/
type Requester func(*core.Transaction) (chan *gofarm.Response, []error)

/*
	Logging
*/
var (
	log             *core.LoggingHandler
	shutdownProgram core.ShutdownLambda
)

/*
	Server configuration
*/

type Config struct {
	NumWorkers int
}

/*
	Server API
*/

type decryptorRequest struct {
	isVerified  bool
	transaction *core.Transaction
	operation   *core.Operation
}

func provisionServerOnce() {
	if serverHandler == nil {
		serverHandler = gofarm.ProvisionServer()
	}
}

func InitializeServer(
	globalKey *rsa.PrivateKey,
	usersSignKeyRequester core.UsersSignKeyRequester,
	keyDecryptor core.Decryptor,
	executorRequester executor.Requester,
	loggingHandler *core.LoggingHandler,
	shutdownLambda core.ShutdownLambda,
) {
	provisionServerOnce()
	serverSingleton.globalKey = globalKey
	serverSingleton.usersSignKeyRequester = usersSignKeyRequester
	serverSingleton.keyDecryptor = keyDecryptor
	serverSingleton.executorRequester = executorRequester
	log = loggingHandler
	shutdownProgram = shutdownLambda
	serverHandler.InitServer(&serverSingleton)
}

func StartServer(conf Config) error {
	provisionServerOnce()
	return serverHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownServer() {
	provisionServerOnce()
	serverHandler.ShutdownServer()
}

/*
	Transaction requests
*/
func MakeUnverifiedEncodedTransactionRequest(encodedRequest []byte) (chan *gofarm.Response, []error) {
	return makeEncodedTransactionRequest(encodedRequest, true)
}

func MakeEncodedTransactionRequest(encodedRequest []byte) (chan *gofarm.Response, []error) {
	return makeEncodedTransactionRequest(encodedRequest, false)
}

func makeEncodedTransactionRequest(encodedRequest []byte, skipPermissions bool) (chan *gofarm.Response, []error) {
	// Decode payload
	transaction := &core.Transaction{}
	err := transaction.Decode(encodedRequest)
	if err != nil {
		return nil, []error{err}
	}

	return makeTransactionRequest(transaction, skipPermissions)
}

func MakeUnverifiedTransactionRequest(transaction *core.Transaction) (chan *gofarm.Response, []error) {
	return makeTransactionRequest(transaction, true)
}

func MakeTransactionRequest(transaction *core.Transaction) (chan *gofarm.Response, []error) {
	return makeTransactionRequest(transaction, false)
}

func makeTransactionRequest(transaction *core.Transaction, skipPermissions bool) (chan *gofarm.Response, []error) {
	log.Debugf(receivedRequestLogMsg)
	nativeResponseChannel, err := serverHandler.MakeRequest(&decryptorRequest{
		isVerified:  !skipPermissions,
		transaction: transaction,
	})
	if err != nil {
		return nil, []error{err}
	}

	return nativeResponseChannel, nil
}

/*
	Operation requests
*/
func MakeOperationRequest(operation *core.Operation) (chan *gofarm.Response, []error) {
	log.Debugf(receivedRequestLogMsg)
	nativeResponseChannel, err := serverHandler.MakeRequest(&decryptorRequest{
		isVerified: true,
		operation:  operation,
	})
	if err != nil {
		return nil, []error{err}
	}

	return nativeResponseChannel, nil
}

/*
	Server implementation
*/

var (
	serverSingleton server
	serverHandler   *gofarm.ServerHandler
)

type server struct {
	// Asymmetric key
	globalKey *rsa.PrivateKey

	// Requester lambdas
	usersSignKeyRequester core.UsersSignKeyRequester
	keyDecryptor          core.Decryptor
	executorRequester     executor.Requester
}

func (sv *server) Start(_ gofarm.Config, _ bool) error {
	log.Debugf(daemonStartLogMsg)
	return nil
}

func (sv *server) Shutdown() error {
	log.Debugf(daemonShutdownLogMsg)
	return nil
}

func decryptTransaction(transaction *core.Transaction, globalKey *rsa.PrivateKey) (*core.Operation, bool) {
	operation, err := transaction.Decrypt(globalKey)
	if err != nil {
		return nil, false
	}
	if len(operation.Payload) == 0 {
		return nil, false
	}
	return operation, true
}

func decryptOperation(operation *core.Operation, keyDecryptor core.Decryptor) ([]byte, bool) {
	payload, err := operation.Decrypt(keyDecryptor)
	return payload, err == nil
}

func verifyPayload(operation *core.Operation, payload []byte, usersSignKeyRequester core.UsersSignKeyRequester) bool {
	keys, err := usersSignKeyRequester([]string{
		operation.Issue.Id,
		operation.Certification.Id,
	})
	if err != nil {
		return false
	}
	issuerSignKey := keys[0]
	certifierSignKey := keys[1]
	verification := operation.Verify(issuerSignKey, certifierSignKey, payload)
	return verification == nil
}

func (sv *server) Work(nativeRequest *gofarm.Request) *gofarm.Response {
	log.Debugf(runningRequestLogMsg)
	decryptorWrapped := (*nativeRequest).(*decryptorRequest)

	var operation *core.Operation = decryptorWrapped.operation

	// Decrypt transaction if any
	if operation == nil {
		var success bool
		if operation, success = decryptTransaction(decryptorWrapped.transaction, sv.globalKey); !success {
			return failRequest(TransactionDecryptionError)
		}
	}

	// Operation decryption
	plaintextBytes, decryptionSuccess := decryptOperation(operation, sv.keyDecryptor)

	// Determine if we should fail
	droppable := operation.ShouldDrop()
	if !decryptionSuccess && droppable {
		return failRequest(PermanentDecryptionError)
	}

	// Verify signatures if not skipping verification
	var signers *core.VerifiedSigners
	var verificationSuccess bool = true
	if decryptorWrapped.isVerified && decryptionSuccess {
		verificationSuccess = verifyPayload(operation, plaintextBytes, sv.usersSignKeyRequester)

		// Only drop request if it's droppable (otherwise skip verification)
		if !verificationSuccess && droppable {
			return failRequest(VerificationError)
		}

		// Build signers structure
		if verificationSuccess {
			signers = &core.VerifiedSigners{
				IssuerId:    operation.Issue.Id,
				CertifierId: operation.Certification.Id,
			}
		}
	}

	// If anything failed, mark for buffering
	var failedEncryptedOperation *core.Operation
	if !decryptionSuccess || !verificationSuccess {
		failedEncryptedOperation = operation
		plaintextBytes = nil
		failedEncryptedOperation.Meta.Buffered = true
	}

	// Send raw bytes and metadata to executor
	ticket, err := sv.executorRequester(
		decryptorWrapped.isVerified,
		operation.Meta.RequestType,
		signers,
		plaintextBytes,
		failedEncryptedOperation,
	)
	if err != nil {
		return failRequest(ExecutorError)
	}

	return successRequest(ticket)
}

func failRequest(errorType int) *gofarm.Response {
	log.Infof(failRequestLogMsg)
	decryptorRespPtr := &DecryptorResponse{
		Result: errorType,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}

func successRequest(ticket status.Ticket) *gofarm.Response {
	log.Debugf(successRequestLogMsg)
	decryptorRespPtr := &DecryptorResponse{
		Result: Success,
		Ticket: ticket,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}
