package keys

import (
	"errors"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
)

/*
	Errors
*/
var (
	invalidRequestFormatError error = errors.New("Invalid request format.")
	addingKeyFailedError      error = errors.New("Failed to add key.")
	encryptionFailedError     error = errors.New("Failed to do encryption operation.")
)

/*
	Logging
*/
var (
	log             *core.LoggingHandler
	shutdownProgram core.ShutdownLambda
)

/*
	Server definitions
*/

type Config struct {
	NumWorkers int
}

type server struct {
	isInitialized bool
	store         *memstore.Memstore
}

var (
	serverSingleton server
	serverHandler   *gofarm.ServerHandler
)

/*
	Server helpers
*/

func provisionServerOnce() {
	if serverHandler == nil {
		serverHandler = gofarm.ProvisionServer()
	}
}

func makeGenericRequest(rqPtr *keyRequest) (chan *gofarm.Response, error) {
	// Validate request
	if !rqPtr.validate() {
		return nil, invalidRequestFormatError
	}

	// Make request to server
	nativeResponseChannel, err := serverHandler.MakeRequest(rqPtr)
	if err != nil {
		return nil, err
	}

	return nativeResponseChannel, nil
}

func makeGenericEncryptionRequest(request *keyRequest) ([]byte, error) {
	// Make request
	nativeResponseChannel, err := makeGenericRequest(request)
	if err != nil {
		return nil, err
	}

	// Wait and pass through result
	nativeResponse, ok := <-nativeResponseChannel
	if ok {
		resp := (*nativeResponse).(*keyResponse)
		if resp.Result == Success {
			return resp.Payload, nil
		}
	}

	return nil, encryptionFailedError
}

/*
	Server API
*/

func StartServer(
	conf Config,
	loggingHandler *core.LoggingHandler,
	shutdownLambda core.ShutdownLambda,
) error {
	provisionServerOnce()
	if !serverSingleton.isInitialized {
		log = loggingHandler
		shutdownProgram = shutdownLambda
		serverSingleton.isInitialized = true
		serverHandler.ResetServer()
		serverHandler.InitServer(&serverSingleton)
	}
	return serverHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownServer() {
	provisionServerOnce()
	serverHandler.ShutdownServer()
}

func AddKey(keyId string, key []byte) error {
	nativeResponseChannel, err := makeGenericRequest(&keyRequest{
		Type:    AddKeyRequest,
		KeyId:   keyId,
		Payload: key,
	})
	if err != nil {
		return err
	}

	// Wait and pass through result
	nativeResponse, ok := <-nativeResponseChannel
	if ok && (*nativeResponse).(*keyResponse).Result == Success {
		return nil
	}

	return addingKeyFailedError
}

func Encrypt(keyId string, plaintext []byte) ([]byte, []byte, error) {
	nonce := core.GenerateSymmetricNonce()

	encrypted, err := makeGenericEncryptionRequest(&keyRequest{
		Type:    EncryptRequest,
		KeyId:   keyId,
		Payload: plaintext,
		Nonce:   nonce,
	})

	if encrypted == nil {
		return nil, nil, err
	}
	return encrypted, nonce, nil
}

func Decrypt(keyId string, nonce []byte, ciphertext []byte) ([]byte, error) {
	decrypted, err := makeGenericEncryptionRequest(&keyRequest{
		Type:    DecryptRequest,
		KeyId:   keyId,
		Payload: ciphertext,
		Nonce:   nonce,
	})

	if decrypted == nil {
		return nil, err
	}
	return decrypted, nil
}

/*
	Server implementation
*/

func (sv *server) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		sv.store = memstore.New(getIndexes())
	}
	log.Debugf(daemonStartLogMsg)
	return nil
}

func (sv *server) Shutdown() error {
	log.Debugf(daemonShutdownLogMsg)
	return nil
}

func (sv *server) Work(request *gofarm.Request) *gofarm.Response {
	log.Debugf(runningRequestLogMsg)

	rqPtr := (*request).(*keyRequest)

	/*
		Run request
	*/
	switch rqPtr.Type {
	case AddKeyRequest:
		sv.store.AddOrGet(rqPtr.makeRecord())
		return successRequest(nil)
	case DecryptRequest:
		// Get key
		storedRecord := sv.store.Get(rqPtr.makeSearchRecord(), recordIdIndex)
		if storedRecord == nil {
			return failRequest(DecryptionFailure)
		}

		// Decrypt
		aead, _ := core.NewAead(storedRecord.(*keyRecord).Key)
		decrypted, err := core.SymmetricDecrypt(
			aead,
			rqPtr.Payload[:0],
			rqPtr.Nonce,
			rqPtr.Payload,
		)
		if err != nil {
			return failRequest(DecryptionFailure)
		} else {
			return successRequest(decrypted)
		}
	case EncryptRequest:
		// Get key
		storedRecord := sv.store.Get(rqPtr.makeSearchRecord(), recordIdIndex)
		if storedRecord == nil {
			return failRequest(EncryptionFailure)
		}

		// Encrypt
		aead, _ := core.NewAead(storedRecord.(*keyRecord).Key)
		payloadCiphertext := core.SymmetricEncrypt(
			aead,
			rqPtr.Payload[:0],
			rqPtr.Nonce,
			rqPtr.Payload,
		)
		return successRequest(payloadCiphertext)
	}

	return nil
}

func failRequest(responseCode keyResponseCode) *gofarm.Response {
	log.Debugf(failRequestLogMsg)
	userRespPtr := &keyResponse{
		Result: responseCode,
	}
	var nativeResp gofarm.Response = userRespPtr
	return &nativeResp
}

func successRequest(payload []byte) *gofarm.Response {
	log.Debugf(successRequestLogMsg)
	userRespPtr := &keyResponse{
		Result:  Success,
		Payload: payload,
	}
	var nativeResp gofarm.Response = userRespPtr
	return &nativeResp
}
