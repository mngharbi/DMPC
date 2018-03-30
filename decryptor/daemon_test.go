package decryptor

import (
	"github.com/mngharbi/DMPC/users"
	"sync"
	"testing"
)

/*
	Dummy subsystem lambdas
*/

func createDummyUsersRequesterFunctor() UsersDecodedRequester {
	return func(*users.UserRequest) (chan *users.UserResponse, []error) {
		channel := make(chan *users.UserResponse, 1)
		channel <- nil
		return channel, nil
	}
}
func createDummyKeyRequesterFunctor() KeyRequester {
	return func(string) []byte {
		return nil
	}
}
func createDummyExecutorRequesterFunctor() ExecutorRequester {
	mutex := &sync.Mutex{}
	ticketNum := 0
	return func(int, string, string, []byte) int {
		mutex.Lock()
		ticketCopy := ticketNum
		ticketNum += 1
		mutex.Unlock()
		return ticketCopy
	}
}

/*
	General tests
*/

func TestStartShutdownSingleWorker(t *testing.T) {
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersRequesterFunctor(), createDummyKeyRequesterFunctor(), createDummyExecutorRequesterFunctor()) {
		return
	}
	ShutdownServer()
}
