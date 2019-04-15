package locker

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Server structure
*/

type server struct {
	isInitialized bool
	lockStores    [1]*memstore.Memstore
}

/*
	Server data
*/

// Indexes used to store resource locks for a specific type
var indexesMap map[string]bool = map[string]bool{
	"id": true,
}

// Server object
var serverSingleton server

/*
	Server implementation
*/
func (sv *server) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize stores (only if starting for the first time)
	if isFirstStart {
		for storeIndex := range sv.lockStores {
			sv.lockStores[storeIndex] = memstore.New(getIndexes())
		}
	}
	log.Debugf(daemonStartLogMsg)
	return nil
}

func (sv *server) Work(request *gofarm.Request) *gofarm.Response {
	log.Debugf(runningRequestLogMsg)

	rq := (*request).(*LockerRequest)

	// Determine memstore based on resource type requested
	targetMemstore := sv.lockStores[rq.Type]

	// Build channel and push result from locking into it in separate goroutine
	responseChannel := make(chan bool)
	go func() {
		if rq.LockingType == core.Locking {
			// Create and lock resource records if they don't exist
			// @TODO: include this in lockresources
			for _, need := range rq.Needs {
				newResourceRecord := &resourceRecord{
					Id:   need.Id,
					lock: &sync.RWMutex{},
				}
				targetMemstore.AddOrGet(newResourceRecord)
			}

			responseChannel <- lockResources(targetMemstore, rq.Needs)
		} else {
			// @TODO: clean up resource after unlocking
			responseChannel <- unlockResources(targetMemstore, rq.Needs)
		}
	}()

	// Log request done
	log.Debugf(doneRequestLogMsg)

	var nativeResp gofarm.Response = responseChannel
	return &nativeResp
}

func (sv *server) Shutdown() error {
	log.Debugf(daemonShutdownLogMsg)
	return nil
}

/*
	Server helpers
*/

func getIndexes() (res []string) {
	for k := range indexesMap {
		res = append(res, k)
	}
	return res
}
