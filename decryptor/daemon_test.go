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
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(getSignKeyCollection(), true), core.DecryptorFunctor(getKeysCollection(), true), executorRequester) {
		return
	}
	ShutdownServer()
}

func TestValidNonEncrypted(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	signKeyCollection := getSignKeyCollection()
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(getKeysCollection(), true), executorRequester) {
		return
	}

	// Create non encrypted payload
	payload := []byte("{}")
	hashedPayload := core.Hash(payload)
	issuerSignature, _ := core.Sign(signKeyCollection[genericIssuerId], hashedPayload[:])
	certifierSignature, _ := core.Sign(signKeyCollection[genericCertifierId], hashedPayload[:])
	operation := core.GenerateOperation(
		false,
		keyId1,
		[]byte{},
		false,
		genericIssuerId,
		issuerSignature,
		false,
		genericCertifierId,
		certifierSignature,
		false,
		core.UsersRequestType,
		payload,
		false,
	)
	operationEncoded, _ := operation.Encode()
	transaction := core.GenerateTransaction(
		false,
		map[string]string{},
		[]byte{},
		false,
		operationEncoded,
		false,
	)

	// Make request and get ticket number
	transactionEncoded, _ := transaction.Encode()
	decryptorResp, ok := makeTransactionRequestAndGetResult(t, transactionEncoded, true)
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
		isVerified: true,
		meta: &core.OperationMetaFields{
			RequestType: core.UsersRequestType,
			Buffered:    false,
		},
		signers: generateGenericSigners(),
		payload: payload,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()
}

func TestValidTransactionEncryptedOnly(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	signKeyCollection := getSignKeyCollection()
	globalKey := core.GeneratePrivateKey()
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(getKeysCollection(), true), executorRequester) {
		return
	}

	// Create non encrypted payload
	payload := []byte("{}")
	hashedPayload := core.Hash(payload)
	issuerSignature, _ := core.Sign(signKeyCollection[genericIssuerId], hashedPayload[:])
	certifierSignature, _ := core.Sign(signKeyCollection[genericCertifierId], hashedPayload[:])
	operation := core.GenerateOperation(
		false,
		keyId1,
		[]byte{},
		false,
		genericIssuerId,
		issuerSignature,
		false,
		genericCertifierId,
		certifierSignature,
		false,
		core.UsersRequestType,
		payload,
		false,
	)
	operationEncoded, _ := operation.Encode()
	transaction, _ := core.GenerateTransactionWithEncryption(
		operationEncoded,
		[]byte(core.CorrectChallenge),
		func(map[string]string) {},
		globalKey,
	)

	// Make request and get ticket number
	transactionEncoded, _ := transaction.Encode()
	decryptorResp, ok := makeTransactionRequestAndGetResult(t, transactionEncoded, true)
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
		isVerified: true,
		meta: &core.OperationMetaFields{
			RequestType: core.UsersRequestType,
			Buffered:    false,
		},
		signers: generateGenericSigners(),
		payload: payload,
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
	payload := []byte("{}")
	operation, issuerKey, certifierKey := core.GenerateOperationWithEncryption(
		keyId1,
		keyCollection[keyId1],
		generateRandomBytes(core.SymmetricNonceSize),
		core.UsersRequestType,
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
	if !resetAndStartServer(t, singleWorkerConfig(), nil, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(keyCollection, true), executorRequester) {
		return
	}

	operationEncoded, _ := operation.Encode()
	transaction := core.GenerateTransaction(
		false,
		map[string]string{},
		[]byte{},
		false,
		operationEncoded,
		false,
	)

	// Make request and get ticket number
	transactionEncoded, _ := transaction.Encode()
	decryptorResp, ok := makeTransactionRequestAndGetResult(t, transactionEncoded, true)
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
		isVerified: true,
		meta: &core.OperationMetaFields{
			RequestType: core.UsersRequestType,
			Buffered:    false,
		},
		signers: generateGenericSigners(),
		payload: payload,
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
	payload := []byte("{}")
	globalKey := core.GeneratePrivateKey()
	transactionEncoded, issuerKey, certifierKey := generateValidEncryptedOperation(
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
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(keyCollection, true), executorRequester) {
		return
	}

	// Make request and get ticket number
	decryptorResp, ok := makeTransactionRequestAndGetResult(t, transactionEncoded, true)
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
		isVerified: true,
		meta: &core.OperationMetaFields{
			RequestType: core.UsersRequestType,
			Buffered:    false,
		},
		signers: generateGenericSigners(),
		payload: payload,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()
}

