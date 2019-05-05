/*
	Test helpers
*/

package pipeline

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/gofarm"
	"math/rand"
	"net/url"
	"testing"
	"time"
)

/*
	Constants
*/

const (
	genericChannelId    string = "CHANNEL_ID"
	genericSubscriberId string = "SUBSCRIBER_ID"
)

/*
	Decryptor dummy
*/

func generateDecryptorRequester(formatSuccess bool, responseSuccess bool) decryptor.Requester {
	if formatSuccess {
		return func(*core.Transaction) (channel chan *gofarm.Response, errs []error) {
			result := decryptor.Success
			if !responseSuccess {
				result += 1
			}
			var resp gofarm.Response = &decryptor.DecryptorResponse{
				Result: result,
				Ticket: status.RequestNewTicket(),
			}
			channel = make(chan *gofarm.Response, 1)
			channel <- &resp
			return channel, nil
		}
	} else {
		return func(*core.Transaction) (channel chan *gofarm.Response, errs []error) {
			return nil, []error{errors.New("Decryptor request failed")}
		}
	}
}

/*
	Unsubscriber dummy
*/

func waitForRandomDuration() {
	duration := time.Duration(rand.Intn(100)) * time.Millisecond
	timer := time.NewTimer(duration)
	<-timer.C
}

func sendUnsubsriberResponseAfterRandomDelay(channel chan *channels.ListenersResponse, response channels.ListenersResponse) {
	waitForRandomDuration()
	listenersReponsePtr := &channels.ListenersResponse{}
	*listenersReponsePtr = response
	channel <- listenersReponsePtr
}

func createUnsubsriber(response channels.ListenersResponse, errReturned error, closeChannel bool) (channels.ListenersRequester, chan interface{}) {
	callsChannel := make(chan interface{}, 0)
	requester := func(request interface{}) (chan *channels.ListenersResponse, error) {
		go func() {
			callsChannel <- request
		}()
		if errReturned != nil {
			return nil, errReturned
		}
		responseChannel := make(chan *channels.ListenersResponse)
		if closeChannel {
			close(responseChannel)
		} else {
			go sendUnsubsriberResponseAfterRandomDelay(responseChannel, response)
		}
		return responseChannel, nil
	}
	return requester, callsChannel
}

func createSuccessUnsubsriber() (channels.ListenersRequester, chan interface{}) {
	return createUnsubsriber(channels.ListenersResponse{
		Result: channels.ListenersSuccess,
	}, nil, false)
}

func createSuccessUnsubsriberNoCalls() channels.ListenersRequester {
	unsubscriber, _ := createSuccessUnsubsriber()
	return unsubscriber
}

/*
	Channel response type
*/

type channelTestStruct struct {
	Channel      chan []byte
	ChannelId    string
	SubscriberId string
}

func (ch *channelTestStruct) GetResponse() ([]byte, bool) {
	resp, ok := <-ch.Channel
	return resp, ok
}

func (ch *channelTestStruct) GetSubscriberId() string {
	return ch.SubscriberId
}

func (ch *channelTestStruct) GetChannelId() string {
	return ch.ChannelId
}

/*
	Status listener dummy
*/

func createStatusSubscriber(errReturned error, eventFeeder func(status.Ticket) []*status.StatusRecord) (status.Subscriber, chan interface{}) {
	callsChannel := make(chan interface{}, 0)
	requester := func(ticket status.Ticket) (status.UpdateChannel, error) {
		go (func() {
			callsChannel <- ticket
		})()
		if errReturned != nil {
			return nil, errReturned
		}
		updateChannel := make(status.UpdateChannel)
		if eventFeeder != nil {
			go func() {
				events := eventFeeder(ticket)
				for _, event := range events {
					updateChannel <- event
				}
			}()
		}
		return updateChannel, nil
	}
	return requester, callsChannel
}

