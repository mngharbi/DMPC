package status

import (
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Server defaults
*/
const DefaultChannelBufferSize = 3

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

func StartListenersServer(conf ListenersServerConfig) (err error) {
	provisionListenersServerOnce()
	if !listenersServerSingleton.isInitialized {
		listenersServerSingleton.isInitialized = true
		listenersServerHandler.ResetServer()
		listenersServerHandler.InitServer(&listenersServerSingleton)
	}
	err = listenersServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
	serversStartWaitGroup.Done()
	return
}

func ShutdownListenersServer() {
	provisionListenersServerOnce()
	listenersServerHandler.ShutdownServer()
}

func AddListener(ticket Ticket) (UpdateChannel, error) {
	// Pass request to server to add it
	listeningRequest := &listeningRequest{
		ticket:  ticket,
		channel: make(UpdateChannel, DefaultChannelBufferSize),
	}

	_, err := listenersServerHandler.MakeRequest(listeningRequest)
	if err != nil {
		close(listeningRequest.channel)
		return listeningRequest.channel, err
	}

	return listeningRequest.channel, nil
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

func doListenerServerWork(statusRecord *StatusRecord, channel UpdateChannel) {
	// If status is done, we only need to put the last status
	if statusRecord.isDone() {
		channel <- statusRecord
		close(channel)
		return
	}

	// Read/Create and lock listeners record
	newListenersRecord := makeEmptyListenersRecord(statusRecord.Id)
	newListenersRecord.lock = &sync.Mutex{}
	listenersRecordObj := listenersStore.AddOrGet(newListenersRecord).(*listenersRecord)
	listenersRecordObj.Lock()

	// Read listeners record again
	listenersRecordObj = listenersStore.Get(listenersRecordObj, listenersMemstoreId).(*listenersRecord)

	// Add channel to listeners
	listenersRecordObj.channels = append(listenersRecordObj.channels, channel)

	listenersRecordObj.Unlock()
}

func (sv *listenersServer) Work(rq *gofarm.Request) (dummyReturnVal *gofarm.Response) {
	dummyReturnVal = nil
	listeningRequest := (*rq).(*listeningRequest)

	// Read/Create and read lock status record
	newStatusRecord := makeStatusEmptyRecord(listeningRequest.ticket)
	currentStatusRecord := newStatusRecord.createOrGet(statusStore)
	currentStatusRecord.RLock()

	// Read record again (avoids race conditions)
	currentStatusRecord = statusStore.Get(currentStatusRecord, statusMemstoreId).(*StatusRecord)

	doListenerServerWork(currentStatusRecord, listeningRequest.channel)

	currentStatusRecord.RUnlock()

	return
}
