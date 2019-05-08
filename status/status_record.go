package status

import (
	"encoding/json"
	"errors"
	"github.com/mngharbi/memstore"
	"reflect"
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
type StatusCode string

const (
	NoStatus      StatusCode = "no_status"
	QueuedStatus  StatusCode = "queued"
	RunningStatus StatusCode = "running"
	SuccessStatus StatusCode = "success"
	FailedStatus  StatusCode = "failed"
)

var (
	statusCodes [5]StatusCode = [5]StatusCode{
		NoStatus,
		QueuedStatus,
		RunningStatus,
		SuccessStatus,
		FailedStatus,
	}
	statusCodesOrder map[StatusCode]int = map[StatusCode]int{
		NoStatus:      0,
		QueuedStatus:  1,
		RunningStatus: 2,
		SuccessStatus: 3,
		FailedStatus:  4,
	}
)

/*
	FailReason codes
*/
type FailReasonCode string

const (
	NoReason       FailReasonCode = "no_failure"
	RejectedReason FailReasonCode = "rejected"
	FailedReason   FailReasonCode = "failed"
)

var (
	failureReasons [5]FailReasonCode = [5]FailReasonCode{
		NoReason,
		RejectedReason,
		FailedReason,
	}
	failureReasonsMap map[FailReasonCode]bool = map[FailReasonCode]bool{
		NoReason:       true,
		RejectedReason: true,
		FailedReason:   true,
	}
)

/*
	Structure of a status record
*/
type StatusRecord struct {
	Id         Ticket         `json:"ticket"`
	Status     StatusCode     `json:"status"`
	FailReason FailReasonCode `json:"fail_status"`
	Payload    interface{}
	Errs       []error `json:"errors"`
	lock       *sync.RWMutex
}

/*
	Encoding
*/

type Encodable interface {
	Encode() ([]byte, error)
}

type ChannelResponse interface {
	GetResponse() ([]byte, bool)
	GetSubscriberId() string
	GetChannelId() string
}

// *StatusRecord -> Json
func (rec *StatusRecord) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(rec)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

func (rec *StatusRecord) GetResponse() ([]byte, bool) {
	if !rec.IsDone() {
		encoded, _ := rec.Encode()
		return encoded, true
	}
	if rec.Status == FailedStatus {
		encoded, _ := rec.Encode()
		return encoded, false
	}
	switch rec.Payload.(type) {
	case Encodable:
		encoded, _ := rec.Payload.(Encodable).Encode()
		return encoded, false
	case []byte:
		return rec.Payload.([]byte), false
	case ChannelResponse:
		return rec.Payload.(ChannelResponse).GetResponse()
	default:
		return nil, false
	}
}

func (rec *StatusRecord) GetSubscriberId() (string, bool) {
	if rec.Status != SuccessStatus {
		return "", false
	}

	switch rec.Payload.(type) {
	case ChannelResponse:
		return rec.Payload.(ChannelResponse).GetSubscriberId(), true
	default:
		return "", false
	}
}

func (rec *StatusRecord) GetChannelId() (string, bool) {
	if rec.Status != SuccessStatus {
		return "", false
	}

	switch rec.Payload.(type) {
	case ChannelResponse:
		return rec.Payload.(ChannelResponse).GetChannelId(), true
	default:
		return "", false
	}
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
	if _, ok := statusCodesOrder[rec.Status]; !ok {
		return statusRangeError
	}

	// Check fail reasons bounds
	if _, ok := failureReasonsMap[rec.FailReason]; !ok {
		return failedRangeError
	}

	return nil
}

func (current *StatusRecord) update(updated *StatusRecord) bool {
	// Don't apply any stale updates
	if statusCodesOrder[current.Status] >= statusCodesOrder[updated.Status] {
		return false
	}

	current.Status = updated.Status
	current.FailReason = updated.FailReason
	current.Payload = updated.Payload
	current.Errs = updated.Errs
	return true
}

func (a *StatusRecord) isSame(b *StatusRecord) bool {
	return a.Id == b.Id &&
		a.Status == b.Status &&
		a.FailReason == b.FailReason &&
		reflect.DeepEqual(a.Payload, b.Payload) &&
		reflect.DeepEqual(a.Errs, b.Errs)
}

func (rec *StatusRecord) createOrGet(mem *memstore.Memstore) *StatusRecord {
	rec.lock = &sync.RWMutex{}
	return mem.AddOrGet(rec).(*StatusRecord)
}

func (rec *StatusRecord) IsDone() bool {
	return rec.Status == SuccessStatus || rec.Status == FailedStatus
}

func makeStatusEmptyRecord(id Ticket) *StatusRecord {
	return &StatusRecord{
		Id:         id,
		Status:     NoStatus,
		FailReason: NoReason,
	}
}
