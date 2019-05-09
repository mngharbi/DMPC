/*
	Helpers for user lock synchronization
	Makes calls to core.Lock for the locking logic
*/

package users

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/memstore"
)

func lockUsers(sv *server, lockNeeds []core.LockNeed) (userRecords []*userRecord, lockingSuccess bool) {
	// Build lock functions
	var userRecordsItems []memstore.Item
	doLocking := core.RecordLockingFunctorGenerator(sv.store, core.Locking, makeSearchByIdRecord, idIndexStr, true, &userRecordsItems)
	doUnlocking := core.RecordLockingFunctorGenerator(sv.store, core.Unlocking, makeSearchByIdRecord, idIndexStr, false, nil)

	// Do locking (rollback unlocking included)
	lockingSuccess = core.Lock(doLocking, doUnlocking, lockNeeds)

	// Build list of records from item interfaces
	for _, recordItem := range userRecordsItems {
		userRecords = append(userRecords, recordItem.(*userRecord))
	}

	return
}

func unlockUsers(sv *server, unlockNeeds []core.LockNeed) (userRecords []*userRecord, unlockingSuccess bool) {
	// Build unlock function
	var userRecordsItems []memstore.Item
	doUnlocking := core.RecordLockingFunctorGenerator(sv.store, core.Unlocking, makeSearchByIdRecord, idIndexStr, true, &userRecordsItems)

	// Do unlocking
	unlockingSuccess = core.Unlock(doUnlocking, unlockNeeds)

	// If locking failed, don't return any results
	if !unlockingSuccess {
		userRecords = []*userRecord{}
	} else {
		// Build list of records from item interfaces
		for _, recordItem := range userRecordsItems {
			userRecords = append(userRecords, recordItem.(*userRecord))
		}
	}
	return
}
