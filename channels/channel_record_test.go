package channels

import (
	"reflect"
	"testing"
	"time"
)

/*
	Test actions
*/
func TestTryOpen(t *testing.T) {
	// Open a buffered record
	rec := &channelRecord{
		state: channelBufferedState,
	}
	if rec.tryOpen(
		genericChannelId,
		nil,
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
	) {
		t.Error("Opening a buffered channel with an invalid opening record should fail")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Time{}},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
	) {
		t.Error("Opening a buffered channel with a zero opening time should fail")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		nil,
	) {
		t.Error("Opening a buffered channel with no permissions record should fail")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{}},
	) {
		t.Error("Opening a buffered channel with no users should fail")
	}

	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
	) {
		t.Error("Opening a buffered channel should not fail")
	}
	if rec.computeState() != channelOpenState {
		t.Error("Opening a buffered channel should move it into open state")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
	) {
		t.Error("Reopening a channel should fail")
	}
}

func TestTryClose(t *testing.T) {
	// Close a buffered record
	rec := &channelRecord{
		state: channelBufferedState,
	}
	if !rec.tryClose(
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
	) {
		t.Error("Closing a buffered channel should not fail")
	}
	if rec.computeState() != channelBufferedState ||
		len(rec.closureAttempts) != 1 {
		t.Error("Closing a buffered channel should add closure attempt")
	}

	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
	) {
		t.Error("Opening a buffered channel should not fail")
	}

	// Close an open record
	if rec.tryClose(
		nil,
	) {
		t.Error("Closing a channel without closure record should fail")
	}
	if rec.tryClose(
		&channelActionRecord{genericIssuerId, genericIssuerId, time.Now()},
	) {
		t.Error("Closing a channel by user that doesn't have permissions should fail")
	}
	if rec.tryClose(
		&channelActionRecord{genericUserId, genericUserId, time.Now().Add(-1 * time.Hour)},
	) {
		t.Error("Closing a channel before its opening time should fail")
	}
	if !rec.tryClose(
		&channelActionRecord{genericUserId, genericUserId, time.Now()},
	) {
		t.Error("Closing a channel should not fail")
	}
	if rec.computeState() != channelClosedState ||
		rec.closure == nil {
		t.Error("Closing an open channel should move it to closed state ")
	}
}

func TestApplyCloseAttemptsInvalid(t *testing.T) {
	currentTime := time.Now()

	// Create buffered channel
	rec := &channelRecord{
		state: channelBufferedState,
	}

	// Add invalid close attempt (no permissions)
	if !rec.tryClose(
		&channelActionRecord{genericIssuerId, genericIssuerId, currentTime.Add(1 * time.Hour)},
	) {
		t.Error("Attempting to close a buffered channel should not fail")
	}

	// Try to apply close attempt before it's open
	if rec.applyCloseAttempts() {
		t.Error("Applying close attempts to a buffered channel should fail")
	}

	// Add invalid close attempt again
	if !rec.tryClose(
		&channelActionRecord{genericIssuerId, genericIssuerId, currentTime.Add(1 * time.Hour)},
	) {
		t.Error("Attempting to close a buffered channel should not fail")
	}
	// Open channel
	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
	) {
		t.Error("Opening a buffered channel should not fail")
	}

	// Try to apply close attempt with no valid attempts
	if rec.applyCloseAttempts() {
		t.Error("No close attempts should work")
	}
}

func TestApplyCloseAttempts(t *testing.T) {
	currentTime := time.Now()

	// Create buffered channel
	rec := &channelRecord{
		state: channelBufferedState,
	}

	// Add valid attempts
	attempts := 5
	for i := attempts; i > 0; i-- {
		if !rec.tryClose(
			&channelActionRecord{genericUserId, genericUserId, currentTime.Add(time.Duration(i) * time.Hour)},
		) {
			t.Error("Attempting to close a buffered channel should not fail")
		}
	}
	if rec.computeState() != channelBufferedState ||
		len(rec.closureAttempts) != attempts {
		t.Error("Attempting to close a buffered channel should add closure attempt")
	}

	// Open channel
	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
	) {
		t.Error("Opening a buffered channel should not fail")
	}

	if !rec.applyCloseAttempts() {
		t.Error("Applying close attempts to an open channel should not fail")
	}
	expectedClosureRecord := &channelActionRecord{genericUserId, genericUserId, currentTime.Add(1 * time.Hour)}
	if rec.computeState() != channelClosedState ||
		!reflect.DeepEqual(rec.closure, expectedClosureRecord) {
		t.Error("Applying close attempts should apply them in ascending time ordering")
	}
}

/*
	Test add message
*/
func TestAddMessage(t *testing.T) {
	// Build record with 3 timestamps
	recVal := channelRecord{
		messageTimestamps: []time.Time{},
	}
	curTime := time.Now()
	pos1Time := curTime.Add(time.Minute)
	for i := 0; i < 3; i++ {
		recVal.messageTimestamps = append(recVal.messageTimestamps, curTime)
		curTime = curTime.Add(time.Hour)
	}
	rec := &channelRecord{}
	*rec = recVal

	// Test valid add message with open state
	*rec = recVal
	rec.state = channelOpenState
	if pos, ok := rec.addMessage(pos1Time); !ok || pos != 1 {
		t.Error("Adding valid message to an open channel should not fail")
	}

	// Test valid add message with closed state
	*rec = recVal
	rec.state = channelClosedState
	if pos, ok := rec.addMessage(pos1Time); !ok || pos != 1 {
		t.Error("Adding valid message to a closed channel should not fail")
	}

	// Test two valid add message with open state
	*rec = recVal
	rec.state = channelOpenState
	if pos, ok := rec.addMessage(pos1Time); !ok || pos != 1 {
		t.Error("Adding valid message to an open channel should not fail")
	}
	if pos, ok := rec.addMessage(pos1Time.Add(time.Minute)); !ok || pos != 2 {
		t.Error("Adding second valid message should update ordering")
	}

	// Add to a buffered record
	*rec = recVal
	rec.state = channelBufferedState
	if _, ok := rec.addMessage(pos1Time); ok {
		t.Error("Adding message to a buffered channel should fail")
	}

	// Add zero timestamp message
	*rec = recVal
	rec.state = channelOpenState
	if _, ok := rec.addMessage(time.Time{}); ok {
		t.Error("Adding message with zero timestamp should fail")
	}
}
