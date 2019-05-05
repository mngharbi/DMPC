package channels

import (
	"errors"
	"github.com/mngharbi/gofarm"
	"sync"
)

/*
	Server definitions
*/

type ListenersRequester func(request interface{}) (chan *ListenersResponse, error)

type ListenersServerConfig struct {
	NumWorkers int
}

type listenersServer struct {
	isInitialized bool
}

var (
	listenersServerSingleton listenersServer
	listenersServerHandler   *gofarm.ServerHandler
)

/*
	Server API
*/

func provisionListenersServerOnce() {
	if listenersServerHandler == nil {
		listenersServerHandler = gofarm.ProvisionServer()
	}
}

func startListenersServer(conf ListenersServerConfig, serversWaitGroup *sync.WaitGroup) (err error) {
	defer serversWaitGroup.Done()
	provisionListenersServerOnce()
	if !listenersServerSingleton.isInitialized {
		listenersServerSingleton.isInitialized = true
		listenersServerHandler.ResetServer()
		listenersServerHandler.InitServer(&listenersServerSingleton)
	}
	return listenersServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func shutdownListenersServer() {
	provisionListenersServerOnce()
	listenersServerHandler.ShutdownServer()
}

/*
	Functional API
*/

func ListenerAction(request interface{}) (chan *ListenersResponse, error) {
	// Sanitize and validate request
	var sanitizingErr error
	switch request.(type) {
	case *SubscribeRequest:
		sanitizingErr = request.(*SubscribeRequest).sanitizeAndValidate()
	case *UnsubscribeRequest:
		sanitizingErr = nil
	default:
		return nil, errors.New("Unrecognized listeners action")
	}
	if sanitizingErr != nil {
		return nil, sanitizingErr
	}

	// Make request to server
	nativeResponseChannel, err := listenersServerHandler.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	// Pass through result
	responseChannel := make(chan *ListenersResponse)
	go func() {
		nativeResponse := <-nativeResponseChannel
		responseChannel <- (*nativeResponse).(*ListenersResponse)
	}()

	return responseChannel, nil
}

/*
	Server implementation
*/

func (sv *listenersServer) Start(_ gofarm.Config, isFirstStart bool) error {
	if isFirstStart {
		listenersStore = &sync.Map{}
	}
	log.Debugf(listenersDaemonStartLogMsg)
	return nil
}

func (sv *listenersServer) Shutdown() error {
	log.Debugf(listenersDaemonShutdownLogMsg)
	return nil
}

func (sv *listenersServer) Work(rqInterface *gofarm.Request) *gofarm.Response {
	log.Debugf(listenersRunningRequestLogMsg)

	resp := &ListenersResponse{
		Result: ListenersSuccess,
	}

	switch (*rqInterface).(type) {
	case *SubscribeRequest:
		rq := (*rqInterface).(*SubscribeRequest)

		resp.ChannelId = rq.ChannelId

		// Get/Lock channel record
		channelRecord := createOrGetChannel(channelsStore, rq.ChannelId)
		channelRecord.Lock()
		defer func() { channelRecord.Unlock() }()

		// Check certifier read permissions
		authorized := true
		if channelRecord.permissions != nil {
			certifierPermissions, certifierFound := channelRecord.permissions.users[rq.Signers.CertifierId]
			authorized = certifierFound && certifierPermissions.read
		}
		if !authorized {
			resp.Result = ListenersUnauthorized
			break
		}

		resp.Channel, resp.SubscriberId = subscribe(rq.ChannelId, rq.Signers.CertifierId)

	case *UnsubscribeRequest:
		rq := (*rqInterface).(*UnsubscribeRequest)

		resp.ChannelId = rq.ChannelId

		// Get/Lock channel record
		channelRecord := createOrGetChannel(channelsStore, rq.ChannelId)
		channelRecord.Lock()
		defer func() { channelRecord.Unlock() }()

		if err := unsubscribe(rq.ChannelId, rq.SubscriberId); err != nil {
			resp.Result = ListenersFailure
			break
		}
	}

	var nativeResponse gofarm.Response = resp
	return &nativeResponse
}
