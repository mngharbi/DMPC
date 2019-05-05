package pipeline

import (
	"errors"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/gofarm"
	"sync"
)

/*
	Errors
*/

var (
	serverNotRunning  error = errors.New("Pipeline not running during the operation.")
	subscriptionError error = errors.New("Failed to unsubscribe.")
)

/*
	Shared variables
*/
var (
	log             *core.LoggingHandler
	serverLock      *sync.RWMutex = &sync.RWMutex{}
	serverSingleton *server       = &server{}
)

/*
	Used to pass transaction to decryptor
*/
func passTransaction(transaction *core.Transaction) (channel chan *gofarm.Response, errs []error) {
	serverLock.RLock()
	defer serverLock.RUnlock()
	if !serverSingleton.isRunning {
		return nil, []error{serverNotRunning}
	}
	return serverSingleton.requester(transaction)
}

/*
	Used to get channel for status updates
*/
func getStatusUpdateChannel(ticket status.Ticket) (status.UpdateChannel, error) {
	serverLock.RLock()
	defer serverLock.RUnlock()
	if !serverSingleton.isRunning {
		return nil, serverNotRunning
	}

	return serverSingleton.statusSubscriber(ticket)
}

/*
	Used to do channel unsubscribe
*/
func doUnsubscribe(channelId string, subscriberId string) error {
	serverLock.RLock()
	defer serverLock.RUnlock()

	if !serverSingleton.isRunning {
		return serverNotRunning
	}

	channel, err := serverSingleton.unsubscriber(&channels.UnsubscribeRequest{
		ChannelId:    channelId,
		SubscriberId: subscriberId,
	})
	if err != nil {
		return err
	}

	resp, ok := <-channel
	if !ok {
		return subscriptionError
	}

	if resp == nil || resp.Result != channels.ListenersSuccess {
		return subscriptionError
	}
	return nil

}

/*
	Server API
*/

func StartServer(config Config, requester decryptor.Requester, unsubscriber channels.ListenersRequester, statusSubscriber status.Subscriber, loggingHandler *core.LoggingHandler) {
	if log == nil {
		log = loggingHandler
	}
	serverLock.Lock()
	serverSingleton.start(config, requester, unsubscriber, statusSubscriber)
	serverLock.Unlock()
}

func ShutdownServer() {
	serverLock.Lock()
	serverSingleton.shutdown()
	serverLock.Unlock()
}
