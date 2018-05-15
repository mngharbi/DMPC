package channels

import (
	"testing"
)

func TestMessagesStartShutdown(t *testing.T) {
	if !resetAndStartMessagesServer(t, multipleWorkersMessagesConfig()) {
		return
	}
	shutdownMessagesServer()
}
