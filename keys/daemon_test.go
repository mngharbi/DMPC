package keys

import (
	"reflect"
	"testing"
)

/*
	General tests
*/
func TestStartShutdownSingleWorker(t *testing.T) {
	if !resetAndStartServer(t) {
		return
	}
	ShutdownServer()
}

/*
	Adding keys
*/
func TestAddKeyServerDown(t *testing.T) {
	keys := getKeysCollection()
	if AddKey(keyId1, keys[keyId1]) == nil {
		t.Error("Adding key while server is down should fail")
	}
}

func TestAddKeyInvalidId(t *testing.T) {
	resetServer()
	if makeAddKeyRequest(t, invalidKeyId, getKeysCollection()[invalidKeyId]) != invalidRequestFormatError {
		t.Error("Adding key with invalid id should fail")
	}
}

func TestAddKeyInvalidKey(t *testing.T) {
	resetServer()
	if makeAddKeyRequest(t, invalidKey, getKeysCollection()[invalidKey]) != invalidRequestFormatError {
		t.Error("Adding invalid key should fail")
	}
}

func TestAddKeyDuplicate(t *testing.T) {
	resetServer()
	keys := getKeysCollection()
	if makeAddKeyRequest(t, keyId1, keys[keyId1]) != nil {
		t.Error("Adding valid key should not fail")
	}
	if makeAddKeyRequest(t, keyId1, keys[keyId2]) != nil {
		t.Error("Adding duplicate keys should not fail")
	}
	finalKey := getKeyRecordById(keyId1).Key
	expectedKey := keys[keyId1]
	if !reflect.DeepEqual(finalKey, expectedKey) {
		t.Errorf("Final key: %+v.\n Should be: %+v", finalKey, expectedKey)
	}
}

/*
	Decryption
*/
func TestDecryptServerDown(t *testing.T) {
	key := getKeysCollection()[keyId1]
	_, _, cipher := getPlainNonceCipher(key)
	if _, err := Decrypt(keyId1, validNonce(), cipher); err == nil {
		t.Error("Decrypting while server is down should fail")
	}
}

func TestDecryptInvalidKeyId(t *testing.T) {
	if !resetAndStartServer(t) {
		return
	}

	key := getKeysCollection()[keyId1]
	_, _, cipher := getPlainNonceCipher(key)
	if _, err := Decrypt(invalidKeyId, validNonce(), cipher); err != invalidRequestFormatError {
		t.Error("Decrypting with invalid key id should fail")
	}

	ShutdownServer()
}

func TestDecryptInvalidNonce(t *testing.T) {
	if !resetAndStartServer(t) {
		return
	}

	key := getKeysCollection()[keyId1]
	_, _, cipher := getPlainNonceCipher(key)
	if _, err := Decrypt(keyId1, invalidNonce(), cipher); err != invalidRequestFormatError {
		t.Error("Decrypting with invalid nonce should fail")
	}

	ShutdownServer()
}

func TestDecryptInexistentKey(t *testing.T) {
	if !resetAndStartServer(t) {
		return
	}

	key := getKeysCollection()[keyId1]
	_, _, cipher := getPlainNonceCipher(key)
	if _, err := Decrypt(keyId1, validNonce(), cipher); err != decryptionFailedError {
		t.Error("Decrypting with inexistent key id should fail")
	}

	ShutdownServer()
}

func TestValidDecrypt(t *testing.T) {
	if !resetAndStartServer(t) {
		return
	}

	key := getKeysCollection()[keyId1]
	if AddKey(keyId1, key) != nil {
		t.Error("Adding valid key should not fail")
	}

	expectedPlain, nonce, cipher := getPlainNonceCipher(key)
	plain, err := Decrypt(keyId1, nonce, cipher)
	if err != nil || !reflect.DeepEqual(plain, expectedPlain) {
		t.Error("Decrypting with existent key id should not fail")
	}

	ShutdownServer()
}
