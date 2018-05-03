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
type Requester func(*core.TemporaryEncryptedOperation) (chan *gofarm.Response, []error)

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
	isVerified bool
	operation  *core.TemporaryEncryptedOperation
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

func MakeUnverifiedEncodedRequest(encodedRequest []byte) (chan *gofarm.Response, []error) {
	return makeEncodedRequest(encodedRequest, true)
}

func MakeEncodedRequest(encodedRequest []byte) (chan *gofarm.Response, []error) {
	return makeEncodedRequest(encodedRequest, false)
}

func makeEncodedRequest(encodedRequest []byte, skipPermissions bool) (chan *gofarm.Response, []error) {
	// Decode payload
	temporaryEncrypted := &core.TemporaryEncryptedOperation{}
	err := temporaryEncrypted.Decode(encodedRequest)
	if err != nil {
		return nil, []error{err}
	}

	return makeRequest(temporaryEncrypted, skipPermissions)
}

func MakeUnverifiedRequest(temporaryEncrypted *core.TemporaryEncryptedOperation) (chan *gofarm.Response, []error) {
	return makeRequest(temporaryEncrypted, true)
}

func MakeRequest(temporaryEncrypted *core.TemporaryEncryptedOperation) (chan *gofarm.Response, []error) {
	return makeRequest(temporaryEncrypted, false)
}

func makeRequest(temporaryEncrypted *core.TemporaryEncryptedOperation, skipPermissions bool) (chan *gofarm.Response, []error) {
	nativeResponseChannel, err := serverHandler.MakeRequest(&decryptorRequest{
		isVerified: !skipPermissions,
		operation:  temporaryEncrypted,
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

func (sv *server) Start(_ gofarm.Config, _ bool) error { return nil }

func (sv *server) Shutdown() error { return nil }

func (sv *server) Work(nativeRequest *gofarm.Request) *gofarm.Response {
	decryptorWrapped := (*nativeRequest).(*decryptorRequest)

	// Temporary decryption
	permanentEncrypted, err := decryptorWrapped.operation.Decrypt(sv.globalKey)
	if err != nil {
		return failRequest(TemporaryDecryptionError)
	}
	if len(permanentEncrypted.Payload) == 0 {
		return failRequest(TemporaryDecryptionError)
	}

	// Permanent decryption
	plaintextBytes, err := permanentEncrypted.Decrypt(sv.keyRequester)
	if err != nil {
		return failRequest(PermanentDecryptionError)
	}

	// Get signing keys from users subsystem
	if decryptorWrapped.isVerified {
		keys, err := sv.usersSignKeyRequester([]string{
			permanentEncrypted.Issue.Id,
			permanentEncrypted.Certification.Id,
		})
		if err != nil {
			return failRequest(VerificationError)
		}
		issuerSignKey := keys[0]
		certifierSignKey := keys[1]
		if verification := permanentEncrypted.Verify(issuerSignKey, certifierSignKey, plaintextBytes); verification != nil {
			return failRequest(VerificationError)
		}
	}

	// Build signers structure
	var signers *core.VerifiedSigners
	if decryptorWrapped.isVerified {
		signers = &core.VerifiedSigners{
			IssuerId:    permanentEncrypted.Issue.Id,
			CertifierId: permanentEncrypted.Certification.Id,
		}
	}

	// Send raw bytes and metadata to executor
	ticket, err := sv.executorRequester(
		decryptorWrapped.isVerified,
		permanentEncrypted.Meta.RequestType,
		signers,
		plaintextBytes,
	)
	if err != nil {
		return failRequest(ExecutorError)
	}

	return successRequest(ticket)
}

func failRequest(errorType int) *gofarm.Response {
	decryptorRespPtr := &DecryptorResponse{
		Result: errorType,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}

func successRequest(ticket status.Ticket) *gofarm.Response {
	decryptorRespPtr := &DecryptorResponse{
		Result: Success,
		Ticket: ticket,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}
