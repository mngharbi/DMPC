/*
	Locking/Unlocking utilities to prevent deadlocks
*/

package core

import (
	"sort"
)
type LockNeed struct {
	WriteLock bool
	Id string
}

type lockNeedCollection []LockNeed
func (s lockNeedCollection) Len() int { return len(s) }
func (s lockNeedCollection) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s lockNeedCollection) Less(i, j int) bool { return s[i].Less(s[j]) }


// Order: read low id < read high id < write low id < write high id
func (a LockNeed) Less (b LockNeed) bool {
	if a.WriteLock != b.WriteLock {
		return !a.WriteLock
	}
	return a.Id < b.Id
}

func sanitizeLockNeeds(lockNeeds []LockNeed) []LockNeed {
	/*
		Remove duplicates and keep higher privelege lock
		0 (default): no locks
		1: read lock
		2: write lock
	*/
	lockMap := map[string]int{}
	for _,lockNeed := range lockNeeds {
		var lockIntMapping int = 1
		if lockNeed.WriteLock {
			lockIntMapping = 2
		}

		// Higher integer overwrites previous (no lock < read lock < write lock)
		if lockIntMapping > lockMap[lockNeed.Id] {
			lockMap[lockNeed.Id] = lockIntMapping
		}
	}

	// Build sanitized lock needs array to prevent deadlocks
	var orderedWriteLocks []LockNeed
	for lockId, lockTypeInt := range lockMap {
		isWriteLock := false
		if lockTypeInt == 2 {
			isWriteLock = true
		}
		orderedWriteLocks = append(orderedWriteLocks, LockNeed{isWriteLock, lockId})
	}
	sort.Sort(lockNeedCollection(orderedWriteLocks))

	return orderedWriteLocks
}

func sanitizeUnlockNeeds(lockNeeds []LockNeed) []LockNeed {
	sanitizedLockNeeds := sanitizeLockNeeds(lockNeeds)

	// Reverse for unlocking
	for i := 0; i < len(sanitizedLockNeeds) / 2; i++ {
		oppositeI := len(sanitizedLockNeeds) - i - 1
		sanitizedLockNeeds[i], sanitizedLockNeeds[oppositeI] = sanitizedLockNeeds[oppositeI], sanitizedLockNeeds[i]
	}

	return sanitizedLockNeeds
}

func Lock(doLock func(string, bool) bool, doUnlock func(string, bool) bool, lockNeeds []LockNeed) {
	// Sanitize lock needs to avoid deadlocks
	sanitizedLockNeeds := sanitizeLockNeeds(lockNeeds)

	// Build map of lock needs
	isWriteLockMap := map[string]bool{}
	for _, lockNeed := range sanitizedLockNeeds {
		isWriteLockMap[lockNeed.Id] = lockNeed.WriteLock
	}

	// Perform locking
	needToUnlock := false
	var successfulLocks []LockNeed
	for _, lockNeed := range sanitizedLockNeeds {
		if !doLock(lockNeed.Id, lockNeed.WriteLock) {
			needToUnlock = true
		} else {
			successfulLocks = append(successfulLocks, lockNeed)
		}
	}

	// If none failed we're done
	if !needToUnlock {
		return
	}

	// Otherwise, unlock everything locked (in reverse)
	for _, lockNeed := range successfulLocks {
		doUnlock(lockNeed.Id, lockNeed.WriteLock)
	}
}

func Unlock(doUnlock func(string, bool) bool, unlockNeeds []LockNeed) {
	// Sanitize unlock needs to avoid deadlocks
	sanitizedUnlockNeeds := sanitizeUnlockNeeds(unlockNeeds)

	// Build map of lock needs
	isWriteLockMap := map[string]bool{}
	for _, lockNeed := range sanitizedUnlockNeeds {
		isWriteLockMap[lockNeed.Id] = lockNeed.WriteLock
	}

	// Perform unlocking
	for _, lockNeed := range sanitizedUnlockNeeds {
		doUnlock(lockNeed.Id, lockNeed.WriteLock)
	}
}
