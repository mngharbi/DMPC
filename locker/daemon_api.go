package locker

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/gofarm"
)

/*
	Requester type
*/
type Requester func(*LockerRequest) (chan bool, []error)

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

var serverHandler *gofarm.ServerHandler

type Config struct {
	NumWorkers int
}

func provisionServerOnce() {
	if serverHandler == nil {
		serverHandler = gofarm.ProvisionServer()
	}
}

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

func RequestLock(rqPtr *LockerRequest) (chan bool, []error) {
	// Sanitize request
	sanitizationErrors := rqPtr.checkAndPrepareRequest()
	if len(sanitizationErrors) != 0 {
		return nil, sanitizationErrors
	}

	// Make request to server
	nativeResponseChannel, err := serverHandler.MakeRequest(rqPtr)
	if err != nil {
		return nil, []error{err}
	}

	// Pass through result channel
	nativeResponse, ok := <-nativeResponseChannel
	responseChannel := (*nativeResponse).(chan bool)
	if !ok {
		close(responseChannel)
	}
	return responseChannel, nil
}