func createGenericStatusSubscriberNoCalls(eventFeeder func(status.Ticket) []*status.StatusRecord) status.Subscriber {
	statusSubscriber, _ := createStatusSubscriber(nil, eventFeeder)
	return statusSubscriber
}

func createSuccessStatusSubscriberNoCalls() status.Subscriber {
	return createGenericStatusSubscriberNoCalls(func(ticket status.Ticket) []*status.StatusRecord {
		return []*status.StatusRecord{
			generateQueuedUpdate(ticket),
			generateRunningUpdate(ticket),
			generateSuccessUpdate(ticket, []byte{32}),
		}
	})
}

/*
	Generators
*/

func generateQueuedUpdate(ticket status.Ticket) *status.StatusRecord {
	return &status.StatusRecord{
		Id:         ticket,
		Status:     status.QueuedStatus,
		FailReason: status.NoReason,
		Payload:    nil,
		Errs:       nil,
	}
}

func generateRunningUpdate(ticket status.Ticket) *status.StatusRecord {
	return &status.StatusRecord{
		Id:         ticket,
		Status:     status.RunningStatus,
		FailReason: status.NoReason,
		Payload:    nil,
		Errs:       nil,
	}
}

func generateSuccessUpdate(ticket status.Ticket, payload interface{}) *status.StatusRecord {
	return &status.StatusRecord{
		Id:         ticket,
		Status:     status.SuccessStatus,
		FailReason: status.NoReason,
		Payload:    payload,
		Errs:       nil,
	}
}

func generateRejectedUpdate(ticket status.Ticket) *status.StatusRecord {
	return &status.StatusRecord{
		Id:         ticket,
		Status:     status.FailedStatus,
		FailReason: status.RejectedReason,
	}
}

func generateFailedUpdate(ticket status.Ticket) *status.StatusRecord {
	return &status.StatusRecord{
		Id:         ticket,
		Status:     status.FailedStatus,
		FailReason: status.FailedReason,
	}
}

func generateValidTransactionJson(statusUpdates bool, result bool) []byte {
	encoded, _ := (&core.Transaction{
		Pipeline: core.PipelineConfig{
			ReadStatusUpdates: statusUpdates,
			ReadResult:        result,
		},
	}).Encode()
	return encoded
}

func generateInvalidTransactionJson() []byte {
	return []byte("}")
}

/*
	Websocket client
*/
const (
	defaultHostname string = "localhost"
	defaultPort     int    = 64928
	defaultPath     string = "/"
)

func openConnection(t *testing.T) *websocket.Conn {
	connUrl := url.URL{
		Scheme: "ws",
		Host:   makeAddrString(defaultHostname, defaultPort),
		Path:   defaultPath,
	}
	conn, _, err := websocket.DefaultDialer.Dial(connUrl.String(), nil)
	if err != nil {
		t.Errorf("Dialing error: %v", err)
		return nil
	}
	return conn
}

func sendMessage(t *testing.T, conn *websocket.Conn, msg []byte) bool {
	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		t.Errorf("Failed to send frame to websocket. Error=%v", err)
		return false
	}
	return true
}

func readMessage(t *testing.T, conn *websocket.Conn) []byte {
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read frame from websocket. Error=%v", err)
		return nil
	}
	return message
}

func closeConnection(t *testing.T, conn *websocket.Conn) bool {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		t.Errorf("Failed to send connection closure message to websocket. Error=%v", err)
		return false
	}
	return true
}

func waitForConnectionClosure(t *testing.T, conn *websocket.Conn) bool {
	// Send server side closure and timeout events to the same channel
	closeChannel := make(chan bool)
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				closeChannel <- true
				break
			}
		}
	}()
	timeout := time.After(5 * time.Second)
	go func() {
		<-timeout
		closeChannel <- false
	}()

	// Wait for first event
	serverClosed := <-closeChannel

	// See who won the race
	if !serverClosed {
		t.Errorf("Server didn't close connection before timeout.")
	}
	return serverClosed
}
