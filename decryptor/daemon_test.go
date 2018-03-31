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

type dummyExecutorEntry struct {
	requestNumber int
	issuerId string
	certifierId string
	payload []byte
}

type dummyExecutorRegistry struct {
	data map[int]dummyExecutorEntry
	ticketNum int
	lock *sync.Mutex
}

func createDummyExecutorRequesterFunctor() (*dummyExecutorRegistry, ExecutorRequester) {
	reg := dummyExecutorRegistry{
		ticketNum: 0,
		lock: &sync.Mutex{},
	}
	requester := func(requestNumber int, issuerId string, certifierId string, payload []byte) int {
		reg.lock.Lock()
		ticketCopy := reg.ticketNum
		reg.data[ticketCopy] = dummyExecutorEntry{
			requestNumber: requestNumber,
			issuerId: issuerId,
			certifierId: certifierId,
			payload: payload,
		}
		reg.ticketNum += 1
		reg.lock.Unlock()
		return ticketCopy
	}
	return &reg, requester
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
	_, executorRequester := createDummyExecutorRequesterFunctor()
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(getSignKeyCollection(), true), createDummyKeyRequesterFunctor(), executorRequester) {
		return
	}
	ShutdownServer()
}
