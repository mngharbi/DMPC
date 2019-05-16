package core

import (
	"github.com/mngharbi/memstore"
	"reflect"
	"testing"
)

/*
	Test structure
*/

type testLocker struct {
	id string
}

func (l *testLocker) Lock() {
	l.doLock(WriteLockType, Locking)
}
func (l *testLocker) Unlock() {
	l.doLock(WriteLockType, Unlocking)
}
func (l *testLocker) RLock() {
	l.doLock(ReadLockType, Locking)
}
func (l *testLocker) RUnlock() {
	l.doLock(ReadLockType, Unlocking)
}

func (l *testLocker) Less(string, interface{}) bool {
	return true
}

/*
	Helpers
*/

var (
	syncLocked       map[string]LockType
	syncLockCalls    []LockNeed
	syncUnlockCalls  []LockNeed
	syncDoubleLock   bool
	syncDoubleUnlock bool
	syncWrongUnlock  bool
)

func resetSync() {
	syncLocked = map[string]LockType{}
	syncLockCalls = nil
	syncUnlockCalls = nil
	syncDoubleLock = false
	syncDoubleUnlock = false
	syncWrongUnlock = false
}

func (l *testLocker) doLock(lockType LockType, lockingType LockingType) {
	id := l.id
	lockedType, isLocked := syncLocked[id]
	if lockingType == Locking && isLocked {
		syncDoubleLock = true
		return
	}
	if lockingType == Unlocking && !isLocked {
		syncDoubleUnlock = true
		return
	}
	if lockingType == Unlocking && isLocked && lockedType != lockType {
		syncWrongUnlock = true
		return
	}
	if lockingType == Locking {
		syncLocked[id] = lockType
	} else {
		delete(syncLocked, id)
	}

	var callsSlice *[]LockNeed
	if lockingType == Locking {
		callsSlice = &syncLockCalls
	} else {
		callsSlice = &syncUnlockCalls
	}
	*callsSlice = append(*callsSlice, LockNeed{
		LockType: lockType,
		Id:       id,
	})
}

func readerFunctor(data []memstore.Item) RecordReader {
	return func(ids []string) []memstore.Item {
		return data
	}
}

/*
	Tests
*/

func checkFlags(t *testing.T) bool {
	if syncDoubleLock {
		t.Error("Lock called twice on the same id")
	}
	if syncDoubleUnlock {
		t.Error("Unlock called twice on the same id")
	}
	if syncWrongUnlock {
		t.Error("Unlock does not match lock")
	}
	return syncDoubleLock && syncDoubleUnlock && syncWrongUnlock
}

