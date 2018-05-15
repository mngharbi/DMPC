package channels

import (
	"sync"
)

// Alias for a channel of messages
type MessageChannel chan Message

/*
	Structure storing message channels
*/
type listenersRecord struct {
	id       string
	lock     *sync.Mutex
	channels []MessageChannel
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

/*
	Utilities
*/
func makeEmptyListenersRecord(id string) *listenersRecord {
	return &listenersRecord{
		id:       id,
		channels: []MessageChannel{},
		lock:     &sync.Mutex{},
	}
}
