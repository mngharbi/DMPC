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
	// Build reader function
	reader := core.RecordReaderFunctor(store, makeSearchByIdRecord, "id", false, nil)

	// Do locking (rollback unlocking included)
	return core.Lock(reader, lockNeeds)
}

func unlockResources(store *memstore.Memstore, unlockNeeds []core.LockNeed) bool {
	// Build unlock function
	reader := core.RecordReaderFunctor(store, makeSearchByIdRecord, "id", false, nil)

	// Do unlocking
	return core.Unlock(reader, unlockNeeds)
}
