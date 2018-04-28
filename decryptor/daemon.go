package decryptor

import (
	"crypto/rsa"
	"github.com/mngharbi/DMPC/core"
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
	Types of lambdas to call other subsystems
*/
type UsersSignKeyRequester func([]string) ([]*rsa.PublicKey, error)
type KeyRequester func(string) []byte
type ExecutorRequester func(bool, int, string, string, []byte) string

/*
	Logging
*/
var (
	log *core.LoggingHandler
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
	usersSignKeyRequester UsersSignKeyRequester,
	keyRequester KeyRequester,
	executorRequester ExecutorRequester,
	loggingHandler *core.LoggingHandler,
) {
	provisionServerOnce()
	log = loggingHandler
	serverSingleton.globalKey = globalKey
	serverSingleton.usersSignKeyRequester = usersSignKeyRequester
	serverSingleton.keyRequester = keyRequester
	serverSingleton.executorRequester = executorRequester
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
	usersSignKeyRequester UsersSignKeyRequester
	keyRequester          KeyRequester
	executorRequester     ExecutorRequester
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

	// Get signing keys from users subsystem
	keys, err := sv.usersSignKeyRequester([]string{
		permanentEncrypted.Issue.Id,
		permanentEncrypted.Certification.Id,
	})
	if err != nil {
		return failRequest(PermanentDecryptionError)
	}
	issuerSignKey := keys[0]
	certifierSignKey := keys[1]

	// Permanent decryption
	plaintextBytes, err := permanentEncrypted.Decrypt(sv.keyRequester, issuerSignKey, certifierSignKey)
	if err != nil {
		return failRequest(PermanentDecryptionError)
	}

	// Send raw bytes and metadata to executor
	ticketId := sv.executorRequester(
		decryptorWrapped.isVerified,
		permanentEncrypted.Meta.RequestType,
		permanentEncrypted.Issue.Id,
		permanentEncrypted.Certification.Id,
		plaintextBytes,
	)

	return successRequest(ticketId)
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
