package channels

import (
	"github.com/mngharbi/DMPC/core"
	"sync"
)

/*
	Structure of a channel buffer
*/
type channelBufferRecord struct {
	id         string
	operations []*core.Operation
	lock       *sync.Mutex
}

/*
	Comparison
*/
func (rec *channelBufferRecord) Less(index string, than interface{}) bool {
	switch index {
	case channelBufferIndexId:
		return rec.id < than.(*channelBufferRecord).id
	}
	return false
}

/*
	Channel buffer record locking
*/
func (rec *channelBufferRecord) Lock() {
	rec.lock.Lock()
}

func (rec *channelBufferRecord) Unlock() {
	rec.lock.Unlock()
}
func (rec *channelBufferRecord) RLock()   {}
func (rec *channelBufferRecord) RUnlock() {}

/*
	Indexing
*/
const (
	channelBufferIndexId string = "id"
)

var channelBufferIndexesMap map[string]bool = map[string]bool{
	channelBufferIndexId: true,
}

func getChannelBufferIndexes() (res []string) {
	for k := range channelBufferIndexesMap {
		res = append(res, k)
	}
	return res
}

/*
	Utilities
*/
func makeEmptyChannelBufferRecord(id string) *channelBufferRecord {
	return &channelBufferRecord{
		id:         id,
		operations: []*core.Operation{},
		lock:       &sync.Mutex{},
	}
}
