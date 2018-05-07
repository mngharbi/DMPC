package pipeline

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/gofarm"
	"sync"
)

/*
	Shared variables
*/
var (
	log             *core.LoggingHandler
	serverLock      *sync.RWMutex = &sync.RWMutex{}
	serverSingleton *server       = &server{}
)

/*
	Used to pass operation to decryptor
*/
func passOperation(operation *core.TemporaryEncryptedOperation) (channel chan *gofarm.Response, errs []error) {
	serverLock.RLock()
	defer serverLock.RUnlock()
	if serverSingleton.isRunning {
		channel, errs = serverSingleton.requester(operation)
	}
	return
}

/*
	Server API
*/

func StartServer(config Config, requester decryptor.Requester, loggingHandler *core.LoggingHandler) {
	if log == nil {
		log = loggingHandler
	}
	serverLock.Lock()
	serverSingleton.start(config, requester)
	serverLock.Unlock()
}

func ShutdownServer() {
	serverLock.Lock()
	serverSingleton.shutdown()
	serverLock.Unlock()
}
