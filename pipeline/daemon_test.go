package pipeline

import (
	"github.com/gorilla/websocket"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/status"
	"reflect"
	"sync"
	"testing"
	"time"
)

/*
	General tests
*/

func TestStartShutdownServer(t *testing.T) {
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		createSuccessUnsubsriberNoCalls(),
		createSuccessStatusSubscriberNoCalls(),
		log,
	)
	ShutdownServer()
}

func testValidTransactionWithConn(t *testing.T, conn *websocket.Conn, statusUpdates bool, result bool) bool {
	expectedResponses := 0
	if statusUpdates {
		expectedResponses = 2
	}
	if result {
		expectedResponses++
	}

	if !sendMessage(t, conn, generateValidTransactionJson(statusUpdates, result)) {
		return false
	}
	for i := 0; i < expectedResponses; i++ {
		if msg := readMessage(t, conn); msg == nil {
			return false
		}
	}

	return true
}

func test2ValidTransactions(t *testing.T, waitGroup *sync.WaitGroup, statusUpdates bool, result bool) {
	conn := openConnection(t)
	defer waitGroup.Done()
	if conn == nil {
		return
	}

	if !testValidTransactionWithConn(t, conn, statusUpdates, result) {
		return
	}

	if !testValidTransactionWithConn(t, conn, statusUpdates, result) {
		return
	}

	if !closeConnection(t, conn) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}
}

func TestSuccessfulConcurrentTransactions(t *testing.T) {
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		createSuccessUnsubsriberNoCalls(),
		createSuccessStatusSubscriberNoCalls(),
		log,
	)

	// Open connections concurrently and send correct operation then expect closure
	waitGroup := &sync.WaitGroup{}
	concurrentConnections := 1
	waitGroup.Add(concurrentConnections)
	for i := 0; i < concurrentConnections; i++ {
		go test2ValidTransactions(t, waitGroup, true, true)
	}
	waitGroup.Wait()

	ShutdownServer()
}

func TestSuccessfulTransactions(t *testing.T) {
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		createSuccessUnsubsriberNoCalls(),
		createSuccessStatusSubscriberNoCalls(),
		log,
	)

	conn := openConnection(t)
	if conn == nil {
		return
	}

	// Silent transaction
	if !testValidTransactionWithConn(t, conn, false, false) {
		return
	}

	// Expect both status update and result
	if !testValidTransactionWithConn(t, conn, true, true) {
		return
	}

	// Expect only status updates
	if !testValidTransactionWithConn(t, conn, true, false) {
		return
	}

	// Expect only result
	if !testValidTransactionWithConn(t, conn, false, true) {
		return
	}

	if !closeConnection(t, conn) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}

	ShutdownServer()
}

func TestChannelResponse(t *testing.T) {
	channelStruct := &channelTestStruct{
		Channel:      make(chan []byte),
		ChannelId:    genericChannelId,
		SubscriberId: genericSubscriberId,
	}

	unsubscriber, unsubscriberCalls := createSuccessUnsubsriber()

	subscriber := createGenericStatusSubscriberNoCalls(func(ticket status.Ticket) []*status.StatusRecord {
		return []*status.StatusRecord{
			generateQueuedUpdate(ticket),
			generateRunningUpdate(ticket),
			generateSuccessUpdate(ticket, channelStruct),
		}
	})

	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		unsubscriber,
		subscriber,
		log,
	)

	conn := openConnection(t)
	if conn == nil {
		return
	}

	// Send valid transaction with status and result requested
	if !sendMessage(t, conn, generateValidTransactionJson(true, true)) {
		return
	}

	// Expect queued and running status update
	for i := 0; i < 2; i++ {
		if msg := readMessage(t, conn); msg == nil {
			return
		}
	}

	// Concurrently send 100 messages into channel
	for i := 0; i < 100; i++ {
		go func(i int) {
			channelStruct.Channel <- []byte{byte(i)}
		}(i)
	}

	// Expect to read those messages from socket
	for i := 0; i < 100; i++ {
		if msg := readMessage(t, conn); msg == nil {
			return
		}
	}

	// Expect drainer to be blocking
	select {
	case <-unsubscriberCalls:
		t.Error("Unexpected unsubscribe call")
	default:
		break
	}

	// Close channel and expect it to unblock and call unsubscribe
	close(channelStruct.Channel)

	// Expect drainer to unblock
	timer := time.NewTimer(time.Second)
	select {
	case unsubscribeCallInterfaced := <-unsubscriberCalls:
		unsubscribeCall := unsubscribeCallInterfaced.(*channels.UnsubscribeRequest)
		expectedUnsubscribeCall := &channels.UnsubscribeRequest{
			ChannelId:    genericChannelId,
			SubscriberId: genericSubscriberId,
		}
		if !reflect.DeepEqual(expectedUnsubscribeCall, unsubscribeCall) {
			t.Errorf("Expected unsubscribe call to match. expected=%+v found=%+v", expectedUnsubscribeCall, unsubscribeCall)
		}
	case <-timer.C:
		t.Error("Expected unsubscribe call to be made")
	}

	ShutdownServer()
}

func TestInvalidTransaction(t *testing.T) {
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		createSuccessUnsubsriberNoCalls(),
		createSuccessStatusSubscriberNoCalls(),
		log,
	)

	// Make valid operation
	conn := openConnection(t)
	if conn == nil {
		return
	}
	if !testValidTransactionWithConn(t, conn, true, true) {
		return
	}

	// Make invalid operation and expect server to close connection
	if !sendMessage(t, conn, generateInvalidTransactionJson()) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}

	ShutdownServer()
}

func TestRejectedTransaction(t *testing.T) {
	// Test that a transaction rejected by decryptor requester is handled correctly
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(false, true),
		createSuccessUnsubsriberNoCalls(),
		createSuccessStatusSubscriberNoCalls(),
		log,
	)

	conn := openConnection(t)
	if conn == nil {
		return
	}
	if !sendMessage(t, conn, generateValidTransactionJson(true, true)) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}

	ShutdownServer()
}

func TestDecryptionFailure(t *testing.T) {
	// Test that an operation that failed decryption is rejected and connection is closed
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, false),
		createSuccessUnsubsriberNoCalls(),
		createSuccessStatusSubscriberNoCalls(),
		log,
	)

	conn := openConnection(t)
	if conn == nil {
		return
	}
	if !sendMessage(t, conn, generateValidTransactionJson(true, true)) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}

	ShutdownServer()

	// Test that a valid operation goes through (to test that restarting resets requester closure)
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		createSuccessUnsubsriberNoCalls(),
		createSuccessStatusSubscriberNoCalls(),
		log,
	)

	conn = openConnection(t)
	if conn == nil {
		return
	}
	if !testValidTransactionWithConn(t, conn, true, true) {
		return
	}
	if !closeConnection(t, conn) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}

	ShutdownServer()
}
