package channels

import (
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Server definitions
*/

type ListenersServerConfig struct {
	NumWorkers int
}

type listenersServer struct {
	isInitialized bool
}

var (
	listenersServerSingleton listenersServer
	listenersServerHandler   *gofarm.ServerHandler
	listenersStore           *memstore.Memstore
)

/*
	Server API
*/

func provisionListenersServerOnce() {
	if listenersServerHandler == nil {
		listenersServerHandler = gofarm.ProvisionServer()
	}
}

func startListenersServer(conf ListenersServerConfig, serversWaitGroup *sync.WaitGroup) (err error) {
	defer serversWaitGroup.Done()
	provisionListenersServerOnce()
	if !listenersServerSingleton.isInitialized {
		listenersServerSingleton.isInitialized = true
		listenersServerHandler.ResetServer()
		listenersServerHandler.InitServer(&listenersServerSingleton)
	}
	return listenersServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func shutdownListenersServer() {
	provisionListenersServerOnce()
	listenersServerHandler.ShutdownServer()
}

/*
	Functional API
*/

/*
	Server implementation
*/

func (sv *listenersServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		listenersStore = memstore.New(getListenersIndexes())
	}
	log.Debugf(listenersDaemonStartLogMsg)
	return nil
}

func (sv *listenersServer) Shutdown() error {
	log.Debugf(listenersDaemonShutdownLogMsg)
	return nil
}

func (sv *listenersServer) Work(rq *gofarm.Request) (dummyReturnVal *gofarm.Response) {
	log.Debugf(listenersRunningRequestLogMsg)

	dummyReturnVal = nil

	return
}
