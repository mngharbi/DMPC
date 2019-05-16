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
	// Build reader functions
	var userRecordsItems []memstore.Item
	reader := core.RecordReaderFunctor(sv.store, makeSearchByIdRecord, idIndexStr, true, &userRecordsItems)

	// Do locking (rollback unlocking included)
	lockingSuccess = core.Lock(reader, lockNeeds)

	// Build list of records from item interfaces
	for _, recordItem := range userRecordsItems {
		if recordItem != nil {
			userRecords = append(userRecords, recordItem.(*userRecord))
		}
	}

	return
}

func unlockUsers(sv *server, unlockNeeds []core.LockNeed) (userRecords []*userRecord, unlockingSuccess bool) {
	// Build unlock function
	var userRecordsItems []memstore.Item
	reader := core.RecordReaderFunctor(sv.store, makeSearchByIdRecord, idIndexStr, true, &userRecordsItems)

	// Do unlocking
	unlockingSuccess = core.Unlock(reader, unlockNeeds)

	// If locking failed, don't return any results
	if !unlockingSuccess {
		userRecords = []*userRecord{}
	} else {
		// Build list of records from item interfaces
		for _, recordItem := range userRecordsItems {
			if recordItem != nil {
				userRecords = append(userRecords, recordItem.(*userRecord))
			}
		}
	}
	return
}
