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
	operation   *core.PermanentEncryptedOperation
}

func provisionServerOnce() {
	if serverHandler == nil {
		serverHandler = gofarm.ProvisionServer()
	}
}

func InitializeServer(
	globalKey *rsa.PrivateKey,
	usersSignKeyRequester core.UsersSignKeyRequester,
	keyRequester core.KeyRequester,
	executorRequester executor.Requester,
	loggingHandler *core.LoggingHandler,
	shutdownLambda core.ShutdownLambda,
) {
	provisionServerOnce()
	serverSingleton.globalKey = globalKey
	serverSingleton.usersSignKeyRequester = usersSignKeyRequester
	serverSingleton.keyRequester = keyRequester
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
func MakeOperationRequest(operation *core.PermanentEncryptedOperation) (chan *gofarm.Response, []error) {
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
	keyRequester          core.KeyRequester
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

func decryptTransaction(transaction *core.Transaction, globalKey *rsa.PrivateKey) (*core.PermanentEncryptedOperation, bool) {
	permanentEncrypted, err := transaction.Decrypt(globalKey)
	if err != nil {
		return nil, false
	}
	if len(permanentEncrypted.Payload) == 0 {
		return nil, false
	}
	return permanentEncrypted, true
}

func decryptOperation(operation *core.PermanentEncryptedOperation, keyRequester core.KeyRequester) ([]byte, bool) {
	payload, err := operation.Decrypt(keyRequester)
	return payload, err == nil
}

func verifyPayload(operation *core.PermanentEncryptedOperation, payload []byte, usersSignKeyRequester core.UsersSignKeyRequester) bool {
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

	var permanentEncrypted *core.PermanentEncryptedOperation = decryptorWrapped.operation

	// Decrypt transaction if any
	if permanentEncrypted == nil {
		var success bool
		if permanentEncrypted, success = decryptTransaction(decryptorWrapped.transaction, sv.globalKey); !success {
			return failRequest(TransactionDecryptionError)
		}
	}

	// Operation decryption
	plaintextBytes, decryptionSuccess := decryptOperation(permanentEncrypted, sv.keyRequester)

	// Determine if we should fail
	droppable := permanentEncrypted.ShouldDrop()
	if !decryptionSuccess && droppable {
		return failRequest(PermanentDecryptionError)
	}

	// Verify signatures if not skipping verification
	var signers *core.VerifiedSigners
	var verificationSuccess bool
	if decryptorWrapped.isVerified && decryptionSuccess {
		verificationSuccess = verifyPayload(permanentEncrypted, plaintextBytes, sv.usersSignKeyRequester)

		// Only drop request if it's droppable (otherwise skip verification)
		if !verificationSuccess && droppable {
			return failRequest(VerificationError)
		}

		// Build signers structure
		if verificationSuccess {
			signers = &core.VerifiedSigners{
				IssuerId:    permanentEncrypted.Issue.Id,
				CertifierId: permanentEncrypted.Certification.Id,
			}
		}
	}

	// If anything failed, mark for buffering
	var failedEncryptedOperation *core.PermanentEncryptedOperation
	if !decryptionSuccess || !verificationSuccess {
		failedEncryptedOperation = permanentEncrypted
		plaintextBytes = nil
		failedEncryptedOperation.Meta.Buffered = true
	}

	// Send raw bytes and metadata to executor
	ticket, err := sv.executorRequester(
		decryptorWrapped.isVerified,
		permanentEncrypted.Meta.RequestType,
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
