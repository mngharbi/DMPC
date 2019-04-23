/*
	Integration testing
	(includes all operations)
*/

package channels

import (
	"github.com/mngharbi/DMPC/status"
	"testing"
)

func TestStartShutdownServers(t *testing.T) {
	operationQueuerDummy, _ := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartBothServers(t, multipleWorkersChannelsConfig(), multipleWorkersMessagesConfig(), multipleWorkersListenersConfig(), operationQueuerDummy) {
		return
	}
	ShutdownServers()
}
