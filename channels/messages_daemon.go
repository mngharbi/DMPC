package channels

import (
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Server definitions
*/

type MessagesServerConfig struct {
	NumWorkers int
}

type messagesServer struct {
	isInitialized bool
}

var (
	messagesServerSingleton messagesServer
	messagesServerHandler   *gofarm.ServerHandler
	bufferStore             *memstore.Memstore
)

/*
	Server API
*/

func provisionMessagesServerOnce() {
	if messagesServerHandler == nil {
		messagesServerHandler = gofarm.ProvisionServer()
	}
}

func startMessagesServer(conf MessagesServerConfig, serversWaitGroup *sync.WaitGroup) (err error) {
	defer serversWaitGroup.Done()
	provisionMessagesServerOnce()
	if !messagesServerSingleton.isInitialized {
		messagesServerSingleton.isInitialized = true
		messagesServerHandler.ResetServer()
		messagesServerHandler.InitServer(&messagesServerSingleton)
	}
	return messagesServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func shutdownMessagesServer() {
	provisionMessagesServerOnce()
	messagesServerHandler.ShutdownServer()
}

/*
	Functional API
*/

/*
	Server implementation
*/

func (sv *messagesServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		bufferStore = memstore.New(getChannelBufferIndexes())
	}
	log.Debugf(messagesDaemonStartLogMsg)
	return nil
}

func (sv *messagesServer) Shutdown() error {
	log.Debugf(messagesDaemonShutdownLogMsg)
	return nil
}

func (sv *messagesServer) Work(rq *gofarm.Request) (dummyReturnVal *gofarm.Response) {
	log.Debugf(messagesRunningRequestLogMsg)

	dummyReturnVal = nil

	return
}
