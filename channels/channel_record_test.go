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
		genericKeyId,
	) {
		t.Error("Opening a buffered channel with an invalid opening record should fail")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Time{}},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
		genericKeyId,
	) {
		t.Error("Opening a buffered channel with a zero opening time should fail")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		nil,
		genericKeyId,
	) {
		t.Error("Opening a buffered channel with no permissions record should fail")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{}},
		genericKeyId,
	) {
		t.Error("Opening a buffered channel with no users should fail")
	}

	if rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
		"",
	) {
		t.Error("Opening a buffered channel with an empty key id should fail")
	}

	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
		genericKeyId,
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
		genericKeyId,
	) {
		t.Error("Reopening a channel should fail")
	}
}

func TestTryClose(t *testing.T) {
	// Close a buffered record
	rec := &channelRecord{
		state: channelBufferedState,
	}
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericCertifierId, time.Now()},
	); !ok {
		t.Error("Closing a buffered channel should not fail")
	}
	if rec.computeState() != channelBufferedState ||
		len(rec.closureAttempts) != 1 {
		t.Error("Closing a buffered channel should add closure attempt")
	}

	currentTime := time.Now()

	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericIssuerId:    {false, false, false},
			genericCertifierId: {true, true, true},
		}},
		genericKeyId,
	) {
		t.Error("Opening a buffered channel should not fail")
	}

	// Add a message after a minute of opening
	if _, ok := rec.addMessage(
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(time.Minute)},
	); !ok {
		t.Error("Adding a valid message should not fail")
	}

	// Add a message after 2 hours of opening
	if _, ok := rec.addMessage(
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(2 * time.Hour)},
	); !ok {
		t.Error("Adding a valid message should not fail")
	}

	// Close an open record
	if _, ok := rec.tryClose(nil); ok {
		t.Error("Closing a channel without closure record should fail")
	}
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericIssuerId, currentTime},
	); ok {
		t.Error("Closing a channel by user that doesn't have permissions should fail")
	}
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(-1 * time.Hour)},
	); ok {
		t.Error("Closing a channel before its opening time should fail")
	}
	remainingMessages, closeOk := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(time.Hour)},
	)
	if !closeOk {
		t.Error("Closing a channel should not fail")
	}
	if rec.computeState() != channelClosedState ||
		rec.closure == nil {
		t.Error("Closing an open channel should move it to closed state ")
	}
	if remainingMessages != 1 {
		t.Error("Closing an open channel with messages should remove late messages")
	}

	// Try to close after first close
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(2 * time.Hour)},
	); ok {
		t.Error("Closing a channel after its closing time should fail")
	}

	// Try to close at the same time with different issuer (alphanumerically higher)
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerIdAfter, genericCertifierId, currentTime.Add(time.Hour)},
	); ok {
		t.Error("Closing a channel again at its closing time with higher issuer id should fail")
	}

	// Try to close at the same time with different issuer (alphanumerically lower)
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerIdBefore, genericCertifierId, currentTime.Add(time.Hour)},
	); !ok {
		t.Error("Closing a channel again at its closing time with lower issuer id should not fail")
	}

	// Try to close before first close
	if remainingMessages, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(time.Minute)},
	); !ok || remainingMessages != 0 {
		t.Error("Closing a channel before its closing time should not fail")
	}

}

