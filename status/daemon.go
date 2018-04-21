/*
	Combined servers API
*/

package status

import (
	"sync"
)

var (
	serversStartWaitGroup sync.WaitGroup
)

func StartServers(statusConf StatusServerConfig, listenersConf ListenersServerConfig) error {
	serversStartWaitGroup.Add(2)
	if err := StartStatusServer(statusConf); err != nil {
		serversStartWaitGroup = sync.WaitGroup{}
		return err
	}
	if err := StartListenersServer(listenersConf); err != nil {
		serversStartWaitGroup = sync.WaitGroup{}
		return err
	}
	serversStartWaitGroup.Wait()
	return nil
}

func ShutdownServers() {
	ShutdownStatusServer()
	ShutdownListenersServer()
}
