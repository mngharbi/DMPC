/*
	Locking/Unlocking utilities to prevent deadlocks
*/

package core

import (
	"github.com/mngharbi/memstore"
	"sort"
)

/*
	Lock type alias
	false = read lock
	true = write lock
*/
type LockType bool

const ReadLockType LockType = false
const WriteLockType LockType = true

/*
	Locking type alias
	false = unlocking
	true = locking
*/
type LockingType bool

const Unlocking LockingType = false
const Locking LockingType = true

/*
	Lock need definition and sanitization
*/
type LockNeed struct {
	LockType LockType
	Id       string
}

type lockNeedCollection []LockNeed

func (s lockNeedCollection) Len() int           { return len(s) }
func (s lockNeedCollection) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s lockNeedCollection) Less(i, j int) bool { return s[i].Less(s[j]) }

// Order: read low id < read high id < write low id < write high id
func (a LockNeed) Less(b LockNeed) bool {
	if a.LockType != b.LockType {
		return a.LockType == ReadLockType
	}
	return a.Id < b.Id
}

func sanitizeLockNeeds(lockNeeds []LockNeed) []LockNeed {
	/*
		Remove duplicates and keep higher privilege lock
		0 (default): no locks
		1: read lock
		2: write lock
	*/
	lockIntMapping := map[LockType]int{
		ReadLockType:  1,
		WriteLockType: 2,
	}

	lockMap := map[string]int{}
	for _, lockNeed := range lockNeeds {
		var lockNeedMapping int = lockIntMapping[lockNeed.LockType]

		// Higher integer overwrites previous (no lock < read lock < write lock)
		if lockNeedMapping > lockMap[lockNeed.Id] {
			lockMap[lockNeed.Id] = lockNeedMapping
		}
	}

	// Build sanitized lock needs array to prevent deadlocks
	var orderedWriteLocks []LockNeed
	for lockId, lockTypeInt := range lockMap {
		lockType := ReadLockType
		if lockTypeInt == lockIntMapping[WriteLockType] {
			lockType = WriteLockType
		}
		orderedWriteLocks = append(orderedWriteLocks, LockNeed{lockType, lockId})
	}
	sort.Sort(lockNeedCollection(orderedWriteLocks))

	return orderedWriteLocks
}

func sanitizeUnlockNeeds(lockNeeds []LockNeed) []LockNeed {
	sanitizedLockNeeds := sanitizeLockNeeds(lockNeeds)

	// Reverse for unlocking
	for i := 0; i < len(sanitizedLockNeeds)/2; i++ {
		oppositeI := len(sanitizedLockNeeds) - i - 1
		sanitizedLockNeeds[i], sanitizedLockNeeds[oppositeI] = sanitizedLockNeeds[oppositeI], sanitizedLockNeeds[i]
	}

	return sanitizedLockNeeds
}

/*
	Generic lock operation on locker
*/

func rwLock(rwlocker RWLocker, lockType LockType, lockingType LockingType) {
	if lockType == WriteLockType {
		if lockingType == Locking {
			rwlocker.Lock()
		} else {
			rwlocker.Unlock()
		}
	} else {
		if lockingType == Locking {
			rwlocker.RLock()
		} else {
			rwlocker.RUnlock()
		}
	}
}

/*
	Generic do locking function
*/
func doLocking(reader RecordReader, lockNeeds []LockNeed, lockingType LockingType) bool {
	// Sanitize lock needs to avoid deadlocks
	var sanitizedLockNeeds []LockNeed
	if lockingType == Locking {
		sanitizedLockNeeds = sanitizeLockNeeds(lockNeeds)
	} else {
		sanitizedLockNeeds = sanitizeUnlockNeeds(lockNeeds)
	}

	// Read records
	var recordIds []string
	for _, lockNeed := range sanitizedLockNeeds {
		recordIds = append(recordIds, lockNeed.Id)
	}
	records := reader(recordIds)

	// Check if any are nil
	for _, record := range records {
		if record == nil {
			return false
		}
	}

	// Perform locking
	for lockNeedIndex, lockNeed := range sanitizedLockNeeds {
		rwLock(records[lockNeedIndex].(RWLocker), lockNeed.LockType, lockingType)
	}

	return true
}

/*
	Generic Locking function
*/

func Lock(reader RecordReader, lockNeeds []LockNeed) bool {
	return doLocking(reader, lockNeeds, Locking)
}

/*
	Generic Unlocking function
*/

func Unlock(reader RecordReader, lockNeeds []LockNeed) bool {
	return doLocking(reader, lockNeeds, Unlocking)
}

/*
	Defines an object that can be (R)locked/(R)unlocked
*/
type RWLocker interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

/*
	Generates a function that can be used to read a set of records
*/

type RecordReader func([]string) []memstore.Item

func RecordReaderFunctor(
	mem *memstore.Memstore,
	makeSearchRecordFunctor func(string) memstore.Item,
	indexString string,
	save bool,
	collectionPtr *[]memstore.Item,
) RecordReader {
	return func(ids []string) []memstore.Item {
		var searchRecords []memstore.Item
		for _, id := range ids {
			searchRecords = append(searchRecords, makeSearchRecordFunctor(id))
		}
		// @TODO: Implement read subset in memstore and use it
		memstoreItems := mem.ApplyDataSubset(searchRecords, indexString, func(item memstore.Item) bool {
			return true
		})
		if save {
			*collectionPtr = memstoreItems
		}
		return memstoreItems
	}
}
