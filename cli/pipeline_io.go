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

func sendMessage(conn *websocket.Conn, msg []byte) bool {
	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Fatalf("Failed to send frame to websocket. Error=%v", err)
		return false
	}
	return true
}

func readMessage(conn *websocket.Conn) []byte {
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("Failed to read frame from websocket. Error=%v", err)
		return nil
	}
	return message
}

func closeConnection(conn *websocket.Conn) bool {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
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

func makeTransactionAndGetResult(transactionEncoded []byte) []byte {
	conn := openConnection()
	sendMessage(conn, transactionEncoded)
	defer func() {
		closeConnection(conn)
	}()
	return readMessage(conn)
}
