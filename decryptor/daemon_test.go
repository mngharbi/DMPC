package decryptor

import (
	"crypto/rsa"
	"github.com/mngharbi/DMPC/core"
	"reflect"
	"testing"
)

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
	issuerSignature, _ := core.Sign(signKeyCollection[genericIssuerId], hashedPayload[:])
	certifierSignature, _ := core.Sign(signKeyCollection[genericCertifierId], hashedPayload[:])
	permanentEncryption := core.GeneratePermanentEncryptedOperation(
		false,
		"NO_KEY",
		[]byte{},
		false,
		genericIssuerId,
		issuerSignature,
		false,
		genericCertifierId,
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
	if decryptorResp.Result != Success {
		t.Errorf("Making request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified:    true,
		requestNumber: 1,
		signers:       generateGenericSigners(),
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
	issuerSignature, _ := core.Sign(signKeyCollection[genericIssuerId], hashedPayload[:])
	certifierSignature, _ := core.Sign(signKeyCollection[genericCertifierId], hashedPayload[:])
	permanentEncryption := core.GeneratePermanentEncryptedOperation(
		false,
		"NO_KEY",
		[]byte{},
		false,
		genericIssuerId,
		issuerSignature,
		false,
		genericCertifierId,
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
	if decryptorResp.Result != Success {
		t.Errorf("Making request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified:    true,
		requestNumber: 1,
		signers:       generateGenericSigners(),
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
		keyId1,
		keyCollection[keyId1],
		generateRandomBytes(core.SymmetricNonceSize),
		1,
		payload,
		genericIssuerId,
		func(b []byte) ([]byte, bool) { return b, false },
		genericCertifierId,
		func(b []byte) ([]byte, bool) { return b, false },
	)

	signKeyCollection := map[string]*rsa.PrivateKey{
		genericIssuerId:    issuerKey,
		genericCertifierId: certifierKey,
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
	if decryptorResp.Result != Success {
		t.Errorf("Making request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified:    true,
		requestNumber: 1,
		signers:       generateGenericSigners(),
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
		keyId1,
		keyCollection[keyId1],
		payload,
		genericIssuerId,
		genericCertifierId,
		globalKey,
	)

	signKeyCollection := map[string]*rsa.PrivateKey{
		genericIssuerId:    issuerKey,
		genericCertifierId: certifierKey,
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
	if decryptorResp.Result != Success {
		t.Errorf("Making request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified:    true,
		requestNumber: 1,
		signers:       generateGenericSigners(),
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
		keyId1,
		keyCollection[keyId1],
		payload,
		genericIssuerId,
		genericCertifierId,
		globalKey,
	)

	_, errs := MakeEncodedRequest(temporaryEncryptionEncoded)
	if len(errs) == 0 {
		t.Errorf("Decryptor should not work while server is not running.")
	}

	signKeyCollection := map[string]*rsa.PrivateKey{
		genericIssuerId:    issuerKey,
		genericCertifierId: certifierKey,
	}

	_, executorRequester := createDummyExecutorRequesterFunctor()
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), createDummyKeyRequesterFunctor(keyCollection), executorRequester) {
		return
	}

	// Make request with invalid json structure
	_, errs = MakeEncodedRequest([]byte("{"))
	if len(errs) == 0 {
		t.Errorf("Decryptor request should not run if request is not encoded propoerly.")
		return
	}

	// Encrypt request with the wrong key
	differentKey := core.GeneratePrivateKey()
	temporaryEncryptionEncodedWrongKey, _, _ := generateValidEncryptedOperation(
		keyId1,
		keyCollection[keyId1],
		payload,
		genericIssuerId,
		genericCertifierId,
		differentKey,
	)
	decryptorResp, ok := makeRequestAndGetResult(t, temporaryEncryptionEncodedWrongKey)
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
		keyId1,
		keyCollection[keyId1],
		payload,
		"NON_EXISTENT",
		genericCertifierId,
		globalKey,
	)
	decryptorResp, ok = makeRequestAndGetResult(t, temporaryEncryptionEncodedNoSignKey)
	if !ok {
		return
	}
	if decryptorResp.Result != VerificationError {
		t.Errorf("Decryptor request should fail if signing key does not exist.")
		return
	}

	// Use wrong permanent encryption key
	temporaryEncryptionEncodedWrongPermanentKey, _, _ := generateValidEncryptedOperation(
		keyId2,
		keyCollection[keyId1],
		payload,
		"NON_EXISTENT",
		genericCertifierId,
		globalKey,
	)
	decryptorResp, ok = makeRequestAndGetResult(t, temporaryEncryptionEncodedWrongPermanentKey)
	if !ok {
		return
	}
	if decryptorResp.Result != VerificationError {
		t.Errorf("Decryptor request should fail if permanent encrypted payload can't be decrypted.")
		return
	}

	ShutdownServer()
}
