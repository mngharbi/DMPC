package channels

import (
	"sync"
)

/*
	Structure of a channel
*/
type channelRecord struct {
	id    string
	keyId string
	lock  *sync.RWMutex
}

/*
	Comparison
*/
func (rec *channelRecord) Less(index string, than interface{}) bool {
	switch index {
	case channelIndexId:
		return rec.id < than.(*channelRecord).id
	case channelIndexKeyId:
		return rec.keyId < than.(*channelRecord).keyId
	}
	return false
}

/*
	Channel record locking
*/
func (rec *channelRecord) Lock() {
	rec.lock.Lock()
}
func (rec *channelRecord) Unlock() {
	rec.lock.Unlock()
}
func (rec *channelRecord) RLock() {
	rec.lock.RLock()
}
func (rec *channelRecord) RUnlock() {
	rec.lock.RUnlock()
}

/*
	Indexing
*/
const (
	channelIndexId    string = "id"
	channelIndexKeyId string = "keyId"
)

var channelIndexesMap map[string]bool = map[string]bool{
	channelIndexId:    true,
	channelIndexKeyId: true,
}

func getChannelIndexes() (res []string) {
	for k := range channelIndexesMap {
		res = append(res, k)
	}
	return res
}
