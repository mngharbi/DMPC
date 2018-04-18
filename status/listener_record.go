package status

import (
	"sync"
)

// Alias for status update channels
type UpdateChannel chan *StatusRecord

/*
	Structure storing channels for tickets
	that don't have a status record yet
*/
type pendingListenersRecord struct {
	id       Ticket
	lock     *sync.RWMutex
	channels []UpdateChannel
}

/*
	Listener record locking
*/
func (rec *pendingListenersRecord) Lock() {
	rec.lock.Lock()
}

func (rec *pendingListenersRecord) RLock() {
	rec.lock.RLock()
}

func (rec *pendingListenersRecord) Unlock() {
	rec.lock.Unlock()
}

func (rec *pendingListenersRecord) RUnlock() {
	rec.lock.RUnlock()
}

// Indexes used to store pending listeners records
var pendingListenersIndexesMap map[string]bool = map[string]bool{
	"id": true,
}

func getPendingListenersIndexes() (res []string) {
	for k := range pendingListenersIndexesMap {
		res = append(res, k)
	}
	return res
}

// Comparison function for pending listeners records (required for memstore)
func (rec *pendingListenersRecord) Less(index string, than interface{}) bool {
	switch index {
	case "id":
		return rec.id < than.(*pendingListenersRecord).id
	}
	return false
}

func makeListenerEmptyRecord(id Ticket) *pendingListenersRecord {
	return &pendingListenersRecord{
		id:       id,
		lock:     &sync.RWMutex{},
		channels: []UpdateChannel{},
	}
}

func makeListenerSearchRecord(id Ticket) *pendingListenersRecord {
	return &pendingListenersRecord{
		id: id,
	}
}
