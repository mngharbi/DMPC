package channels

import (
	"github.com/mngharbi/memstore"
	"sync"
)

// Alias for a channel of events
type EventChannel chan *Event

/*
	Structure storing message channels
*/
type listenersRecord struct {
	id       string
	lock     *sync.Mutex
	channels []EventChannel
}

/*
	Utilities
*/

func makeEmptyListenersRecord(id string) *listenersRecord {
	return &listenersRecord{
		id:       id,
		lock:     &sync.Mutex{},
		channels: []EventChannel{},
	}
}

func createOrGetListenersRecord(listenersStore *memstore.Memstore, id string) *listenersRecord {
	newRecord := makeEmptyListenersRecord(id)
	return listenersStore.AddOrGet(newRecord).(*listenersRecord)
}

func notifyListeners(listenersStore *memstore.Memstore, id string, event *Event) {
	// Create/Get records
	listenersRecord := createOrGetListenersRecord(listenersStore, id)

	// Lock then unlock at the end
	listenersRecord.Lock()
	defer func() { listenersRecord.Unlock() }()

	// Send event to all listeners
	for _, listenerChannel := range listenersRecord.channels {
		listenerChannel <- event
	}
}

func makeSearchListenersRecord(id string) *listenersRecord {
	return &listenersRecord{
		id: id,
	}
}

/*
	Comparison
*/
func (rec *listenersRecord) Less(index string, than interface{}) bool {
	switch index {
	case listenersIndexId:
		return rec.id < than.(*listenersRecord).id
	}
	return false
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
func (rec *listenersRecord) RLock()   {}
func (rec *listenersRecord) RUnlock() {}

/*
	Indexing
*/
const (
	listenersIndexId string = "id"
)

var listenersIndexesMap map[string]bool = map[string]bool{
	listenersIndexId: true,
}

func getListenersIndexes() (res []string) {
	for k := range listenersIndexesMap {
		res = append(res, k)
	}
	return res
}
