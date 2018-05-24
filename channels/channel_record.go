package channels

import (
	"sort"
	"sync"
	"time"
)

/*
	Channel permissions record
*/
type channelPermissionRecord struct {
	read  bool
	write bool
	close bool
}
type channelPermissionsRecord struct {
	users map[string]*channelPermissionRecord
}

/*
	Channel opening record
*/
type channelActionCollection []*channelActionRecord

func (s channelActionCollection) Len() int           { return len(s) }
func (s channelActionCollection) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s channelActionCollection) Less(i, j int) bool { return s[i].timestamp.Before(s[j].timestamp) }

type channelActionRecord struct {
	issuerId    string
	certifierId string
	timestamp   time.Time
}

/*
	Channel duration record
*/
type channelDurationRecord struct {
	opened time.Time
	closed time.Time
}

/*
	Channel state
*/
type channelState int

const (
	channelBufferedState channelState = iota
	channelOpenState
	channelClosedState
	channelInconsistentState
)

/*
	Structure of a channel
*/
type channelRecord struct {
	// Identifiers
	id    string
	keyId string

	// Properties
	duration        *channelDurationRecord
	permissions     *channelPermissionsRecord
	opening         *channelActionRecord
	closure         *channelActionRecord
	closureAttempts channelActionCollection

	// State management
	state channelState
	lock  *sync.RWMutex
}

/*
	Open channel action
	Note: does not include verifying global permissions
*/
func (rec *channelRecord) tryOpen(id string, opening *channelActionRecord, permissions *channelPermissionsRecord) bool {
	if rec.state != channelBufferedState ||
		opening == nil ||
		opening.timestamp.IsZero() ||
		permissions == nil ||
		len(permissions.users) == 0 {
		return false
	}

	// Mark as open
	rec.id = id
	rec.opening = opening
	rec.permissions = permissions
	rec.duration = &channelDurationRecord{
		opened: opening.timestamp,
	}
	rec.state = channelOpenState
	return true
}

/*
	Close channel action
	Note: does not include verifying global permissions
*/
func (rec *channelRecord) tryClose(closure *channelActionRecord) bool {
	if rec.state == channelClosedState ||
		closure == nil ||
		closure.timestamp.IsZero() {
		return false
	}

	// Buffer closure if channel is still buffered
	if rec.state == channelBufferedState {
		rec.closureAttempts = append(rec.closureAttempts, closure)
		return true
	}

	// Closure should be after opening
	if rec.duration == nil || rec.duration.opened.After(closure.timestamp) {
		return false
	}

	// Determine if we can close
	var canClose bool = false
	if permissionRecord, ok := rec.permissions.users[closure.certifierId]; ok && permissionRecord != nil {
		canClose = permissionRecord.close
	}
	if !canClose {
		return false
	}

	// Mark as closed
	rec.closure = closure
	rec.closureAttempts = nil
	rec.duration.closed = closure.timestamp
	rec.state = channelClosedState
	return true
}

/*
	Applying close attempts
	Note: does not include verifying global permissions
*/
func (rec *channelRecord) applyCloseAttempts() bool {
	if rec.state != channelOpenState ||
		len(rec.closureAttempts) == 0 {
		return false
	}

	// Sort attempts by by closure dates
	sort.Sort(rec.closureAttempts)

	// Call close on every closure action
	defer func() { rec.closureAttempts = nil }()
	for _, attempt := range rec.closureAttempts {
		if rec.tryClose(attempt) {
			return true
		}
	}

	return false
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
