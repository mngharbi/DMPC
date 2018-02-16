/*
	Locking/Unlocking utilities to prevent deadlocks
*/

package users

import (
	"github.com/mngharbi/memstore"
	"sort"
)
type lockNeed struct {
	writeLock	bool
	userId		string
}

type lockNeedCollection []lockNeed
func (s lockNeedCollection) Len() int { return len(s) }
func (s lockNeedCollection) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s lockNeedCollection) Less(i, j int) bool { return s[i].Less(s[j]) }


// Order: read low id < read high id < write low id < write high id
func (a lockNeed) Less (b lockNeed) bool {
	if a.writeLock != b.writeLock {
		return !a.writeLock
	}
	return a.userId < b.userId
}

func sanitizeLockNeeds(lockNeeds []lockNeed) []lockNeed {
	/*
		Remove duplicates and keep higher privelege lock
		0 (default): no locks
		1: read lock
		2: write lock
	*/
	var usersLockMap map[string]int
	for _,lockNeed := range lockNeeds {
		var lockIntMapping int = 1
		if lockNeed.writeLock {
			lockIntMapping = 2
		}

		// Higher integer overwrites previous (no lock < read lock < write lock)
		if lockIntMapping > usersLockMap[lockNeed.userId] {
			usersLockMap[lockNeed.userId] = lockIntMapping
		}
	}

	// Build sanitized lock needs array to prevent deadlocks
	var orderedWriteLocks []lockNeed
	for lockUserId, lockTypeInt := range usersLockMap {
		isWriteLock := false
		if lockTypeInt == 1 {
			isWriteLock = true
		}
		orderedWriteLocks = append(orderedWriteLocks, lockNeed{isWriteLock, lockUserId})
	}
	sort.Sort(lockNeedCollection(orderedWriteLocks))

	return orderedWriteLocks
}

func sanitizeUnlockNeeds(lockNeeds []lockNeed) []lockNeed {
	sanitizedLockNeeds := sanitizeLockNeeds(lockNeeds)

	// Reverse for unlocking
	for i := 0; i < len(sanitizedLockNeeds) / 2; i++ {
		oppositeI := len(sanitizedLockNeeds) - i - 1
		sanitizedLockNeeds[i], sanitizedLockNeeds[oppositeI] = sanitizedLockNeeds[oppositeI], sanitizedLockNeeds[i]
	}

	return sanitizedLockNeeds
}

func lockOperationFunctor(isWriteLockMap map[string]bool, isLock bool) (func(memstore.Item) (memstore.Item, bool)) {
	return func(obj memstore.Item) (memstore.Item, bool) {
		objCopy := obj.(userRecord)

		if isWriteLockMap[objCopy.Id] {
			if isLock {
				objCopy.Lock.Lock()
			} else {
				objCopy.Lock.Unlock()
			}
		} else {
			if isLock {
				objCopy.Lock.RLock()
			} else {
				objCopy.Lock.RUnlock()
			}
		}
		return obj, true
	}
}

func lockFunctor(isWriteLockMap map[string]bool) (func(memstore.Item) (memstore.Item, bool)) {
	return lockOperationFunctor(isWriteLockMap, true)
}

func unlockFunctor(isWriteLockMap map[string]bool) (func(memstore.Item) (memstore.Item, bool)) {
	return lockOperationFunctor(isWriteLockMap, false)
}

func lockUsers(sv *server, lockNeeds []lockNeed) ([]*userRecord, bool) {
	// Sanitize lock needs to avoid deadlocks
	sanitizedLockNeeds := sanitizeLockNeeds(lockNeeds)

	// Build map of lock needs
	isWriteLockMap := map[string]bool{}
	for _, lockNeed := range sanitizedLockNeeds {
		isWriteLockMap[lockNeed.userId] = lockNeed.writeLock
	}

	// Build subset of users for search
	var userSearchRecords []memstore.Item
	for _, lockNeed := range sanitizedLockNeeds {
		userSearchRecords = append(userSearchRecords, userRecord{Id: lockNeed.userId})
	}

	// Atomically acquire locks (use id as index)
	lockResultsItems := sv.store.UpdateDataSubset(userSearchRecords, "id", lockFunctor(isWriteLockMap))

	// Determine if any failed (non-existent user)
	needToUnlock := false
	var lockResultsFiltered []*userRecord
	for _, lockResult := range lockResultsItems {
		if lockResult == nil {
			needToUnlock = true
		} else {
			var userRecordPtr *userRecord
			*userRecordPtr = lockResult.(userRecord)
			lockResultsFiltered = append(lockResultsFiltered, userRecordPtr)
		}
	}

	// If none failed we're done
	if !needToUnlock {
		return lockResultsFiltered, true
	}

	// Otherwise, unlock everything locked (in reverse)
	userSearchRecordsExistent := []memstore.Item{}
	for userSearchIndex, userSearchObject := range userSearchRecords {
		if lockResultsItems[userSearchIndex] != nil {
			userSearchRecordsExistent = append([]memstore.Item{userSearchObject}, userSearchRecordsExistent...)
		}
	}
	sv.store.UpdateDataSubset(userSearchRecordsExistent, "id", unlockFunctor(isWriteLockMap))
	return lockResultsFiltered, false
}

func unlockUsers(sv *server, unlockNeeds []lockNeed) {
	// Sanitize unlock needs to avoid deadlocks
	sanitizedUnlockNeeds := sanitizeUnlockNeeds(unlockNeeds)

	// Build map of lock needs
	isWriteLockMap := map[string]bool{}
	for _, lockNeed := range sanitizedUnlockNeeds {
		isWriteLockMap[lockNeed.userId] = lockNeed.writeLock
	}

	// Build subset of users for search
	userSearchRecords := []memstore.Item{}
	for _, lockNeed := range sanitizedUnlockNeeds {
		userSearchRecords = append(userSearchRecords, userRecord{Id: lockNeed.userId})
	}

	// Atomically release locks (use id as index)
	sv.store.UpdateDataSubset(userSearchRecords, "id", unlockFunctor(isWriteLockMap))
}
