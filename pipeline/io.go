package pipeline

import (
	"github.com/gorilla/websocket"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/status"
	"io"
	"sync"
)

type Conversation struct {
	socket        *websocket.Conn
	outgoingQueue chan status.Ticket
	lock          *sync.Mutex
}

func closeConnectionForInvalidData(c *Conversation) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, ""))
}

func (c *Conversation) reader() {
	for {
		var transaction core.Transaction
		if err := c.socket.ReadJSON(&transaction); err == io.EOF {
			return
		} else if err != nil {
			log.Debugf(invalidOperationLogMsg)
			closeConnectionForInvalidData(c)
			return
		} else {
			// Wait for ticket and push to outgoing queue in another goroutine
			if channel, errs := passOperation(&transaction); errs == nil {
				go func() {
					nativeResp := <-channel
					if nativeResp != nil {
						resp := (*nativeResp).(*decryptor.DecryptorResponse)
						if resp.Result == decryptor.Success {
							c.outgoingQueue <- resp.Ticket
						} else {
							closeConnectionForInvalidData(c)
						}
					}
				}()
			} else {
				closeConnectionForInvalidData(c)
			}
		}
	}
}

func (c *Conversation) writer() {
	for ticket := range c.outgoingQueue {
		c.lock.Lock()
		err := c.socket.WriteJSON(string(ticket))
		c.lock.Unlock()
		if err != nil {
			closeConnectionForInvalidData(c)
			return
		}
	}
}

func NewConversation(socket *websocket.Conn) {
	c := &Conversation{
		socket:        socket,
		outgoingQueue: make(chan status.Ticket),
		lock:          &sync.Mutex{},
	}

	go c.reader()
	go c.writer()
}
