package decryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"reflect"
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

func createDummyUsersSignKeyRequesterFunctor(collection map[string]*rsa.PrivateKey, success bool) UsersSignKeyRequester {
	return func(keysIds []string) ([]*rsa.PublicKey, error) {
		res := []*rsa.PublicKey{}
		for _, keyId := range keysIds {
			res = append(res, &(collection[keyId].PublicKey))
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
	isVerified    bool
	requestNumber int
	issuerId      string
	certifierId   string
	payload       []byte
}

type dummyExecutorRegistry struct {
	data      map[int]dummyExecutorEntry
	ticketNum int
	lock      *sync.Mutex
}

func (reg *dummyExecutorRegistry) getEntry(id int) dummyExecutorEntry {
	reg.lock.Lock()
	entryCopy := reg.data[id]
	reg.lock.Unlock()
	return entryCopy
}

func createDummyExecutorRequesterFunctor() (*dummyExecutorRegistry, ExecutorRequester) {
	reg := dummyExecutorRegistry{
		data:      map[int]dummyExecutorEntry{},
		ticketNum: 0,
		lock:      &sync.Mutex{},
	}
	requester := func(isVerified bool, requestNumber int, issuerId string, certifierId string, payload []byte) int {
		reg.lock.Lock()
		ticketCopy := reg.ticketNum
		reg.data[ticketCopy] = dummyExecutorEntry{
			isVerified:    isVerified,
			requestNumber: requestNumber,
			issuerId:      issuerId,
			certifierId:   certifierId,
			payload:       payload,
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

func getSignKeyCollection() map[string]*rsa.PrivateKey {
	return map[string]*rsa.PrivateKey{
		"ISSUER_KEY":    core.GeneratePrivateKey(),
		"CERTIFIER_KEY": core.GeneratePrivateKey(),
	}
}

func getKeysCollection() map[string][]byte {
	return map[string][]byte{
		"KEY_1": generateRandomBytes(core.SymmetricKeySize),
		"KEY_2": generateRandomBytes(core.SymmetricKeySize),
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

func TestValidNonEncrypted(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	signKeyCollection := getSignKeyCollection()
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), createDummyKeyRequesterFunctor(getKeysCollection()), executorRequester) {
		return
	}

	// Create non encrypted payload
	payload := []byte("PAYLOAD")
	hashedPayload := core.Hash(payload)
	issuerSignature, _ := core.Sign(signKeyCollection["ISSUER_KEY"], hashedPayload[:])
	certifierSignature, _ := core.Sign(signKeyCollection["CERTIFIER_KEY"], hashedPayload[:])
	permanentEncryption := core.GeneratePermanentEncryptedOperation(
		false,
		"NO_KEY",
		[]byte{},
		false,
		"ISSUER_KEY",
		issuerSignature,
		false,
		"CERTIFIER_KEY",
		certifierSignature,
		false,
		1,
		payload,
		false,
	)
	permanentEncryptionEncoded, _ := permanentEncryption.Encode()
	temporaryEncryption := core.GenerateTemporaryEncryptedOperation(
		false,
		map[string]string{},
		[]byte{},
		false,
		permanentEncryptionEncoded,
		false,
	)

	// Make request and get ticket number
	temporaryEncryptionEncoded, _ := temporaryEncryption.Encode()
	channel, errs := MakeRequest(temporaryEncryptionEncoded)
	if errs != nil {
		t.Errorf("Valid non encrypted request should not fail. errs=%v", errs)
		return
	}
	nativeRespPtr := <-channel
	decryptorResp := (*nativeRespPtr).(*DecryptorResponse)
	if decryptorResp.Result != Success ||
		decryptorResp.Ticket != 0 {
		t.Errorf("Making request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified:    true,
		requestNumber: 1,
		issuerId:      "ISSUER_KEY",
		certifierId:   "CERTIFIER_KEY",
		payload:       payload,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()
}

func TestValidTemporaryEncryptedOnly(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	signKeyCollection := getSignKeyCollection()
	globalKey := core.GeneratePrivateKey()
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), createDummyKeyRequesterFunctor(getKeysCollection()), executorRequester) {
		return
	}

	// Create non encrypted payload
	payload := []byte("PAYLOAD")
	hashedPayload := core.Hash(payload)
	issuerSignature, _ := core.Sign(signKeyCollection["ISSUER_KEY"], hashedPayload[:])
	certifierSignature, _ := core.Sign(signKeyCollection["CERTIFIER_KEY"], hashedPayload[:])
	permanentEncryption := core.GeneratePermanentEncryptedOperation(
		false,
		"NO_KEY",
		[]byte{},
		false,
		"ISSUER_KEY",
		issuerSignature,
		false,
		"CERTIFIER_KEY",
		certifierSignature,
		false,
		1,
		payload,
		false,
	)
	permanentEncryptionEncoded, _ := permanentEncryption.Encode()
	temporaryEncryption, _ := core.GenerateTemporaryEncryptedOperationWithEncryption(
		permanentEncryptionEncoded,
		[]byte(core.CorrectChallenge),
		func(map[string]string) { },
		globalKey,
	)

	// Make request and get ticket number
	temporaryEncryptionEncoded, _ := temporaryEncryption.Encode()
	channel, errs := MakeRequest(temporaryEncryptionEncoded)
	if errs != nil {
		t.Errorf("Valid non encrypted request should not fail. errs=%v", errs)
		return
	}
	nativeRespPtr := <-channel
	decryptorResp := (*nativeRespPtr).(*DecryptorResponse)
	if decryptorResp.Result != Success ||
		decryptorResp.Ticket != 0 {
		t.Errorf("Making request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified:    true,
		requestNumber: 1,
		issuerId:      "ISSUER_KEY",
		certifierId:   "CERTIFIER_KEY",
		payload:       payload,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()
}

func TestValidPermanentEncryptedOnly(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	keyCollection := getKeysCollection()

	// Create non encrypted payload
	payload := []byte("PAYLOAD")
	permanentEncryption, issuerKey, certifierKey := core.GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_1",
		keyCollection["KEY_1"],
		generateRandomBytes(core.SymmetricNonceSize),
		1,
		payload,
		"ISSUER_KEY",
		func(b []byte) ([]byte, bool) { return b, false },
		"CERTIFIER_KEY",
		func(b []byte) ([]byte, bool) { return b, false },
	)

	signKeyCollection := map[string]*rsa.PrivateKey{
		"ISSUER_KEY":    issuerKey,
		"CERTIFIER_KEY": certifierKey,
	}

	// Start server
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), createDummyKeyRequesterFunctor(keyCollection), executorRequester) {
		return
	}

	permanentEncryptionEncoded, _ := permanentEncryption.Encode()
	temporaryEncryption := core.GenerateTemporaryEncryptedOperation(
		false,
		map[string]string{},
		[]byte{},
		false,
		permanentEncryptionEncoded,
		false,
	)

	// Make request and get ticket number
	temporaryEncryptionEncoded, _ := temporaryEncryption.Encode()
	channel, errs := MakeRequest(temporaryEncryptionEncoded)
	if errs != nil {
		t.Errorf("Valid non encrypted request should not fail. errs=%v", errs)
		return
	}
	nativeRespPtr := <-channel
	decryptorResp := (*nativeRespPtr).(*DecryptorResponse)
	if decryptorResp.Result != Success ||
		decryptorResp.Ticket != 0 {
		t.Errorf("Making request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified:    true,
		requestNumber: 1,
		issuerId:      "ISSUER_KEY",
		certifierId:   "CERTIFIER_KEY",
		payload:       payload,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()
}
