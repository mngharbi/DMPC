package channels

import (
	"testing"
)

func TestChannelsStartShutdown(t *testing.T) {
	if !resetAndStartChannelsServer(t, multipleWorkersChannelsConfig()) {
		return
	}
	shutdownChannelsServer()
}