/*
	Encrypted at both levels, no signatures, no verification
*/
func TestValidTemporaryPermanentEncryptedUnverified(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	keyCollection := getKeysCollection()

	// Create encrypted payload
	payload := []byte("{}")
	globalKey := core.GeneratePrivateKey()
	transactionEncoded, issuerKey, certifierKey := generateValidEncryptedOperation(
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
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(keyCollection, true), executorRequester) {
		return
	}

	// Make request and get ticket number
	decryptorResp, ok := makeTransactionRequestAndGetResult(t, transactionEncoded, false)
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
		isVerified: false,
		meta: &core.OperationMetaFields{
			RequestType: core.UsersRequestType,
			Buffered:    false,
		},
		signers: nil,
		payload: payload,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()
}

func TestOperationEncryption(t *testing.T) {
	reg, executorRequester := createDummyExecutorRequesterFunctor()
	keyCollection := getKeysCollection()

	// Setup operation
	payload := []byte("{}")
	globalKey := core.GeneratePrivateKey()
	operation, issuerKey, certifierKey := core.GenerateOperationWithEncryption(
		"",
		keyCollection[keyId1],
		generateRandomBytes(core.SymmetricNonceSize),
		core.AddMessageType,
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

	/*
		Not buffered add message: not correctly encrypted
	*/
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(keyCollection, true), executorRequester) {
		return
	}
	decryptorResp, ok := makeOperationRequestAndGetResult(t, operation)
	if !ok {
		return
	}
	if decryptorResp.Result != Success {
		t.Errorf("Making operation request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry := reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected := dummyExecutorEntry{
		isVerified: true,
		meta: &core.OperationMetaFields{
			RequestType: core.AddMessageType,
			Buffered:    true,
		},
		signers:         nil,
		payload:         nil,
		failedOperation: operation,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()

	/*
		Not buffered add message: not correctly signed
	*/
	operation.Encryption.KeyId = keyId1
	operation.Issue.Signature = ""
	operation.Meta.Buffered = false

	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(keyCollection, true), executorRequester) {
		return
	}
	decryptorResp, ok = makeOperationRequestAndGetResult(t, operation)
	if !ok {
		return
	}
	if decryptorResp.Result != Success {
		t.Errorf("Making operation request failed. decryptorResp=%+v", decryptorResp)
		return
	}

	// Check entry with the ticket number
	executorEntry = reg.getEntry(decryptorResp.Ticket)
	executorEntryExpected = dummyExecutorEntry{
		isVerified: true,
		meta: &core.OperationMetaFields{
			RequestType: core.AddMessageType,
			Buffered:    true,
		},
		signers:         nil,
		payload:         nil,
		failedOperation: operation,
	}
	if !reflect.DeepEqual(executorEntry, executorEntryExpected) {
		t.Errorf("Executor entry doesn't match. executorEntry=%+v, executorEntryExpected=%+v", executorEntry, executorEntryExpected)
		return
	}

	ShutdownServer()

	/*
		Buffered add message: not correctly encrypted
	*/
	operation.Encryption.KeyId = ""
	operation.Meta.Buffered = true

	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(keyCollection, true), executorRequester) {
		return
	}
	decryptorResp, ok = makeOperationRequestAndGetResult(t, operation)
	if !ok {
		return
	}
	if decryptorResp.Result != PermanentDecryptionError {
		t.Errorf("Making operation request should fail. decryptorResp=%+v", decryptorResp)
		return
	}

	ShutdownServer()
}

