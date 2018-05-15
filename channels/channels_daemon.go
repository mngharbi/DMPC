package channels

import (
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Server definitions
*/

type ChannelsServerConfig struct {
	NumWorkers int
}

type channelsServer struct {
	isInitialized bool
}

var (
	channelsServerSingleton channelsServer
	channelsServerHandler   *gofarm.ServerHandler
	channelsStore           *memstore.Memstore
)

/*
	Server API
*/

func provisionChannelsServerOnce() {
	if channelsServerHandler == nil {
		channelsServerHandler = gofarm.ProvisionServer()
	}
}

func startChannelsServer(conf ChannelsServerConfig, serversWaitGroup *sync.WaitGroup) (err error) {
	defer serversWaitGroup.Done()
	provisionChannelsServerOnce()
	if !channelsServerSingleton.isInitialized {
		channelsServerSingleton.isInitialized = true
		channelsServerHandler.ResetServer()
		channelsServerHandler.InitServer(&channelsServerSingleton)
	}
	return channelsServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func shutdownChannelsServer() {
	provisionChannelsServerOnce()
	channelsServerHandler.ShutdownServer()
}

/*
	Functional API
*/

/*
	Server implementation
*/

func (sv *channelsServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		channelsStore = memstore.New(getChannelIndexes())
	}
	log.Debugf(channelsDaemonStartLogMsg)
	return nil
}

func (sv *channelsServer) Shutdown() error {
	log.Debugf(channelsDaemonShutdownLogMsg)
	return nil
}

func (sv *channelsServer) Work(rq *gofarm.Request) (dummyReturnVal *gofarm.Response) {
	log.Debugf(channelsRunningRequestLogMsg)

	dummyReturnVal = nil

	return
}
