package status

import (
	"errors"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Errors
*/
var (
	statusRangeError error = errors.New("Status code is out of bounds.")
	failedRangeError error = errors.New("Failed status code is out of bounds.")
)

/*
	Status codes
*/
type StatusCode int

const (
	QueuedStatus = iota
	RunningStatus
	SuccessStatus
	FailedStatus
)

/*
	FailReason codes
*/
type FailReasonCode int

const (
	NoReason = iota
	RejectedReason
	FailedReason
)

/*
	Structure of a status record
*/
type StatusRecord struct {
	Id         Ticket
	Status     StatusCode
	FailReason FailReasonCode
	Payload    []byte
	Errs       []error
	lock       *sync.RWMutex
}

/*
	Record locking
*/
func (rec *StatusRecord) Lock() {
	rec.lock.Lock()
}

func (rec *StatusRecord) RLock() {
	rec.lock.RLock()
}

func (rec *StatusRecord) Unlock() {
	rec.lock.Unlock()
}

func (rec *StatusRecord) RUnlock() {
	rec.lock.RUnlock()
}

// Index used to store status
const (
	statusMemstoreId string = "id"
)

var statusIndexesMap map[string]bool = map[string]bool{
	statusMemstoreId: true,
}

func getStatusIndexes() (res []string) {
	for k := range statusIndexesMap {
		res = append(res, k)
	}
	return res
}

// Comparison function for status records (required for memstore)
func (rec *StatusRecord) Less(index string, than interface{}) bool {
	switch index {
	case statusMemstoreId:
		return rec.Id < than.(*StatusRecord).Id
	}
	return false
}

/*
	Utilities
*/
func (rec *StatusRecord) check() error {
	// Check status bounds
	if !(QueuedStatus <= rec.Status && rec.Status <= FailedStatus) {
		return statusRangeError
	}

	// Check fail reasons bounds
	if !(NoReason <= rec.FailReason && rec.FailReason <= FailedReason) {
		return failedRangeError
	}

	return nil
}

func (current *StatusRecord) update(updated *StatusRecord) bool {
	// Don't apply any stale updates
	if current.Status >= updated.Status {
		return false
	}

	current.Status = updated.Status
	current.FailReason = updated.FailReason
	current.Payload = updated.Payload
	current.Errs = updated.Errs
	return true
}

func (rec *StatusRecord) createOrGet(mem *memstore.Memstore) *StatusRecord {
	rec.lock = &sync.RWMutex{}
	return mem.AddOrGet(rec).(*StatusRecord)
}

func (rec *StatusRecord) isDone() bool {
	return rec.Status >= SuccessStatus
}

func makeStatusEmptyRecord(id Ticket) *StatusRecord {
	return &StatusRecord{
		Id: id,
	}
}
