package status

import (
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
)

/*
	Server API
*/

type StatusServerConfig struct {
	NumWorkers int
}

func provisionStatusServerOnce() {
	if statusServerHandler == nil {
		statusServerHandler = gofarm.ProvisionServer()
	}
}

func startStatusServer(conf StatusServerConfig) (err error) {
	provisionStatusServerOnce()
	if !statusServerSingleton.isInitialized {
		statusServerSingleton.isInitialized = true
		statusServerHandler.ResetServer()
		statusServerHandler.InitServer(&statusServerSingleton)
	}
	err = statusServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
	serversStartWaitGroup.Done()
	return
}

func shutdownStatusServer() {
	provisionStatusServerOnce()
	statusServerHandler.ShutdownServer()
}

func UpdateStatus(ticket Ticket, status StatusCode, failReason FailReasonCode, payload []byte, errs []error) error {
	statusRecord := &StatusRecord{
		Id:         ticket,
		Status:     status,
		FailReason: failReason,
		Payload:    payload,
		Errs:       errs,
	}

	// Check record
	if err := statusRecord.check(); err != nil {
		return err
	}

	// Make request to server
	if _, err := statusServerHandler.MakeRequest(statusRecord); err != nil {
		return err
	}

	return nil
}

/*
	Server implementation
*/

type statusServer struct {
	isInitialized bool
}

var (
	statusServerSingleton statusServer
	statusServerHandler   *gofarm.ServerHandler
	statusStore           *memstore.Memstore
)

func (sv *statusServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		statusStore = memstore.New(getStatusIndexes())
	}
	return nil
}

func (sv *statusServer) Shutdown() error { return nil }

func doStatusUpdate(currentRecord *StatusRecord, changedRecord *StatusRecord) {
	// Update record
	recordChanged := currentRecord.update(changedRecord)
	if !recordChanged {
		return
	}

	/*
		Get listeners record

		Note: listeners record is implicitly locked
		because adding listeners takes a read lock on the status record
	*/
	listenersRecordItem := listenersStore.Get(makeEmptyListenersRecord(currentRecord.Id), statusMemstoreId)
	if listenersRecordItem == nil {
		return
	}
	listenersRecord := listenersRecordItem.(*listenersRecord)

	// Send update to all listners
	for _, updateChannel := range listenersRecord.channels {
		updateChannel <- currentRecord
	}

	// If final update, close all listener channels and delete listener record
	if currentRecord.isDone() {
		for _, updateChannel := range listenersRecord.channels {
			close(updateChannel)
		}
		listenersRecord.channels = nil
		listenersRecord.lock = nil
		statusStore.Delete(currentRecord, statusMemstoreId)
	}
}

func (sv *statusServer) Work(rq *gofarm.Request) (dummyReturnVal *gofarm.Response) {
	dummyReturnVal = nil

	changedRecord := (*rq).(*StatusRecord)

	// Read/Create and write lock status record
	currentRecord := changedRecord.createOrGet(statusStore)
	currentRecord.Lock()

	// Read status record again (avoids race conditions)
	currentRecord = statusStore.Get(currentRecord, statusMemstoreId).(*StatusRecord)

	doStatusUpdate(currentRecord, changedRecord)

	currentRecord.Unlock()

	return
}
