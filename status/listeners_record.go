package status

import (
	"sync"
)

// Alias for status update channels
type UpdateChannel chan *StatusRecord

/*
	Structure of listening request
*/
type listeningRequest struct {
	ticket  Ticket
	channel UpdateChannel
}

/*
	Structure storing channels for tickets
	that don't have a status record yet
*/
type listenersRecord struct {
	id       Ticket
	lock     *sync.Mutex
	channels []UpdateChannel
}

/*
	Listener record locking
*/
func (rec *listenersRecord) Lock() {
	rec.lock.Lock()
}

func (rec *listenersRecord) Unlock() {
	rec.lock.Unlock()
}

// Dummies defined to implement core.RWRecordLocker
func (rec *listenersRecord) RLock()   {}
func (rec *listenersRecord) RUnlock() {}

// Index used to store listeners
const (
	listenersMemstoreId string = "id"
)

var listenersIndexesMap map[string]bool = map[string]bool{
	"id": true,
}

func getListenersIndexes() (res []string) {
	for k := range listenersIndexesMap {
		res = append(res, k)
	}
	return res
}

// Comparison function for pending listeners records (required for memstore)
func (rec *listenersRecord) Less(index string, than interface{}) bool {
	switch index {
	case "id":
		return rec.id < than.(*listenersRecord).id
	}
	return false
}

func makeEmptyListenersRecord(id Ticket) *listenersRecord {
	return &listenersRecord{
		id:       id,
		channels: []UpdateChannel{},
		lock:     &sync.Mutex{},
	}
}
