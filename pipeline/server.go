package pipeline

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mngharbi/DMPC/decryptor"
	"net"
	"net/http"
)

/*
	Server configuration
*/
type Config struct {
	CheckOrigin bool
	Hostname    string
	Port        int
}

/*
	Server structure
*/
type server struct {
	isRunning     bool
	isInitialized bool
	handler       *http.Server
	listener      net.Listener
	requester     decryptor.Requester
}

/*
	Resets listener and handlers
*/
func (sv *server) reset(config Config, requester decryptor.Requester) {
	// Initialize handler
	if !sv.isInitialized {
		upgrader := makeUpgrader(config)

		// Upgrade HTTP requests to websockets and start conversation
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			socket, _ := upgrader.Upgrade(w, r, nil)
			NewConversation(socket)
		})
	}
	sv.isInitialized = true

	// Make server handler
	addrString := config.makeAddrString()
	serverHandler := &http.Server{
		Addr: addrString,
	}
	sv.handler = serverHandler
	sv.requester = requester

	// Server should start listening on address
	var err error
	sv.listener, err = net.Listen("tcp", addrString)
	if err != nil {
		log.Fatalf("Pipeline server could not start listening on %v. Error: %v", addrString, err)
	}

	// Mark as running
	sv.isRunning = true

	// Start serving in separate goroutine
	go serverHandler.Serve(sv.listener)
}

/*
	Starts server by resetting it if it's not already running
*/
func (sv *server) start(config Config, requester decryptor.Requester) {
	if !sv.isRunning {
		sv.reset(config, requester)
	}
}

/*
	Shuts down server if it's running
*/
func (sv *server) shutdown() {
	if sv.isRunning {
		sv.isRunning = false
		sv.requester = nil
		sv.handler.Shutdown(nil)
		sv.listener.Close()
	}
}

/*
	Utilities
*/
func makeUpgrader(config Config) websocket.Upgrader {
	var checkOriginLambda func(*http.Request) bool = nil
	if !config.CheckOrigin {
		checkOriginLambda = func(*http.Request) bool { return true }
	}

	return websocket.Upgrader{
		CheckOrigin: checkOriginLambda,
	}
}

func makeAddrString(hostname string, port int) string {
	return fmt.Sprintf("%v:%v", hostname, port)
}

func (config *Config) makeAddrString() string {
	return makeAddrString(config.Hostname, config.Port)
}
