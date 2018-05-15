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
	Generic Locking function
*/
func Lock(doLock func(string, LockType) bool, doUnlock func(string, LockType) bool, lockNeeds []LockNeed) bool {
	// Sanitize lock needs to avoid deadlocks
	sanitizedLockNeeds := sanitizeLockNeeds(lockNeeds)

	// Perform locking
	needToUnlock := false
	var successfulLocks []LockNeed
	for _, lockNeed := range sanitizedLockNeeds {
		if !doLock(lockNeed.Id, lockNeed.LockType) {
			needToUnlock = true
		} else {
			successfulLocks = append(successfulLocks, lockNeed)
		}
	}

	// If none failed we're done
	if !needToUnlock {
		return true
	}

	// Otherwise, unlock everything locked (in reverse)
	for _, lockNeed := range successfulLocks {
		doUnlock(lockNeed.Id, lockNeed.LockType)
	}

	return false
}

/*
	Generic Unlocking function
*/
func Unlock(doUnlock func(string, LockType) bool, unlockNeeds []LockNeed) bool {
	// Sanitize unlock needs to avoid deadlocks
	sanitizedUnlockNeeds := sanitizeUnlockNeeds(unlockNeeds)

	// Perform unlocking
	unlockSuccess := true
	for _, lockNeed := range sanitizedUnlockNeeds {
		if !doUnlock(lockNeed.Id, lockNeed.LockType) {
			unlockSuccess = false
		}
	}

	return unlockSuccess
}

/*
	Defines an object that can be (R)locked/(R)unlocked
*/
type RWRecordLocker interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

/*
	Generates a functor that (r)locks/(r)unlocks RWRecordLocker
*/
func lockingFunctorGenerator(lockType LockType, lockingType LockingType) func(memstore.Item) (memstore.Item, bool) {
	return func(obj memstore.Item) (memstore.Item, bool) {
		rwlocker := obj.(RWRecordLocker)

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
		return obj, true
	}
}

/*
	Generates a functor that can be used with a generic RWRecordLocker for Lock/Unlock with Lockneeds
*/
func RecordLockingFunctorGenerator(
	mem *memstore.Memstore,
	lockingType LockingType,
	makeSearchRecordFunctor func(string) memstore.Item,
	indexString string,
	save bool,
	collectionPtr *[]memstore.Item,
) func(string, LockType) bool {
	return func(id string, lockType LockType) bool {
		searchRecord := makeSearchRecordFunctor(id)
		memstoreItem := mem.UpdateData(searchRecord, indexString, lockingFunctorGenerator(lockType, lockingType))
		if memstoreItem == nil {
			return false
		} else if save {
			*collectionPtr = append(*collectionPtr, memstoreItem)
		}
		return true
	}
}
