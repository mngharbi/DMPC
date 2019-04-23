/*
	Combined servers API
*/

package channels

import (
	"github.com/mngharbi/DMPC/core"
	"sync"
)

var (
	log             *core.LoggingHandler
	shutdownProgram core.ShutdownLambda
)

/*
	Servers API
*/
func StartServers(
	channelsConfig ChannelsServerConfig,
	messagesConfig MessagesServerConfig,
	listenersConfig ListenersServerConfig,
	operationQueuer core.OperationQueuer,
	loggingHandler *core.LoggingHandler,
	shutdownLambda core.ShutdownLambda,
) error {
	log = loggingHandler
	shutdownProgram = shutdownLambda
	serversWaitGroup := &sync.WaitGroup{}
	serversWaitGroup.Add(3)
	if err := startChannelsServer(channelsConfig, serversWaitGroup); err != nil {
		return err
	}
	if err := startMessagesServer(messagesConfig, operationQueuer, serversWaitGroup); err != nil {
		return err
	}
	if err := startListenersServer(listenersConfig, serversWaitGroup); err != nil {
		return err
	}
	serversWaitGroup.Wait()
	return nil
}

func ShutdownServers() {
	shutdownListenersServer()
	shutdownMessagesServer()
	shutdownChannelsServer()
}
