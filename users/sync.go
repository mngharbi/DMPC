/*
	Helpers for user lock synchronization
	Makes calls to core.Lock for the locking logic
*/

package users

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/memstore"
)

func memstoreLockingFunctor(lockType core.LockType, isLock bool) (func(memstore.Item) (memstore.Item, bool)) {
	return func(obj memstore.Item) (memstore.Item, bool) {
		objCopy := obj.(userRecord)

		if lockType == core.WriteLockType {
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

func coreLockingFunctor(sv *server, collectionPtr *[]*userRecord, isLock bool) (func(string, core.LockType) bool) {
	return func(id string, lockType core.LockType) bool {
		memstoreItem := sv.store.UpdateData(userRecord{Id: id}, "id", memstoreLockingFunctor(lockType, isLock))
		if memstoreItem == nil {
			return false
		} else {
			*collectionPtr = append(*collectionPtr, memstoreItem.(*userRecord))
			return true
		}
	}
}

func lockUsers(sv *server, lockNeeds []core.LockNeed) (userRecords []*userRecord, lockingSuccess bool) {
	// Build lock functions
	doLocking := coreLockingFunctor(sv, &userRecords, true)
	doUnlocking := coreLockingFunctor(sv, &userRecords, false)

	// Do locking (rollback unlocking included)
	lockingSuccess = core.Lock(doLocking, doUnlocking, lockNeeds)

	// If locking failed, don't return any results
	if !lockingSuccess {
		userRecords = []*userRecord{}
	}
	return
}

func unlockUsers(sv *server, unlockNeeds []core.LockNeed) (userRecords []*userRecord, unlockingSuccess bool) {
	// Build unlock function
	doUnlocking := coreLockingFunctor(sv, &userRecords, false)

	// Do unlocking
	unlockingSuccess = core.Unlock(doUnlocking, unlockNeeds)

	// If locking failed, don't return any results
	if !unlockingSuccess {
		userRecords = []*userRecord{}
	}
	return
}
