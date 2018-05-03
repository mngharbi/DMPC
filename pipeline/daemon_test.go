package pipeline

import (
	"testing"
	//"time"
	"sync"
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
		log,
	)
	ShutdownServer()
}

func testValidOperations(t *testing.T, waitGroup *sync.WaitGroup) {
	conn := openConnection(t)
	defer waitGroup.Done()
	if conn == nil {
		return
	}
	if !sendMessage(t, conn, generateValidOperationJson()) {
		return
	}
	if msg := readMessage(t, conn); msg == nil {
		return
	}
	if !sendMessage(t, conn, generateValidOperationJson()) {
		return
	}
	if msg := readMessage(t, conn); msg == nil {
		return
	}
	if !closeConnection(t, conn) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}
}

func TestSuccessfulOperations(t *testing.T) {
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		log,
	)

	// Open connections concurrently and send correct operation then expect closure
	waitGroup := &sync.WaitGroup{}
	concurrentConnections := 30
	waitGroup.Add(concurrentConnections)
	for i := 0; i < concurrentConnections; i++ {
		go testValidOperations(t, waitGroup)
	}
	waitGroup.Wait()

	ShutdownServer()
}

func TestInvalidOperation(t *testing.T) {
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, true),
		log,
	)

	// Make valid operation and expect to return ticket
	conn := openConnection(t)
	if conn == nil {
		return
	}
	if !sendMessage(t, conn, generateValidOperationJson()) {
		return
	}
	if msg := readMessage(t, conn); msg == nil {
		return
	}

	// Make invalid operation and expect server to close connection
	if !sendMessage(t, conn, generateInvalidOperationJson()) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}

	ShutdownServer()
}

func TestRejectedOperation(t *testing.T) {
	// Test that an operation rejected by decryptor requester are rejected and connection is closed
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(false, true),
		log,
	)

	conn := openConnection(t)
	if conn == nil {
		return
	}
	if !sendMessage(t, conn, generateValidOperationJson()) {
		return
	}
	if !waitForConnectionClosure(t, conn) {
		return
	}

	ShutdownServer()

	// Test that an operation that failed decryption is rejected and connection is closed
	StartServer(
		Config{
			CheckOrigin: false,
			Hostname:    defaultHostname,
			Port:        defaultPort,
		},
		generateDecryptorRequester(true, false),
		log,
	)

	conn = openConnection(t)
	if conn == nil {
		return
	}
	if !sendMessage(t, conn, generateValidOperationJson()) {
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
		log,
	)

	conn = openConnection(t)
	if conn == nil {
		return
	}
	if !sendMessage(t, conn, generateValidOperationJson()) {
		return
	}
	if msg := readMessage(t, conn); msg == nil {
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
