package status

import (
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
)

/*
	Server API
*/

type ListenersServerConfig struct {
	NumWorkers int
}

func provisionListenersServerOnce() {
	if listenersServerHandler == nil {
		listenersServerHandler = gofarm.ProvisionServer()
	}
}

func StartListenersServer(conf ListenersServerConfig) error {
	provisionListenersServerOnce()
	if !listenersServerSingleton.isInitialized {
		listenersServerSingleton.isInitialized = true
		listenersServerHandler.ResetServer()
		listenersServerHandler.InitServer(&listenersServerSingleton)
	}
	return listenersServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownListenersServer() {
	provisionListenersServerOnce()
	listenersServerHandler.ShutdownServer()
}

/*
	Server implementation
*/

type listenersServer struct {
	isInitialized bool
}

var (
	listenersServerSingleton listenersServer
	listenersServerHandler   *gofarm.ServerHandler
	listenersStore           *memstore.Memstore
)

func (sv *listenersServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		listenersStore = memstore.New(getListenersIndexes())
	}
	return nil
}

func (sv *listenersServer) Shutdown() error { return nil }

func (sv *listenersServer) Work(rq *gofarm.Request) *gofarm.Response {
	return nil
}