func TestApplyCloseAttemptsInvalid(t *testing.T) {
	currentTime := time.Now()

	// Create buffered channel
	rec := &channelRecord{
		state: channelBufferedState,
	}

	// Add invalid close attempt (no permissions)
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericIssuerId, currentTime.Add(1 * time.Hour)},
	); !ok {
		t.Error("Attempting to close a buffered channel should not fail")
	}

	// Try to apply close attempt before it's open
	if rec.applyCloseAttempts() {
		t.Error("Applying close attempts to a buffered channel should fail")
	}

	// Add invalid close attempt again
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericIssuerId, currentTime.Add(1 * time.Hour)},
	); !ok {
		t.Error("Attempting to close a buffered channel should not fail")
	}
	// Open channel
	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericUserId: {true, true, true},
		}},
		genericKeyId,
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
		if _, ok := rec.tryClose(
			&channelActionRecord{genericUserId, genericUserId, currentTime.Add(time.Duration(i) * time.Hour)},
		); !ok {
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
		genericKeyId,
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
	currentTime := time.Now()

	// Define valid add message action
	validAddMessageActionPtr := &channelActionRecord{
		issuerId:    genericIssuerId,
		certifierId: genericCertifierId,
		timestamp:   currentTime,
	}

	// Create buffered channel
	rec := &channelRecord{
		state: channelBufferedState,
	}

	// Adding message to buffered channel
	if _, ok := rec.addMessage(validAddMessageActionPtr); ok {
		t.Error("Adding message to a buffered channel should fail")
	}

	// Open channel (-1 hour)
	if !rec.tryOpen(
		genericChannelId,
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(-1 * time.Hour)},
		&channelPermissionsRecord{users: map[string]*channelPermissionRecord{
			genericCertifierId: {true, true, true},
		}},
		genericKeyId,
	) {
		t.Error("Opening a buffered channel should not fail")
	}

	// Adding 3 valid messages with 1 minute interval should not fail
	tmpTime := currentTime
	tmpChannelAction := &channelActionRecord{}
	*tmpChannelAction = *validAddMessageActionPtr
	for i := 0; i < 3; i++ {
		tmpChannelAction.timestamp = tmpTime
		if pos, ok := rec.addMessage(tmpChannelAction); !ok || pos != i {
			t.Error("Adding valid message to an open channel should not fail")
		}
		tmpTime = tmpTime.Add(time.Minute)
	}

	// Adding valid message with no permissions should fail
	*tmpChannelAction = *validAddMessageActionPtr
	tmpChannelAction.certifierId = genericIssuerId
	if _, ok := rec.addMessage(tmpChannelAction); ok {
		t.Error("Adding valid message without permissions should fail")
	}

	// Adding message with invalid timestamp should fail
	*tmpChannelAction = *validAddMessageActionPtr
	tmpChannelAction.timestamp = time.Time{}
	if _, ok := rec.addMessage(tmpChannelAction); ok {
		t.Error("Adding message with invalid timestamp should fail")
	}

	// Adding message before opening time
	*tmpChannelAction = *validAddMessageActionPtr
	tmpChannelAction.timestamp = currentTime.Add(-2 * time.Hour)
	if _, ok := rec.addMessage(tmpChannelAction); ok {
		t.Error("Adding message to open channel should fail if it's before opening time")
	}

	// Adding message at opening time
	*tmpChannelAction = *validAddMessageActionPtr
	tmpChannelAction.timestamp = currentTime.Add(-1 * time.Hour)
	if _, ok := rec.addMessage(tmpChannelAction); !ok {
		t.Error("Adding message to open channel should not fail if it's at opening time")
	}

	// Close channel (+1 hour)
	if _, ok := rec.tryClose(
		&channelActionRecord{genericIssuerId, genericCertifierId, currentTime.Add(1 * time.Hour)},
	); !ok {
		t.Error("Attempting to close an open channel should not fail")
	}

	// Adding message to closed channel before closing time (+1 second)
	*tmpChannelAction = *validAddMessageActionPtr
	tmpChannelAction.timestamp = currentTime.Add(time.Second)
	if pos, ok := rec.addMessage(tmpChannelAction); !ok || pos != 2 {
		t.Error("Adding message to closed channel should not fail if it's before closure")
	}

	// Adding message to closed channel at closing time
	*tmpChannelAction = *validAddMessageActionPtr
	tmpChannelAction.timestamp = currentTime.Add(time.Hour)
	if _, ok := rec.addMessage(tmpChannelAction); !ok {
		t.Error("Adding message to closed channel should not fail if it's at closure")
	}

	// Adding message to closed channel after closing time (+2 hour)
	*tmpChannelAction = *validAddMessageActionPtr
	tmpChannelAction.timestamp = currentTime.Add(2 * time.Hour)
	if _, ok := rec.addMessage(tmpChannelAction); ok {
		t.Error("Adding message to closed channel should fail if it's after closure")
	}
}
