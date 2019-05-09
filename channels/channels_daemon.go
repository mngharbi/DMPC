package channels

import (
	"errors"
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Server definitions
*/

type ChannelActionRequester func(request interface{}) (chan *ChannelsResponse, error)

type ChannelsServerConfig struct {
	NumWorkers int
}

type channelsServer struct {
	isInitialized bool
}

var (
	channelsServerSingleton channelsServer
	channelsServerHandler   *gofarm.ServerHandler
	channelsStore           *memstore.Memstore
)

/*
	Server API
*/

func provisionChannelsServerOnce() {
	if channelsServerHandler == nil {
		channelsServerHandler = gofarm.ProvisionServer()
	}
}

func startChannelsServer(conf ChannelsServerConfig, serversWaitGroup *sync.WaitGroup) (err error) {
	defer serversWaitGroup.Done()
	provisionChannelsServerOnce()
	if !channelsServerSingleton.isInitialized {
		channelsServerSingleton.isInitialized = true
		channelsServerHandler.ResetServer()
		channelsServerHandler.InitServer(&channelsServerSingleton)
	}
	return channelsServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func shutdownChannelsServer() {
	provisionChannelsServerOnce()
	channelsServerHandler.ShutdownServer()
}

/*
	Functional API
*/

func ChannelAction(request interface{}) (chan *ChannelsResponse, error) {
	// Sanitize and validate request
	var sanitizingErr error
	switch request.(type) {
	case *ReadChannelRequest:
		sanitizingErr = request.(*ReadChannelRequest).sanitizeAndValidate()
	case *OpenChannelRequest:
		sanitizingErr = request.(*OpenChannelRequest).sanitizeAndValidate()
	case *CloseChannelRequest:
		sanitizingErr = request.(*CloseChannelRequest).sanitizeAndValidate()
	default:
		return nil, errors.New("Unrecognized channel action")
	}
	if sanitizingErr != nil {
		return nil, sanitizingErr
	}

	// Make request to server
	nativeResponseChannel, err := channelsServerHandler.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	// Pass through result
	responseChannel := make(chan *ChannelsResponse)
	go func() {
		nativeResponse := <-nativeResponseChannel
		responseChannel <- (*nativeResponse).(*ChannelsResponse)
	}()

	return responseChannel, nil
}

/*
	Server implementation
*/

func (sv *channelsServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		channelsStore = memstore.New(getChannelIndexes())
	}
	log.Debugf(channelsDaemonStartLogMsg)
	return nil
}

func (sv *channelsServer) Shutdown() error {
	log.Debugf(channelsDaemonShutdownLogMsg)
	return nil
}

func (sv *channelsServer) Work(rqInterface *gofarm.Request) (dummyReturnVal *gofarm.Response) {
	log.Debugf(channelsRunningRequestLogMsg)

	resp := &ChannelsResponse{
		Result: ChannelsSuccess,
	}

	switch (*rqInterface).(type) {
	case *ReadChannelRequest:
		rq := (*rqInterface).(*ReadChannelRequest)

		// Get/Lock channel
		channelRecord := getChannel(channelsStore, rq.Id)
		if channelRecord == nil {
			resp.Result = ChannelsFailure
			break
		}
		channelRecord.RLock()
		defer func() { channelRecord.RUnlock() }()

		// Build object
		resp.Channel = &ChannelObject{}
		resp.Channel.buildFromRecord(channelRecord)

	case *OpenChannelRequest:
		rq := (*rqInterface).(*OpenChannelRequest)
		channelRecord := createOrGetChannel(channelsStore, rq.Channel.Id)
		channelRecord.Lock()
		defer func() { channelRecord.Unlock() }()

		// Try to open channel
		actionRecord := &channelActionRecord{
			issuerId:    rq.Signers.IssuerId,
			certifierId: rq.Signers.CertifierId,
			timestamp:   rq.Timestamp,
		}
		permissionsRecord := &channelPermissionsRecord{}
		permissionsRecord.build(&rq.Channel.Permissions)
		openSuccess := channelRecord.tryOpen(rq.Channel.Id, actionRecord, permissionsRecord, rq.Channel.KeyId)
		if !openSuccess {
			resp.Result = ChannelsFailure
			break
		}

		// Feed buffered operations into decryptor
		// @TODO: move buffer queuer outside of server object
		// @TODO: make requests concurrently
		channelBuffer := createOrGetChannelBuffer(bufferStore, rq.Channel.Id)
		channelBuffer.Lock()
		defer func() { channelBuffer.Unlock() }()
		for _, operationPtr := range channelBuffer.operations {
			respChannel, _ := messagesServerSingleton.operationQueuer(operationPtr)
			_, ok := <-respChannel
			if !ok {
				// This occurs when decryptor is shut down prematurely
				resp.Result = BufferError
				break
			}
		}
		if resp.Result == BufferError {
			break
		}

		// Empty buffer
		channelBuffer.operations = nil

		// Remove unauthorized listeners
		unsubscribeUnauthorized(channelRecord.id, channelRecord.permissions)

		// Notify (early) listeners of channel opening
		publish(rq.Channel.Id, makeOpenEvent(channelRecord.duration.opened))

		// Apply early closures
		if channelRecord.applyCloseAttempts() {
			publish(rq.Channel.Id, makeCloseEvent(channelRecord.duration.closed, 0))
		}

	case *CloseChannelRequest:
		rq := (*rqInterface).(*CloseChannelRequest)

		// Get/Lock channel
		channelRecord := createOrGetChannel(channelsStore, rq.Id)
		channelRecord.Lock()
		defer func() { channelRecord.Unlock() }()

		// Try to close channel
		actionRecord := &channelActionRecord{
			issuerId:    rq.Signers.IssuerId,
			certifierId: rq.Signers.CertifierId,
			timestamp:   rq.Timestamp,
		}
		remainingMessages, closeSuccess := channelRecord.tryClose(actionRecord)
		if !closeSuccess {
			resp.Result = ChannelsFailure
			break
		}

		// Only notify if channel is closed now
		if channelRecord.state == channelClosedState {
			publish(rq.Id, makeCloseEvent(channelRecord.duration.closed, remainingMessages))
		}
	}

	var respInterface gofarm.Response = resp
	return &respInterface
}
