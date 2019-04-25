package channels

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/gofarm"
	"github.com/mngharbi/memstore"
	"sync"
)

/*
	Server definitions
*/

type MessageAdder func(request *AddMessageRequest) (chan *MessagesResponse, error)
type OperationBufferer func(request *BufferOperationRequest) (chan *MessagesResponse, error)

type MessagesServerConfig struct {
	NumWorkers int
}

type messagesServer struct {
	isInitialized   bool
	operationQueuer core.OperationQueuer
}

var (
	messagesServerSingleton messagesServer
	messagesServerHandler   *gofarm.ServerHandler
	bufferStore             *memstore.Memstore
)

/*
	Server API
*/

func provisionMessagesServerOnce() {
	if messagesServerHandler == nil {
		messagesServerHandler = gofarm.ProvisionServer()
	}
}

func startMessagesServer(conf MessagesServerConfig, operationQueuer core.OperationQueuer, serversWaitGroup *sync.WaitGroup) (err error) {
	defer serversWaitGroup.Done()
	provisionMessagesServerOnce()
	if !messagesServerSingleton.isInitialized {
		messagesServerSingleton.isInitialized = true
		messagesServerSingleton.operationQueuer = operationQueuer
		messagesServerHandler.ResetServer()
		messagesServerHandler.InitServer(&messagesServerSingleton)
	}
	return messagesServerHandler.StartServer(gofarm.Config{NumWorkers: conf.NumWorkers})
}

func shutdownMessagesServer() {
	provisionMessagesServerOnce()
	messagesServerHandler.ShutdownServer()
}

/*
	Functional API
*/

func genericPassthroughRequest(request interface{}) (chan *MessagesResponse, error) {
	// Make request to server
	nativeResponseChannel, err := messagesServerHandler.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	// Pass through result
	responseChannel := make(chan *MessagesResponse)
	go func() {
		nativeResponse := <-nativeResponseChannel
		responseChannel <- (*nativeResponse).(*MessagesResponse)
	}()

	return responseChannel, nil
}

func AddMessage(request *AddMessageRequest) (chan *MessagesResponse, error) {
	// Sanitize and validate request
	err := request.sanitizeAndValidate()
	if err != nil {
		return nil, err
	}

	return genericPassthroughRequest(request)
}

func BufferOperation(request *BufferOperationRequest) (chan *MessagesResponse, error) {
	// Sanitize and validate request
	err := request.sanitizeAndValidate()
	if err != nil {
		return nil, err
	}

	return genericPassthroughRequest(request)
}

/*
	Server implementation
*/

func (sv *messagesServer) Start(_ gofarm.Config, isFirstStart bool) error {
	// Initialize store (only if starting for the first time)
	if isFirstStart {
		bufferStore = memstore.New(getChannelBufferIndexes())
	}
	log.Debugf(messagesDaemonStartLogMsg)
	return nil
}

func (sv *messagesServer) Shutdown() error {
	log.Debugf(messagesDaemonShutdownLogMsg)
	return nil
}

func (sv *messagesServer) Work(rqInterface *gofarm.Request) *gofarm.Response {
	log.Debugf(messagesRunningRequestLogMsg)

	var statusCode MessagesStatusCode = MessagesSuccess

	switch (*rqInterface).(type) {
	case *AddMessageRequest:
		rq := (*rqInterface).(*AddMessageRequest)

		// Get/Lock channel
		channelRecord := createOrGetChannel(channelsStore, rq.ChannelId)
		channelRecord.Lock()
		defer func() { channelRecord.Unlock() }()

		// Try to add message
		actionRecord := &channelActionRecord{
			issuerId:    rq.Signers.IssuerId,
			certifierId: rq.Signers.CertifierId,
			timestamp:   rq.Timestamp,
		}
		messagePosition, addSuccess := channelRecord.addMessage(actionRecord)
		if !addSuccess {
			statusCode = MessagesDropped
			break
		}

		// Notify listeners of message
		notifyListeners(listenersStore, rq.ChannelId, makeMessageEvent(rq.Timestamp, messagePosition, rq.Message))

	case *BufferOperationRequest:
	}

	var resp gofarm.Response = &MessagesResponse{
		Result: statusCode,
	}
	return &resp
}
