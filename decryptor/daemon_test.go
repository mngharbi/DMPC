package decryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"sync"
	"testing"
)

/*
	Test helpers
*/

func generateRandomBytes(nbBytes int) (bytes []byte) {
	bytes = make([]byte, nbBytes)
	rand.Read(bytes)
	return
}

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

func createDummyKeyRequesterFunctor(collection map[string][]byte) KeyRequester {
	return func(keyId string) []byte {
		return collection[keyId]
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

func getKeysCollection() map[string][]byte {
	return map[string][]byte{
		"KEY_1": generateRandomBytes(core.AsymmetricKeySizeBytes),
		"KEY_2": generateRandomBytes(core.AsymmetricKeySizeBytes),
	}
}

/*
	General tests
*/

func TestStartShutdownSingleWorker(t *testing.T) {
	_, executorRequester := createDummyExecutorRequesterFunctor()
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(getSignKeyCollection(), true), createDummyKeyRequesterFunctor(getKeysCollection()), executorRequester) {
		return
	}
	ShutdownServer()
}
