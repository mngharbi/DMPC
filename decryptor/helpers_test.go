/*
	Test helpers
*/

package decryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/status"
	"sync"
	"testing"
)

/*
	Server
*/

func resetAndStartServer(
	t *testing.T,
	conf Config,
	globalKey *rsa.PrivateKey,
	usersSignKeyRequester core.UsersSignKeyRequester,
	keyDecryptor core.Decryptor,
	executorRequester executor.Requester,
) bool {
	serverSingleton = server{}
	InitializeServer(globalKey, usersSignKeyRequester, keyDecryptor, executorRequester, log, shutdownProgram)
	err := StartServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func makeTransactionRequestAndGetResult(
	t *testing.T,
	requestBytes []byte,
) (*DecryptorResponse, bool) {
	channel, errs := MakeEncodedTransactionRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Decryptor should pass along request.")
		return nil, false
	}
	nativeRespPtr := <-channel
	return (*nativeRespPtr).(*DecryptorResponse), true
}

func makeOperationRequestAndGetResult(
	t *testing.T,
	operation *core.Operation,
) (*DecryptorResponse, bool) {
	channel, errs := MakeOperationRequest(operation)
	if len(errs) != 0 {
		t.Errorf("Decryptor should pass along request.")
		return nil, false
	}
	nativeRespPtr := <-channel
	return (*nativeRespPtr).(*DecryptorResponse), true
}

func multipleWorkersConfig() Config {
	return Config{
		NumWorkers: 6,
	}
}

func singleWorkerConfig() Config {
	return Config{
		NumWorkers: 1,
	}
}

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
	operation, issuerKey, certifierKey := core.GenerateOperationWithEncryption(
		keyId,
		key,
		generateRandomBytes(core.SymmetricNonceSize),
		core.UsersRequestType,
		payload,
		issuerId,
		func(b []byte) ([]byte, bool) { return b, false },
		certifierId,
		func(b []byte) ([]byte, bool) { return b, false },
	)

	operationEncoded, _ := operation.Encode()
	transaction, _ := core.GenerateTransactionWithEncryption(
		operationEncoded,
		[]byte(core.CorrectChallenge),
		func(map[string]string) {},
		globalKey,
	)

	transactionEncoded, _ := transaction.Encode()

	return transactionEncoded, issuerKey, certifierKey
}

/*
	Dummy subsystem lambdas
*/

func createDummyUsersSignKeyRequesterFunctor(collection map[string]*rsa.PrivateKey, success bool) core.UsersSignKeyRequester {
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

func createDummyKeyRequesterFunctor(collection map[string][]byte) core.KeyRequester {
	return func(keyId string) []byte {
		return collection[keyId]
	}
}

type dummyExecutorEntry struct {
	isVerified      bool
	requestType     core.RequestType
	signers         *core.VerifiedSigners
	payload         []byte
	failedOperation *core.Operation
}

type dummyExecutorRegistry struct {
	data map[status.Ticket]dummyExecutorEntry
	lock *sync.Mutex
}

func (reg *dummyExecutorRegistry) getEntry(id status.Ticket) dummyExecutorEntry {
	reg.lock.Lock()
	entryCopy := reg.data[id]
	reg.lock.Unlock()
	return entryCopy
}

func createDummyExecutorRequesterFunctor() (*dummyExecutorRegistry, executor.Requester) {
	reg := dummyExecutorRegistry{
		data: map[status.Ticket]dummyExecutorEntry{},
		lock: &sync.Mutex{},
	}
	requester := func(isVerified bool, requestType core.RequestType, signers *core.VerifiedSigners, payload []byte, failedOperation *core.Operation) (status.Ticket, error) {
		reg.lock.Lock()
		ticketCopy := status.RequestNewTicket()
		reg.data[ticketCopy] = dummyExecutorEntry{
			isVerified:      isVerified,
			requestType:     requestType,
			signers:         signers,
			payload:         payload,
			failedOperation: failedOperation,
		}
		reg.lock.Unlock()
		return ticketCopy, nil
	}
	return &reg, requester
}

/*
	Collections
*/

const (
	genericIssuerId    string = "ISSUER_ID"
	genericCertifierId string = "CERTIFIER_ID"
	keyId1             string = "KEY_1"
	keyId2             string = "KEY_2"
)

func generateSigners(issuerId string, certifierId string) *core.VerifiedSigners {
	return &core.VerifiedSigners{
		IssuerId:    issuerId,
		CertifierId: certifierId,
	}
}

func generateGenericSigners() *core.VerifiedSigners {
	return generateSigners(genericIssuerId, genericCertifierId)
}

func getSignKeyCollection() map[string]*rsa.PrivateKey {
	return map[string]*rsa.PrivateKey{
		genericIssuerId:    core.GeneratePrivateKey(),
		genericCertifierId: core.GeneratePrivateKey(),
	}
}

func getKeysCollection() map[string][]byte {
	return map[string][]byte{
		keyId1: generateRandomBytes(core.SymmetricKeySize),
		keyId2: generateRandomBytes(core.SymmetricKeySize),
	}
}
