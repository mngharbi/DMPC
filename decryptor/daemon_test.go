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

func generateValidEncryptedOperation(
	keyId string,
	key []byte,
	payload []byte,
	issuerId string,
	certifierId string,
	globalKey *rsa.PrivateKey,
) ([]byte, *rsa.PrivateKey, *rsa.PrivateKey) {
	permanentEncryption, issuerKey, certifierKey := core.GeneratePermanentEncryptedOperationWithEncryption(
		keyId,
		key,
		generateRandomBytes(core.SymmetricNonceSize),
		1,
		payload,
		issuerId,
		func(b []byte) ([]byte, bool) { return b, false },
		certifierId,
		func(b []byte) ([]byte, bool) { return b, false },
	)

	permanentEncryptionEncoded, _ := permanentEncryption.Encode()
	temporaryEncryption, _ := core.GenerateTemporaryEncryptedOperationWithEncryption(
		permanentEncryptionEncoded,
		[]byte(core.CorrectChallenge),
		func(map[string]string) {},
		globalKey,
	)

	temporaryEncryptionEncoded, _ := temporaryEncryption.Encode()

	return temporaryEncryptionEncoded, issuerKey, certifierKey
}

/*
	Dummy subsystem lambdas
*/

func createDummyUsersSignKeyRequesterFunctor(collection map[string]*rsa.PrivateKey, success bool) UsersSignKeyRequester {
	notFoundError := errors.New("Could not find signing key.")
	return func(keysIds []string) ([]*rsa.PublicKey, error) {
		res := []*rsa.PublicKey{}
		for _, keyId := range keysIds {
			privateKey, ok := collection[keyId]
			if !ok {
				return nil, notFoundError
			}
			res = append(res, &(privateKey.PublicKey))
		}
		if !success {
			return nil, notFoundError
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
	decryptorResp, ok := makeRequestAndGetResult(t, temporaryEncryptionEncoded)
	if !ok {
		return
	}
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
		func(map[string]string) {},
		globalKey,
	)

	// Make request and get ticket number
	temporaryEncryptionEncoded, _ := temporaryEncryption.Encode()
	decryptorResp, ok := makeRequestAndGetResult(t, temporaryEncryptionEncoded)
	if !ok {
		return
	}
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
	decryptorResp, ok := makeRequestAndGetResult(t, temporaryEncryptionEncoded)
	if !ok {
		return
	}
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

func TestValidTemporaryPermanentEncrypted(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	keyCollection := getKeysCollection()

	// Create encrypted payload
	payload := []byte("PAYLOAD")
	globalKey := core.GeneratePrivateKey()
	temporaryEncryptionEncoded, issuerKey, certifierKey := generateValidEncryptedOperation(
		"KEY_1",
		keyCollection["KEY_1"],
		payload,
		"ISSUER_KEY",
		"CERTIFIER_KEY",
		globalKey,
	)

	signKeyCollection := map[string]*rsa.PrivateKey{
		"ISSUER_KEY":    issuerKey,
		"CERTIFIER_KEY": certifierKey,
	}

	// Start server
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), createDummyKeyRequesterFunctor(keyCollection), executorRequester) {
		return
	}

	// Make request and get ticket number
	decryptorResp, ok := makeRequestAndGetResult(t, temporaryEncryptionEncoded)
	if !ok {
		return
	}
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

func TestInvalidOperationEncoding(t *testing.T) {
	// Make request while server is not running
	keyCollection := getKeysCollection()
	payload := []byte("PAYLOAD")
	globalKey := core.GeneratePrivateKey()
	temporaryEncryptionEncoded, issuerKey, certifierKey := generateValidEncryptedOperation(
		"KEY_1",
		keyCollection["KEY_1"],
		payload,
		"ISSUER_KEY",
		"CERTIFIER_KEY",
		globalKey,
	)

	_, errs := MakeRequest(temporaryEncryptionEncoded)

	if len(errs) == 0 {
		t.Errorf("Decryptor should not work while server is not running.")
	}

	signKeyCollection := map[string]*rsa.PrivateKey{
		"ISSUER_KEY":    issuerKey,
		"CERTIFIER_KEY": certifierKey,
	}

	_, executorRequester := createDummyExecutorRequesterFunctor()
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), createDummyKeyRequesterFunctor(keyCollection), executorRequester) {
		return
	}

	// Make empty request
	decryptorResp, ok := makeRequestAndGetResult(t, []byte{})
	if !ok {
		return
	}
	if decryptorResp.Result != TemporaryDecryptionError {
		t.Errorf("Decryptor request should fail if empty")
		return
	}

	// Make request with invalid json structure
	decryptorResp, ok = makeRequestAndGetResult(t, []byte("{"))
	if !ok {
		return
	}
	if decryptorResp.Result != TemporaryDecryptionError {
		t.Errorf("Decryptor request should fail if request is not encoded propoerly.")
		return
	}

	// Encrypt request with the wrong key
	differentKey := core.GeneratePrivateKey()
	temporaryEncryptionEncodedWrongKey, _, _ := generateValidEncryptedOperation(
		"KEY_1",
		keyCollection["KEY_1"],
		payload,
		"ISSUER_KEY",
		"CERTIFIER_KEY",
		differentKey,
	)
	decryptorResp, ok = makeRequestAndGetResult(t, temporaryEncryptionEncodedWrongKey)
	if !ok {
		return
	}
	if decryptorResp.Result != TemporaryDecryptionError {
		t.Errorf("Decryptor request should fail if request is not temporarily encrypted with the right key.")
		return
	}

	// Test empty decrypted permanent payload
	temporaryEncryptionNoPayload, _ := core.GenerateTemporaryEncryptedOperationWithEncryption(
		[]byte{},
		[]byte(core.CorrectChallenge),
		func(map[string]string) {},
		globalKey,
	)
	temporaryEncryptionNoPayloadEncoded, _ := temporaryEncryptionNoPayload.Encode()
	decryptorResp, ok = makeRequestAndGetResult(t, temporaryEncryptionNoPayloadEncoded)
	if !ok {
		return
	}
	if decryptorResp.Result != TemporaryDecryptionError {
		t.Errorf("Decryptor request should fail if temporary encrypted payload is empty.")
		return
	}

	// Use inexistent signing key
	temporaryEncryptionEncodedNoSignKey, _, _ := generateValidEncryptedOperation(
		"KEY_1",
		keyCollection["KEY_1"],
		payload,
		"NOT_EXISTENT",
		"CERTIFIER_KEY",
		globalKey,
	)
	decryptorResp, ok = makeRequestAndGetResult(t, temporaryEncryptionEncodedNoSignKey)
	if !ok {
		return
	}
	if decryptorResp.Result != PermanentDecryptionError {
		t.Errorf("Decryptor request should fail if singing key does not exist.")
		return
	}

	// Use wrong permanent encryption key
	temporaryEncryptionEncodedWrongPermanentKey, _, _ := generateValidEncryptedOperation(
		"KEY_2",
		keyCollection["KEY_1"],
		payload,
		"NOT_EXISTENT",
		"CERTIFIER_KEY",
		globalKey,
	)
	decryptorResp, ok = makeRequestAndGetResult(t, temporaryEncryptionEncodedWrongPermanentKey)
	if !ok {
		return
	}
	if decryptorResp.Result != PermanentDecryptionError {
		t.Errorf("Decryptor request should fail if permanent encrypted payload can't be decrypted.")
		return
	}

	ShutdownServer()
}
