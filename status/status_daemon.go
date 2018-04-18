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

func StartStatusServer(conf StatusServerConfig) error {
	provisionStatusServerOnce()
	if !statusServerSingleton.isInitialized {
		statusServerSingleton.isInitialized = true
		statusServerHandler.ResetServer()
		statusServerHandler.InitServer(&statusServerSingleton)
	}
	return statusServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func ShutdownStatusServer() {
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
	err := statusRecord.checkAndSanitize()
	if err != nil {
		return err
	}

	// Make request to server
	_, err = statusServerHandler.MakeRequest(statusRecord)
	if err != nil {
		return err
	}

	return nil
}

/*
	Server implementation
*/

type statusServer struct {
	isInitialized bool
	store         *memstore.Memstore
}

var (
	statusServerSingleton statusServer
	statusServerHandler   *gofarm.ServerHandler
)

func (sv *statusServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		sv.store = memstore.New(getStatusIndexes())
	}
	return nil
}

func (sv *statusServer) Shutdown() error { return nil }

func (sv *statusServer) Work(_ *gofarm.Request) *gofarm.Response { return nil }
