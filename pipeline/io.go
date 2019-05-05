package pipeline

import (
	"github.com/gorilla/websocket"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/status"
	"io"
	"sync"
)

/*
	Conversation definition
*/

type Conversation struct {
	socket                   *websocket.Conn
	incomingQueue            chan *core.Transaction
	quitChannel              chan bool
	closing                  bool
	transactionConversations []*TransactionConversation
	lock                     *sync.Mutex
}

func NewConversation(socket *websocket.Conn) {
	c := &Conversation{
		socket:                   socket,
		incomingQueue:            make(chan *core.Transaction),
		quitChannel:              make(chan bool, 1),
		closing:                  false,
		transactionConversations: []*TransactionConversation{},
		lock:                     &sync.Mutex{},
	}

	go c.reader()
	go c.dispatcher()
}

/*
	Conversation daemons
*/

func (c *Conversation) reader() {
	for {
		transaction := c.read()
		if transaction == nil {
			close(c.incomingQueue)
			break
		}
		log.Debugf(readTransactionLogMsg)
		c.incomingQueue <- transaction
	}
}

func (c *Conversation) dispatcher() {
	for transaction := range c.incomingQueue {
		c.dispatch(transaction)
	}
	c.informQuit()
}

/*
	Helpers
*/

func (c *Conversation) doClose() {
	if c.closing {
		return
	}
	c.closing = true
	c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, ""))
	c.quitChannel <- true
}

func (c *Conversation) close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.doClose()
}

func (c *Conversation) write(data []byte) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.socket.WriteMessage(websocket.TextMessage, data); err != nil {
		c.doClose()
	}
}

func (c *Conversation) read() *core.Transaction {
	transaction := &core.Transaction{}
	if err := c.socket.ReadJSON(transaction); err == io.EOF {
		c.close()
		return nil
	} else if err != nil {
		log.Infof(invalidTransactionLogMsg)
		c.close()
		return nil
	}
	return transaction
}

func (c *Conversation) dispatch(transaction *core.Transaction) {
	c.lock.Lock()
	defer c.lock.Unlock()
	tc := &TransactionConversation{
		parentConversation: c,
		transaction:        transaction,
		quitChannel:        make(chan bool, 1),
		doneChannel:        make(chan bool, 1),
	}
	c.transactionConversations = append(c.transactionConversations, tc)
	go tc.writer()
}

func (c *Conversation) informQuit() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, transactionConversations := range c.transactionConversations {
		transactionConversations.quitChannel <- true
	}
}

/*
	Transaction conversation
*/

type TransactionConversation struct {
	parentConversation *Conversation

	transaction *core.Transaction

	// Pushed from outside the transaction conversation
	quitChannel chan bool

	// Pushes the last update when transaction updates are done
	doneChannel chan bool
}

func (tc *TransactionConversation) responseDrainer(lastStatusUpdate *status.StatusRecord) {
	for {
		encoded, open := lastStatusUpdate.GetResponse()
		if encoded != nil && tc.transaction.Pipeline.ReadResult {
			tc.parentConversation.write(encoded)
		}
		if !open {
			tc.doneChannel <- true
			close(tc.doneChannel)
			break
		}
	}
}

func (tc *TransactionConversation) writer() {
	// Make request to executor
	channel, errs := passTransaction(tc.transaction)
	if errs != nil {
		tc.parentConversation.close()
		log.Debugf(transactionRejected, errs)
		return
	}

	// Wait for ticket
	nativeResp := <-channel
	if nativeResp == nil {
		tc.parentConversation.close()
		log.Debugf(invalidDecryptorResponse, nativeResp)
		return
	}
	resp := (*nativeResp).(*decryptor.DecryptorResponse)
	if resp.Result != decryptor.Success {
		tc.parentConversation.close()
		return
	}
	ticket := resp.Ticket

	// Listen to updates on ticket
	updateChannel, err := getStatusUpdateChannel(ticket)
	if err != nil {
		log.Debugf(updateChannelFailureLogMsg, ticket, err)
		return
	}

	// Read and write updates
	var lastStatusUpdate *status.StatusRecord
	for lastStatusUpdate = range updateChannel {
		if lastStatusUpdate.IsDone() {
			break
		}
		if tc.transaction.Pipeline.ReadStatusUpdates {
			encoded, _ := lastStatusUpdate.GetResponse()
			tc.parentConversation.write(encoded)
		}
	}

	// Drain results
	go tc.responseDrainer(lastStatusUpdate)

	// Wait for either results to be done or quit signal
	select {
	case <-tc.doneChannel:
		break
	case <-tc.quitChannel:
		break
	}

	// Once done, possibly unsubscribe
	subscriberId, hasSubscriber := lastStatusUpdate.GetSubscriberId()
	if hasSubscriber {
		channelId, _ := lastStatusUpdate.GetChannelId()
		if err = doUnsubscribe(channelId, subscriberId); err != nil {
			log.Debugf(unsubscribeFailedLogMsg, ticket, err)
			return
		}
	}
}
