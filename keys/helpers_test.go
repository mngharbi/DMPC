/*
	Test helpers
*/

package keys

import (
	"crypto/rand"
	"github.com/mngharbi/DMPC/core"
	"testing"
)

/*
	Server
*/

func resetAndStartServer(t *testing.T) bool {
	resetServer()
	return startServer(t)
}

func resetServer() {
	serverSingleton = server{}
}

func startServer(t *testing.T) bool {
	err := StartServer(multipleWorkersConfig(), log, shutdownProgram)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func multipleWorkersConfig() Config {
	return Config{
		NumWorkers: 6,
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

func makeAddKeyRequest(t *testing.T, keyId string, key []byte) error {
	if !startServer(t) {
		return nil
	}
	err := AddKey(keyId, key)
	ShutdownServer()
	return err
}

func getKeyRecordById(id string) *keyRecord {
	item := serverSingleton.store.Get(&keyRecord{Id: id}, recordIdIndex)
	if item != nil {
		return item.(*keyRecord)
	} else {
		return nil
	}
}

func encrypt(key []byte, payload []byte, nonce []byte) []byte {
	aead, _ := core.NewAead(key)
	encrypted := core.SymmetricEncrypt(
		aead,
		payload[:0],
		nonce,
		payload,
	)
	return encrypted
}

/*
	Collections
*/

const (
	keyId1       string = "KEY_1"
	keyId2       string = "KEY_2"
	invalidKey   string = "INVALID"
	invalidKeyId string = ""
)

func getKeysCollection() map[string][]byte {
	return map[string][]byte{
		keyId1:       generateRandomBytes(core.SymmetricKeySize),
		keyId2:       generateRandomBytes(core.SymmetricKeySize),
		invalidKey:   generateRandomBytes(1 + core.SymmetricKeySize),
		invalidKeyId: generateRandomBytes(core.SymmetricKeySize),
	}
}

func getPlainNonceCipher(key []byte) ([]byte, []byte, []byte) {
	plain := generateRandomBytes(5 * core.SymmetricKeySize)
	nonce := validNonce()
	return plain, nonce, encrypt(key, plain, nonce)
}

func validNonce() []byte {
	return generateRandomBytes(core.SymmetricNonceSize)
}

func invalidNonce() []byte {
	return generateRandomBytes(1 + core.SymmetricNonceSize)
}
