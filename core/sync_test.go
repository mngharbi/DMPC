package core

import (
	"reflect"
	"testing"
)

/*
	Locking
*/

func TestNoLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, true, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("No locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			locked, expectedLocked,
			unlocked, expectedUnlocked,
			initialMap, expectedMap,
		)
	}
}

func TestAllWriteLock(t *testing.T) {
	var locked []LockNeed
	expectedLocked := []LockNeed{
		{true, "1"},
		{true, "2"},
		{true, "3"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{
		"1": true,
		"2": true,
		"3": true,
	}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, true, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{true, "1"},
		{true, "2"},
		{true, "3"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Write locking failed, results:\n locked: %v\n expectedLocked: %v\n unlocked: %v\n expectedUnlocked: %v\n initialMap: %v\n expectedMap: %v\n",
			locked, expectedLocked,
			unlocked, expectedUnlocked,
			initialMap, expectedMap,
		)
	}
}

func TestAllReadLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{
		{false, "1"},
		{false, "2"},
		{false, "3"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{
		"1": false,
		"2": false,
		"3": false,
	}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, true, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{false, "1"},
		{false, "2"},
		{false, "3"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Read locking failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestDuplicateWriteLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{
		{true, "1"},
		{true, "2"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{
		"1": true,
		"2": true,
	}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, true, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{true, "1"},
		{true, "2"},
		{true, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Duplicate write locking failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestDuplicateReadLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{
		{false, "1"},
		{false, "2"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{
		"1": false,
		"2": false,
	}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, true, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{false, "1"},
		{false, "2"},
		{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Read locking failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestOverwritingReadLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{
		{false, "2"},
		{true, "1"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{
		"1": true,
		"2": false,
	}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, true, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{false, "1"},
		{true, "1"},
		{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Overwriting write lock, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestOverwritingReadAfterWriteLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{
		{false, "2"},
		{true, "1"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{
		"1": true,
		"2": false,
	}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, true, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{true, "1"},
		{false, "1"},
		{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Overwriting read lock failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

/*
	Locking rollback
*/

func RollbackFailedWriteLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{
		{false, "2"},
		{true, "1"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{true, "1"},
		{false, "2"},
	}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, false, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{true, "1"},
		{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Rollback locks, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func RollbackDuplicateFailedWriteLock(t *testing.T) {
	locked := []LockNeed{}
	expectedLocked := []LockNeed{
		{false, "2"},
		{true, "1"},
	}

	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{true, "1"},
		{false, "2"},
	}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{}

	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initialMap, false, true, &locked, &lockDuplication)
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		{true, "1"},
		{false, "1"},
		{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(locked, expectedLocked) || !reflect.DeepEqual(unlocked, expectedUnlocked) ||
		!reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Rollback locks, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

/*
	Unlocking
*/

func TestNoUnlock(t *testing.T) {
	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{}

	initialMap := map[string]LockType{}
	expectedMap := map[string]LockType{}

	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(unlocked, expectedUnlocked) || !reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("No unlocking failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestAllWriteUnlock(t *testing.T) {
	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{true, "3"},
		{true, "2"},
		{true, "1"},
	}

	initialMap := map[string]LockType{
		"1": true,
		"2": true,
		"3": true,
	}
	expectedMap := map[string]LockType{}

	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		{true, "1"},
		{true, "2"},
		{true, "3"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(unlocked, expectedUnlocked) || !reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Write unlocking failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestAllReadUnlock(t *testing.T) {
	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{false, "3"},
		{false, "2"},
		{false, "1"},
	}

	initialMap := map[string]LockType{
		"1": false,
		"2": false,
		"3": false,
	}
	expectedMap := map[string]LockType{}

	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		{false, "1"},
		{false, "2"},
		{false, "3"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(unlocked, expectedUnlocked) || !reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Read unlocking failed, results:\n result: %v\n expected: %v\n", initialMap, initialMap)
	}
}

func TestDuplicateWriteUnlock(t *testing.T) {
	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{true, "2"},
		{true, "1"},
	}

	initialMap := map[string]LockType{
		"1": true,
		"2": true,
	}
	expectedMap := map[string]LockType{}

	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		{true, "1"},
		{true, "2"},
		{true, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(unlocked, expectedUnlocked) || !reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Duplicate write unlocking failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestDuplicateReadUnlock(t *testing.T) {
	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{false, "2"},
		{false, "1"},
	}

	initialMap := map[string]LockType{
		"1": false,
		"2": false,
	}
	expectedMap := map[string]LockType{}

	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		{false, "1"},
		{false, "2"},
		{false, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(unlocked, expectedUnlocked) || !reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Duplicate read unlocking failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestOverwritingReadUnlock(t *testing.T) {
	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{true, "1"},
		{false, "2"},
	}

	initialMap := map[string]LockType{
		"1": true,
		"2": false,
	}
	expectedMap := map[string]LockType{}

	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		{false, "1"},
		{true, "1"},
		{false, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(unlocked, expectedUnlocked) || !reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Overwriting write unlock, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

func TestOverwritingReadAfterWriteUnlock(t *testing.T) {
	unlocked := []LockNeed{}
	expectedUnlocked := []LockNeed{
		{true, "1"},
		{false, "2"},
	}

	initialMap := map[string]LockType{
		"1": true,
		"2": false,
	}
	expectedMap := map[string]LockType{}

	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initialMap, true, false, &unlocked, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		{true, "1"},
		{false, "1"},
		{false, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(unlocked, expectedUnlocked) || !reflect.DeepEqual(initialMap, expectedMap) {
		t.Errorf("Overwriting read unlock failed, results:\n result: %v\n expected: %v\n", initialMap, expectedMap)
	}
}

/*
	Helpers
*/

func lockingFunctor(dst map[string]LockType, success bool, lock bool, called *[]LockNeed, duplication *bool) func(string, LockType) bool {
	return func(fId string, fType LockType) bool {
		*called = append(*called, LockNeed{fType, fId})

		if !success {
			return false
		}

		if lock {
			if _, ok := dst[fId]; ok {
				*duplication = true
			}

			dst[fId] = fType
		} else {
			if _, ok := dst[fId]; ok {
				delete(dst, fId)
			} else {
				*duplication = true
			}
		}
		return true
	}
}
