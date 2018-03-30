package decryptor

import (
	"crypto/rsa"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
	"github.com/mngharbi/gofarm"
	"time"
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
type UsersDecodedRequester func(*users.UserRequest) (chan *users.UserResponse, []error)
type KeyRequester func(string) []byte
type ExecutorRequester func(int, string, string, []byte) int

/*
	Server API
*/
func InitializeServer(
	globalKey *rsa.PrivateKey,
	usersRequester UsersDecodedRequester,
	keyRequester KeyRequester,
	executorRequester ExecutorRequester,
) {
	serverSingleton.globalKey = globalKey
	serverSingleton.usersRequester = usersRequester
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

	return nativeResponseChannel, []error{err}
}

/*
	Server implementation
*/

type server struct {
	// Asymmetric key
	globalKey *rsa.PrivateKey

	// Requester lambdas
	usersRequester    UsersDecodedRequester
	keyRequester      KeyRequester
	executorRequester ExecutorRequester
}

var serverSingleton server

func (sv *server) Start(_ gofarm.Config, _ bool) error { return nil }

func (sv *server) Shutdown() error { return nil }

func (sv *server) Work(nativeRequest *gofarm.Request) *gofarm.Response {
	decryptorWrapped := (*nativeRequest).(*decryptorRequest)

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

	// Get signing keys from users subsystem
	usersChannel, errs := sv.usersRequester(&users.UserRequest{
		Type: users.ReadRequest,
		Fields: []string{
			permanentEncrypted.Issue.Id,
			permanentEncrypted.Certification.Id,
		},
		Timestamp: time.Now(),
	})
	if len(errs) > 0 {
		return failRequest(PermanentDecryptionError)
	}
	usersRespPtr := <-usersChannel
	if usersRespPtr.Result != users.Success {
		return failRequest(PermanentDecryptionError)
	}
	issuerSignKey, _ := core.StringToAsymKey(usersRespPtr.Data[0].SignKey)
	certifierSignKey, _ := core.StringToAsymKey(usersRespPtr.Data[1].SignKey)

	// Permanent encryption
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
