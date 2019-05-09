package cli

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

const (
	pipelinePath string = "/"
)

/*
	Pipeline client API
*/

func openConnection() *websocket.Conn {
	// Get configuration structure
	conf := GetConfig()

	connUrl := url.URL{
		Scheme: "ws",
		Host:   makeAddrString(conf.Pipeline.Hostname, conf.Pipeline.Port),
		Path:   pipelinePath,
	}
	conn, _, err := websocket.DefaultDialer.Dial(connUrl.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to pipeline. error: %v", err)
		return nil
	}
	return conn
}

func doSendMessage(conn *websocket.Conn, msg []byte) error {
	return conn.WriteMessage(websocket.TextMessage, msg)
}

func sendMessage(conn *websocket.Conn, msg []byte) bool {
	err := doSendMessage(conn, msg)
	if err != nil {
		log.Fatalf("Failed to send frame to websocket. Error=%v", err)
		return false
	}
	return true
}

func doReadMessage(conn *websocket.Conn) ([]byte, error) {
	_, message, err := conn.ReadMessage()
	return message, err
}

func readMessage(conn *websocket.Conn) []byte {
	message, err := doReadMessage(conn)
	if err != nil {
		log.Fatalf("Failed to read frame from websocket. Error=%v", err)
		return nil
	}
	return message
}

func doCloseConnection(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
}

func closeConnection(conn *websocket.Conn) bool {
	err := doCloseConnection(conn)
	if err != nil {
		log.Fatalf("Failed to send connection closure message to websocket. Error=%v", err)
		return false
	}
	return true
}

/*
	Helpers
*/

func makeAddrString(hostname string, port int) string {
	return fmt.Sprintf("%v:%v", hostname, port)
}

func drainResponses(conn *websocket.Conn, incoming chan []byte) {
	for {
		message, err := doReadMessage(conn)
		if err != nil || len(message) == 0 {
			close(incoming)
			break
		} else {
			incoming <- message
		}
	}
}

func interactWithPipeline(incoming chan []byte, outgoing chan []byte) {
	conn := openConnection()
	defer func() {
		closeConnection(conn)
	}()

	// Read and push into incoping pipeline in separate go routine
	go drainResponses(conn, incoming)

	for msg := range outgoing {
		if err := doSendMessage(conn, msg); err != nil {
			break
		}
	}
}

func runAndWrite(outgoing chan []byte) {
	// Make incoming channel and start writing/reading to pipeline
	incoming := make(chan []byte)
	go interactWithPipeline(incoming, outgoing)

	// Write everything to stdout
	for msg := range incoming {
		fmt.Println(string(msg))
	}
}

func runOneTransactionAndWrite(transactionEncoded []byte) {
	// Make outgoing channel and push transaction into it
	outgoing := make(chan []byte, 1)
	outgoing <- transactionEncoded

	runAndWrite(outgoing)
}