func TestInvalidOperationEncoding(t *testing.T) {
	// Make request while server is not running
	keyCollection := getKeysCollection()
	payload := []byte("PAYLOAD")
	globalKey := core.GeneratePrivateKey()
	transactionEncoded, issuerKey, certifierKey := generateValidEncryptedOperation(
		keyId1,
		keyCollection[keyId1],
		payload,
		genericIssuerId,
		genericCertifierId,
		globalKey,
	)

	_, errs := MakeEncodedTransactionRequest(transactionEncoded)
	if len(errs) == 0 {
		t.Errorf("Decryptor should not work while server is not running.")
	}

	signKeyCollection := map[string]*rsa.PrivateKey{
		genericIssuerId:    issuerKey,
		genericCertifierId: certifierKey,
	}

	_, executorRequester := createDummyExecutorRequesterFunctor()
	if !resetAndStartServer(t, singleWorkerConfig(), globalKey, createDummyUsersSignKeyRequesterFunctor(signKeyCollection, true), core.DecryptorFunctor(keyCollection, true), executorRequester) {
		return
	}

	// Make request with invalid json structure
	_, errs = MakeEncodedTransactionRequest([]byte("{"))
	if len(errs) == 0 {
		t.Errorf("Decryptor request should not run if request is not encoded properly.")
		return
	}

	// Encrypt request with the wrong key
	differentKey := core.GeneratePrivateKey()
	transactionEncodedWrongKey, _, _ := generateValidEncryptedOperation(
		keyId1,
		keyCollection[keyId1],
		payload,
		genericIssuerId,
		genericCertifierId,
		differentKey,
	)
	decryptorResp, ok := makeTransactionRequestAndGetResult(t, transactionEncodedWrongKey, true)
	if !ok {
		return
	}
	if decryptorResp.Result != TransactionDecryptionError {
		t.Errorf("Decryptor request should fail if transaction is not encrypted with the right key.")
		return
	}

	// Test empty decrypted permanent payload
	transactionNoPayload, _ := core.GenerateTransactionWithEncryption(
		[]byte{},
		[]byte(core.CorrectChallenge),
		func(map[string]string) {},
		globalKey,
	)
	transactionNoPayloadEncoded, _ := transactionNoPayload.Encode()
	decryptorResp, ok = makeTransactionRequestAndGetResult(t, transactionNoPayloadEncoded, true)
	if !ok {
		return
	}
	if decryptorResp.Result != TransactionDecryptionError {
		t.Errorf("Decryptor request should fail if transaction payload is empty.")
		return
	}

	// Use inexistent signing key
	transactionEncodedNoSignKey, _, _ := generateValidEncryptedOperation(
		keyId1,
		keyCollection[keyId1],
		payload,
		"NON_EXISTENT",
		genericCertifierId,
		globalKey,
	)
	decryptorResp, ok = makeTransactionRequestAndGetResult(t, transactionEncodedNoSignKey, true)
	if !ok {
		return
	}
	if decryptorResp.Result != VerificationError {
		t.Errorf("Decryptor request should fail if signing key does not exist.")
		return
	}

	// Use wrong permanent encryption key
	transactionEncodedWrongPermanentKey, _, _ := generateValidEncryptedOperation(
		keyId2,
		keyCollection[keyId1],
		payload,
		"NON_EXISTENT",
		genericCertifierId,
		globalKey,
	)
	decryptorResp, ok = makeTransactionRequestAndGetResult(t, transactionEncodedWrongPermanentKey, true)
	if !ok {
		return
	}
	if decryptorResp.Result != PermanentDecryptionError {
		t.Errorf("Decryptor request should fail if permanent encrypted payload can't be decrypted. %+v", decryptorResp)
		return
	}

	ShutdownServer()
}
