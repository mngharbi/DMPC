package channels

import (
	"testing"
)

func TestListenersStartShutdown(t *testing.T) {
	if !resetAndStartListenersServer(t, multipleWorkersListenersConfig()) {
		return
	}
	shutdownListenersServer()
}
