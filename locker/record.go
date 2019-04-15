package locker

import (
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Record of a resource
*/
type resourceRecord struct {
	Id   string
	lock *sync.RWMutex
}

/*
	Resource ordering
*/
func (rec *resourceRecord) Less(index string, than interface{}) bool {
	switch index {
	case "id":
		return rec.Id < than.(*resourceRecord).Id
	}
	return false
}

/*
	Resource locking
*/

// Read lock
func (record *resourceRecord) RLock() {
	record.lock.RLock()
}

// Read unlock
func (record *resourceRecord) RUnlock() {
	record.lock.RUnlock()
}

// Write lock
func (record *resourceRecord) Lock() {
	record.lock.Lock()
}

// Write unlock
func (record *resourceRecord) Unlock() {
	record.lock.Unlock()
}

/*
	Utilities
*/

// Make a dummy resource record pointer for search from an id
func makeSearchByIdRecord(id string) memstore.Item {
	return &resourceRecord{
		Id: id,
	}
}
