package channels

import (
	"testing"
)

const (
	genericChannelId string = "channelId"
)

func TestListenersStartShutdown(t *testing.T) {
	if !resetAndStartListenersServer(t, multipleWorkersListenersConfig()) {
		return
	}
	shutdownListenersServer()
}

func TestListenersServerDown(t *testing.T) {
	if ch, err := AddListener(genericChannelId); ch != nil || err == nil {
		t.Error("Adding listener while server is down should fail.")
	}
}

func TestListenersAddEarly(t *testing.T) {
	if !resetAndStartListenersServer(t, multipleWorkersListenersConfig()) {
		return
	}
	defer shutdownListenersServer()

	// Add early listeners
	numListeners := 5
	for i := 0; i < numListeners; i++ {
		_, err := AddListener(genericChannelId)
		if err != nil {
			t.Errorf("Adding early listeners should not fail. err=%+v", err)
		}
	}

	// Expect it to be added to store
	listenersRecordFound := getListenersRecordById(listenersStore, genericChannelId)
	if listenersStore.Len() != 1 ||
		listenersRecordFound == nil ||
		len(listenersRecordFound.channels) != numListeners {
		t.Error("Early listeners should be added to store.")
	}
}
