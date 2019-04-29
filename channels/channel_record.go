package channels

import (
	"github.com/mngharbi/memstore"
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

func (rec *channelPermissionsRecord) build(obj *ChannelPermissionsObject) {
	rec.users = map[string]*channelPermissionRecord{}
	for userId, userPermissions := range obj.Users {
		rec.users[userId] = &channelPermissionRecord{
			read:  userPermissions.Read,
			write: userPermissions.Write,
			close: userPermissions.Close,
		}
	}
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

var objectStateMapping []ChannelObjectState = []ChannelObjectState{
	ChannelObjectBufferedState,
	ChannelObjectOpenState,
	ChannelObjectClosedState,
	ChannelObjectInconsistentState,
}

/*
	Structure of a channel
*/
type channelRecord struct {
	// Identifiers
	id string

	// Properties
	duration        *channelDurationRecord
	permissions     *channelPermissionsRecord
	opening         *channelActionRecord
	closure         *channelActionRecord
	closureAttempts channelActionCollection
	keyId           string

	// Message timestamps (to determine order)
	// @TODO: Use a tree for O(log n) add message
	messageTimestamps []time.Time

	// State management
	state channelState
	lock  *sync.RWMutex
}

/*
	Utilities
*/

func makeEmptyChannelRecord(id string) *channelRecord {
	return &channelRecord{
		id:    id,
		lock:  &sync.RWMutex{},
		state: channelBufferedState,
	}
}

func getChannel(channelsStore *memstore.Memstore, id string) *channelRecord {
	newRecord := makeEmptyChannelRecord(id)
	channelRec := channelsStore.Get(newRecord, channelIndexId)
	if channelRec != nil {
		return channelRec.(*channelRecord)
	} else {
		return nil
	}
}

func createOrGetChannel(channelsStore *memstore.Memstore, id string) *channelRecord {
	newRecord := makeEmptyChannelRecord(id)
	return channelsStore.AddOrGet(newRecord).(*channelRecord)
}

/*
	Open channel action
	Note: does not include verifying global permissions
*/
func (rec *channelRecord) tryOpen(id string, opening *channelActionRecord, permissions *channelPermissionsRecord, keyId string) bool {
	if rec.state != channelBufferedState ||
		opening == nil ||
		opening.timestamp.IsZero() ||
		permissions == nil ||
		len(permissions.users) == 0 ||
		len(keyId) == 0 {
		return false
	}

	// Mark as open
	rec.id = id
	rec.keyId = keyId
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
	Returns number of messages left in the channel
	Note: does not include verifying global permissions
	@TODO: Handle cutting out messages that fall outside duration
*/
func (rec *channelRecord) tryClose(closure *channelActionRecord) (int, bool) {
	if rec.state == channelInconsistentState ||
		closure == nil ||
		closure.timestamp.IsZero() {
		return 0, false
	}

	// Buffer closure if channel is still buffered
	if rec.state == channelBufferedState {
		rec.closureAttempts = append(rec.closureAttempts, closure)
		return 0, true
	}

	// Closure should be after opening
	if rec.duration == nil || rec.duration.opened.After(closure.timestamp) {
		return 0, false
	}

	// Determine if we can close
	var canClose bool = false
	if permissionRecord, ok := rec.permissions.users[closure.certifierId]; ok && permissionRecord != nil {
		canClose = permissionRecord.close
	}
	if !canClose {
		return 0, false
	}

	// Determine if there were earlier closure attempts (if channel is closed already)
	if rec.state == channelClosedState && rec.duration.closed.Before(closure.timestamp) {
		return 0, false
	}

	// If timestamp is equal, rely on issuer id as tiebreaker
	if rec.state == channelClosedState &&
		rec.duration.closed.Equal(closure.timestamp) &&
		rec.closure.issuerId <= closure.issuerId {
		return 0, false
	}

	// Mark as closed
	rec.closure = closure
	rec.closureAttempts = nil
	rec.duration.closed = closure.timestamp
	rec.state = channelClosedState

	// Remove late messages
	filteredMessageTimestamps := []time.Time{}
	for _, messageTimestamp := range rec.messageTimestamps {
		if messageTimestamp.Before(rec.duration.closed) {
			filteredMessageTimestamps = append(filteredMessageTimestamps, messageTimestamp)
		}
	}
	rec.messageTimestamps = filteredMessageTimestamps

	return len(filteredMessageTimestamps), true
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
		if _, ok := rec.tryClose(attempt); ok {
			return true
		}
	}

	return false
}

/*
	Add message to channel
	Returns position of message
	@TODO verify duration and permissions
	@TODO: Switch to vector clock -> timestamp -> message hash for sorting
*/
func (rec *channelRecord) addMessage(addMessageAction *channelActionRecord) (int, bool) {
	if rec.state == channelBufferedState ||
		rec.state == channelInconsistentState ||
		rec.duration.opened.After(addMessageAction.timestamp) ||
		(rec.state == channelClosedState && rec.duration.closed.Before(addMessageAction.timestamp)) ||
		addMessageAction.timestamp.IsZero() {
		return 0, false
	}

	// Determine if we can write
	var canWrite bool = false
	if permissionRecord, ok := rec.permissions.users[addMessageAction.certifierId]; ok && permissionRecord != nil {
		canWrite = permissionRecord.write
	}
	if !canWrite {
		return 0, false
	}

	messagePosition := 0
	for _, timestamp := range rec.messageTimestamps {
		if addMessageAction.timestamp.After(timestamp) {
			messagePosition++
		}
	}
	rec.messageTimestamps = append(rec.messageTimestamps, addMessageAction.timestamp)
	return messagePosition, true
}

/*
	Comparison
*/
func (rec *channelRecord) Less(index string, than interface{}) bool {
	switch index {
	case channelIndexId:
		return rec.id < than.(*channelRecord).id
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
	channelIndexId string = "id"
)

var channelIndexesMap map[string]bool = map[string]bool{
	channelIndexId: true,
}

func getChannelIndexes() (res []string) {
	for k := range channelIndexesMap {
		res = append(res, k)
	}
	return res
}
