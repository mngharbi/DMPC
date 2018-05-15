/*
	Integration testing
	(includes all operations)
*/

package channels

import (
	"testing"
)

func TestStartShutdownServers(t *testing.T) {
	if !resetAndStartBothServers(t, multipleWorkersChannelsConfig(), multipleWorkersMessagesConfig(), multipleWorkersListenersConfig()) {
		return
	}
	ShutdownServers()
}
