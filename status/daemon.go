/*
	Combined servers API
*/

package status

import (
	"github.com/mngharbi/DMPC/core"
	"sync"
)

var (
	serversStartWaitGroup sync.WaitGroup
)

/*
	Logging
*/
var (
	log *core.LoggingHandler
)

/*
	Logging
*/
func StartServers(loggingHandler *core.LoggingHandler, statusConf StatusServerConfig, listenersConf ListenersServerConfig) error {
	log = loggingHandler
	serversStartWaitGroup.Add(2)
	if err := startStatusServer(statusConf); err != nil {
		serversStartWaitGroup = sync.WaitGroup{}
		return err
	}
	if err := startListenersServer(listenersConf); err != nil {
		serversStartWaitGroup = sync.WaitGroup{}
		return err
	}
	serversStartWaitGroup.Wait()
	return nil
}

func ShutdownServers() {
	shutdownStatusServer()
	shutdownListenersServer()
}
