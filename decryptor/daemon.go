package decryptor

import (
	"crypto/rsa"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/gofarm"
)

type Config struct {
	NumWorkers int
}

type decryptorRequest struct {
	isVerified bool
	rawRequest []byte
}

/*
	Logging
*/
var (
	log             *core.LoggingHandler
	shutdownProgram core.ShutdownLambda
)

/*
	Server API
*/

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

func MakeUnverifiedRequest(rawRequest []byte) (chan *gofarm.Response, []error) {
	return makeRequest(rawRequest, true)
}

func MakeRequest(rawRequest []byte) (chan *gofarm.Response, []error) {
	return makeRequest(rawRequest, false)
}

func makeRequest(rawRequest []byte, skipPermissions bool) (chan *gofarm.Response, []error) {
	nativeResponseChannel, err := serverHandler.MakeRequest(&decryptorRequest{
		isVerified: !skipPermissions,
		rawRequest: rawRequest,
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

	if len(decryptorWrapped.rawRequest) == 0 {
		return failRequest(TemporaryDecryptionError)
	}

	// Decode payload
	temporaryEncrypted := &core.TemporaryEncryptedOperation{}
	err := temporaryEncrypted.Decode(decryptorWrapped.rawRequest)
	if err != nil {
		return failRequest(TemporaryDecryptionError)
	}

	// Temporary decryption
	permanentEncrypted, err := temporaryEncrypted.Decrypt(sv.globalKey)
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
	ticketId, err := sv.executorRequester(
		decryptorWrapped.isVerified,
		permanentEncrypted.Meta.RequestType,
		signers,
		plaintextBytes,
	)
	if err != nil {
		return failRequest(ExecutorError)
	}

	// @TODO: Change type to status.Ticket
	return successRequest(string(ticketId))
}

func failRequest(errorType int) *gofarm.Response {
	decryptorRespPtr := &DecryptorResponse{
		Result: errorType,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}

func successRequest(ticketId string) *gofarm.Response {
	decryptorRespPtr := &DecryptorResponse{
		Result: Success,
		Ticket: ticketId,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}
