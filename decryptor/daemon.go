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
type ExecutorRequester func(int, string, string, []byte) int

/*
	Server API
*/
func InitializeServer(
	globalKey *rsa.PrivateKey,
	usersSignKeyRequester UsersSignKeyRequester,
	keyRequester KeyRequester,
	executorRequester ExecutorRequester,
) {
	serverSingleton.globalKey = globalKey
	serverSingleton.usersSignKeyRequester = usersSignKeyRequester
	serverSingleton.keyRequester = keyRequester
	serverSingleton.executorRequester = executorRequester
	gofarm.InitServer(&serverSingleton)
}

func StartServer(conf Config) error {
	return gofarm.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownServer() {
	gofarm.ShutdownServer()
}

func MakeUnverifiedRequest(rawRequest []byte) (chan *gofarm.Response, []error) {
	return makeRequest(rawRequest, true)
}

func MakeRequest(rawRequest []byte) (chan *gofarm.Response, []error) {
	return makeRequest(rawRequest, false)
}

func makeRequest(rawRequest []byte, skipPermissions bool) (chan *gofarm.Response, []error) {
	nativeResponseChannel, err := gofarm.MakeRequest(&decryptorRequest{
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

type server struct {
	// Asymmetric key
	globalKey *rsa.PrivateKey

	// Requester lambdas
	usersSignKeyRequester UsersSignKeyRequester
	keyRequester          KeyRequester
	executorRequester     ExecutorRequester
}

var serverSingleton server

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
	ticketNb := sv.executorRequester(
		permanentEncrypted.Meta.RequestType,
		permanentEncrypted.Issue.Id,
		permanentEncrypted.Certification.Id,
		plaintextBytes,
	)

	return successRequest(ticketNb)
}

func failRequest(errorType int) *gofarm.Response {
	decryptorRespPtr := &DecryptorResponse{
		Result: errorType,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}

func successRequest(ticketNb int) *gofarm.Response {
	decryptorRespPtr := &DecryptorResponse{
		Result: Success,
		Ticket: ticketNb,
	}

	var nativeResp gofarm.Response = decryptorRespPtr
	return &nativeResp
}
