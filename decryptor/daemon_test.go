package decryptor

import (
	"crypto/rsa"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"sync"
	"testing"
)

/*
	Dummy subsystem lambdas
*/

func createDummyUsersSignKeyRequesterFunctor(collection map[string]*rsa.PublicKey, success bool) UsersSignKeyRequester {
	return func(keysIds []string) ([]*rsa.PublicKey, error) {
		res := []*rsa.PublicKey{}
		for _, keyId := range keysIds {
			res = append(res, collection[keyId])
		}
		if !success {
			return nil, errors.New("Could not find signing key.")
		}
		return res, nil
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
	Collections
*/

func getSignKeyCollection() map[string]*rsa.PublicKey {
	return map[string]*rsa.PublicKey{
		"ISSUER_KEY":    core.GeneratePublicKey(),
		"CERTIFIER_KEY": core.GeneratePublicKey(),
	}
}

/*
	General tests
*/

func TestStartShutdownSingleWorker(t *testing.T) {
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(getSignKeyCollection(), true), createDummyKeyRequesterFunctor(), createDummyExecutorRequesterFunctor()) {
		return
	}
	ShutdownServer()
}
