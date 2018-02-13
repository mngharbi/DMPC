package core

import (
	"testing"
	"reflect"
)

/*
	Locking
*/

func TestNoLock(t *testing.T) {
	expected := map[string]bool {}
	initial := map[string]bool{}

	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, true, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("No locking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestAllWriteLock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": true,
		"3": true,
	}

	initial := map[string]bool{}

	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, true, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{true, "2"},
		LockNeed{true, "3"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Write locking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestAllReadLock(t *testing.T) {
	expected := map[string]bool {
		"1": false,
		"2": false,
		"3": false,
	}

	initial := map[string]bool{}


	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, true, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{false, "1"},
		LockNeed{false, "2"},
		LockNeed{false, "3"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Read locking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestDuplicateWriteLock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": true,
	}

	initial := map[string]bool{}

	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, true, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{true, "2"},
		LockNeed{true, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Duplicate write locking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestDuplicateReadLock(t *testing.T) {
	expected := map[string]bool {
		"1": false,
		"2": false,
	}

	initial := map[string]bool{}


	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, true, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{false, "1"},
		LockNeed{false, "2"},
		LockNeed{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Read locking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestOverwritingReadLock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": false,
	}

	initial := map[string]bool{}


	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, true, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{false, "1"},
		LockNeed{true, "1"},
		LockNeed{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Overwriting write lock, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestOverwritingReadAfterWriteLock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": false,
	}

	initial := map[string]bool{}


	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, true, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{false, "1"},
		LockNeed{false, "2"},
	})

	if lockDuplication {
		t.Error("Lock called twice on the same id")
	}
	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Overwriting read lock failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

/*
	Locking rollback
*/

func RollbackFailedWriteLock(t *testing.T) {
	expected := map[string]bool {}

	initial := map[string]bool{}


	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, false, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{false, "2"},
	})

	if lockCounter != 0 {
		t.Error("Lock called instead of failing")
	}
	if unlockCounter != 2 {
		t.Error("Unlock not called twice on lock failure")
	}

	if len(initial) == len(expected) && !reflect.DeepEqual(initial, expected) {
		t.Error("Rollback locks, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func RollbackDuplicateFailedWriteLock(t *testing.T) {
	expected := map[string]bool {}

	initial := map[string]bool{}


	var lockCounter, unlockCounter int
	var lockDuplication, unlockDuplication bool
	doLockSuccess := lockingFunctor(initial, false, true, &lockCounter, &lockDuplication)
	doUnlockFail := lockingFunctor(initial, false, false, &unlockCounter, &unlockDuplication)

	Lock(doLockSuccess, doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{false, "1"},
		LockNeed{false, "2"},
	})

	if lockCounter != 0 {
		t.Error("Lock called instead of failing")
	}
	if unlockCounter != 2 {
		t.Error("Unlock not called twice on lock failure")
	}

	if len(initial) == len(expected) && !reflect.DeepEqual(initial, expected) {
		t.Error("Rollback locks, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

/*
	Unlocking
*/

func TestNoUnlock(t *testing.T) {
	expected := map[string]bool {}
	initial := map[string]bool{}

	var unlockCounter int
	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initial, true, true, &unlockCounter, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("No unlocking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestAllWriteUnlock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": true,
		"3": true,
	}

	initial := map[string]bool{}

	var unlockCounter int
	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initial, true, true, &unlockCounter, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{true, "2"},
		LockNeed{true, "3"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Write unlocking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestAllReadUnlock(t *testing.T) {
	expected := map[string]bool {
		"1": false,
		"2": false,
		"3": false,
	}

	initial := map[string]bool{}

	var unlockCounter int
	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initial, true, true, &unlockCounter, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		LockNeed{false, "1"},
		LockNeed{false, "2"},
		LockNeed{false, "3"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Read unlocking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestDuplicateWriteUnlock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": true,
	}

	initial := map[string]bool{}

	var unlockCounter int
	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initial, true, true, &unlockCounter, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{true, "2"},
		LockNeed{true, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Duplicate write unlocking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestDuplicateReadUnlock(t *testing.T) {
	expected := map[string]bool {
		"1": false,
		"2": false,
	}

	initial := map[string]bool{}

	var unlockCounter int
	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initial, true, true, &unlockCounter, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		LockNeed{false, "1"},
		LockNeed{false, "2"},
		LockNeed{false, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Read unlocking failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestOverwritingReadUnlock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": false,
	}

	initial := map[string]bool{}

	var unlockCounter int
	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initial, true, true, &unlockCounter, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		LockNeed{false, "1"},
		LockNeed{true, "1"},
		LockNeed{false, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Overwriting write unlock, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

func TestOverwritingReadAfterWriteUnlock(t *testing.T) {
	expected := map[string]bool {
		"1": true,
		"2": false,
	}

	initial := map[string]bool{}

	var unlockCounter int
	var unlockDuplication bool
	doUnlockFail := lockingFunctor(initial, true, true, &unlockCounter, &unlockDuplication)

	Unlock(doUnlockFail, []LockNeed{
		LockNeed{true, "1"},
		LockNeed{false, "1"},
		LockNeed{false, "2"},
	})

	if unlockDuplication {
		t.Error("Unlock called twice on the same id")
	}

	if !reflect.DeepEqual(initial, expected) {
		t.Error("Overwriting read unlock failed, results:\n result: %v\n expected: %v\n", initial, expected)
	}
}

/*
	Helpers
*/

func lockingFunctor (dst map[string]bool, success bool, lock bool, lockCounter *int, duplication *bool) (func(string, bool) bool) {
	return func(fId string, fType bool) bool {
		if !success {
			return false
		}

		if lock {
			*lockCounter += 1
			if _, ok := dst[fId]; ok {
				*duplication = true
			}

			dst[fId] = fType
		} else {
			*lockCounter += 1
			if _, ok := dst[fId]; ok {
				*duplication = true
				delete(dst, fId)
			}
		}
		return true
	}
}