func TestNoLock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{}
	expectedLocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor(nil)

	Lock(reader, []LockNeed{})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("No locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestAllWriteLock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{
		"1": WriteLockType,
		"2": WriteLockType,
		"3": WriteLockType,
	}
	expectedLocked := []LockNeed{
		{WriteLockType, "1"},
		{WriteLockType, "2"},
		{WriteLockType, "3"},
	}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
		&testLocker{"3"},
	})

	Lock(reader, []LockNeed{
		{WriteLockType, "1"},
		{WriteLockType, "2"},
		{WriteLockType, "3"},
	})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Write locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestAllReadLock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{
		"1": ReadLockType,
		"2": ReadLockType,
		"3": ReadLockType,
	}
	expectedLocked := []LockNeed{
		{ReadLockType, "1"},
		{ReadLockType, "2"},
		{ReadLockType, "3"},
	}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
		&testLocker{"3"},
	})

	Lock(reader, []LockNeed{
		{ReadLockType, "1"},
		{ReadLockType, "2"},
		{ReadLockType, "3"},
	})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Read locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestDuplicateWriteLock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{
		"1": WriteLockType,
		"2": WriteLockType,
	}
	expectedLocked := []LockNeed{
		{WriteLockType, "1"},
		{WriteLockType, "2"},
	}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
	})

	Lock(reader, []LockNeed{
		{WriteLockType, "1"},
		{WriteLockType, "2"},
		{WriteLockType, "2"},
	})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Double locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestDuplicateReadLock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{
		"1": ReadLockType,
		"2": ReadLockType,
	}
	expectedLocked := []LockNeed{
		{ReadLockType, "1"},
		{ReadLockType, "2"},
	}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
	})

	Lock(reader, []LockNeed{
		{ReadLockType, "1"},
		{ReadLockType, "2"},
		{ReadLockType, "2"},
	})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Double read locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestOverwritingReadLock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{
		"1": WriteLockType,
		"2": ReadLockType,
	}
	expectedLocked := []LockNeed{
		{ReadLockType, "2"},
		{WriteLockType, "1"},
	}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
	})

	Lock(reader, []LockNeed{
		{ReadLockType, "1"},
		{WriteLockType, "1"},
		{ReadLockType, "2"},
	})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Overwrite read locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestOverwritingReadLockReverse(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{
		"1": WriteLockType,
		"2": ReadLockType,
	}
	expectedLocked := []LockNeed{
		{ReadLockType, "2"},
		{WriteLockType, "1"},
	}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
	})

	Lock(reader, []LockNeed{
		{WriteLockType, "1"},
		{ReadLockType, "1"},
		{ReadLockType, "2"},
	})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Reverse overwrite read locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestNoUnlock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{}
	expectedLocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	reader := readerFunctor(nil)

	Unlock(reader, []LockNeed{})

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("No unlocking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestAllWriteLockUnlock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{}
	expectedLocked := []LockNeed{
		{WriteLockType, "1"},
		{WriteLockType, "2"},
		{WriteLockType, "3"},
	}
	expectedUnlocked := []LockNeed{
		{WriteLockType, "3"},
		{WriteLockType, "2"},
		{WriteLockType, "1"},
	}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
		&testLocker{"3"},
	})

	lockNeeds := []LockNeed{
		{WriteLockType, "1"},
		{WriteLockType, "2"},
		{WriteLockType, "3"},
	}
	Lock(reader, lockNeeds)
	Unlock(reader, lockNeeds)

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Write lock/unlock failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestAllReadUnlock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{}
	expectedLocked := []LockNeed{
		{ReadLockType, "1"},
		{ReadLockType, "2"},
		{ReadLockType, "3"},
	}
	expectedUnlocked := []LockNeed{
		{ReadLockType, "3"},
		{ReadLockType, "2"},
		{ReadLockType, "1"},
	}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
		&testLocker{"2"},
		&testLocker{"3"},
	})

	lockNeeds := []LockNeed{
		{ReadLockType, "1"},
		{ReadLockType, "2"},
		{ReadLockType, "3"},
	}
	Lock(reader, lockNeeds)
	Unlock(reader, lockNeeds)

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Read lock/unlock failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestDuplicateWriteLockUnlock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{}
	expectedLocked := []LockNeed{
		{WriteLockType, "1"},
	}
	expectedUnlocked := []LockNeed{
		{WriteLockType, "1"},
	}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
	})

	lockNeeds := []LockNeed{
		{WriteLockType, "1"},
		{WriteLockType, "1"},
	}
	Lock(reader, lockNeeds)
	Unlock(reader, lockNeeds)

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Read lock/unlock failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}

func TestDuplicateReadUnlock(t *testing.T) {
	resetSync()

	expectedMap := map[string]LockType{}
	expectedLocked := []LockNeed{
		{ReadLockType, "1"},
	}
	expectedUnlocked := []LockNeed{
		{ReadLockType, "1"},
	}

	reader := readerFunctor([]memstore.Item{
		&testLocker{"1"},
	})

	lockNeeds := []LockNeed{
		{ReadLockType, "1"},
		{ReadLockType, "1"},
	}
	Lock(reader, lockNeeds)
	Unlock(reader, lockNeeds)

	if !checkFlags(t) {
		return
	}

	if !reflect.DeepEqual(syncLockCalls, expectedLocked) || !reflect.DeepEqual(syncUnlockCalls, expectedUnlocked) ||
		!reflect.DeepEqual(syncLocked, expectedMap) {
		t.Errorf("Read lock/unlock failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			syncLockCalls, expectedLocked,
			syncUnlockCalls, expectedUnlocked,
			syncLocked, expectedMap,
		)
	}
}
