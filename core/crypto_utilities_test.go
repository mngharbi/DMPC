package core

import (
	"reflect"
	"testing"
)

const invalidPemString string = "INVALID"

func TestEncodeDecodePublicKey(t *testing.T) {
	key := GeneratePublicKey()
	keyEncoded := PublicAsymKeyToString(key)
	keyDecoded, err := PublicStringToAsymKey(keyEncoded)
	if err != nil || !reflect.DeepEqual(key, keyDecoded) {
		t.Errorf("Public key encode/decode test failed.")
	}
}

func TestEncodeDecodePrivateKey(t *testing.T) {
	key := GeneratePrivateKey()
	keyEncoded := PrivateAsymKeyToString(key)
	keyDecoded, err := PrivateStringToAsymKey(keyEncoded)
	if err != nil || !reflect.DeepEqual(key, keyDecoded) {
		t.Errorf("Private key encode/decode test failed.")
	}
}
