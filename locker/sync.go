/*
	Helpers for user lock synchronization
	Makes calls to core.Lock for the locking logic
*/

package locker

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/memstore"
)

func lockResources(store *memstore.Memstore, lockNeeds []core.LockNeed) bool {
	// Build lock functions
	doLocking := core.RecordLockingFunctorGenerator(store, core.Locking, makeSearchByIdRecord, "id", false, nil)
	doUnlocking := core.RecordLockingFunctorGenerator(store, core.Unlocking, makeSearchByIdRecord, "id", false, nil)

	// Do locking (rollback unlocking included)
	return core.Lock(doLocking, doUnlocking, lockNeeds)
}

func unlockResources(store *memstore.Memstore, unlockNeeds []core.LockNeed) bool {
	// Build unlock function
	doUnlocking := core.RecordLockingFunctorGenerator(store, core.Unlocking, makeSearchByIdRecord, "id", false, nil)

	// Do unlocking
	return core.Unlock(doUnlocking, unlockNeeds)
}
