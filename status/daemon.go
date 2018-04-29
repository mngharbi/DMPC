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
	log             *core.LoggingHandler
	shutdownProgram core.ShutdownLambda
)

/*
	Logging
*/
func StartServers(
	statusConf StatusServerConfig,
	listenersConf ListenersServerConfig,
	loggingHandler *core.LoggingHandler,
	shutdownLambda core.ShutdownLambda,
) error {
	log = loggingHandler
	shutdownProgram = shutdownLambda
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
