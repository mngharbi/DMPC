package core

import (
	"reflect"
	"testing"
)

/*
	Transaction parsing
*/
func TestTransactionDecodeValid(t *testing.T) {
	valid := []byte(`{
		"version": 0.1,

		"encryption": {
			"encrypted": false,
			"challenges": {"CIPHER": "CHALLENGE_CIPHER"},
			"nonce": "NO_ONCE"
		},

		"transmission": {},

		"payload": "BASE64_CIPHER"
	}`)

	var rawOp Transaction
	err := rawOp.Decode(valid)

	if err != nil {
		t.Error("Decoding Failed")
	}

	if rawOp.Version != 0.1 {
		t.Error("Version not decoded properly")
	}

	if !(rawOp.Encryption.Encrypted == false &&
		rawOp.Encryption.Challenges != nil &&
		rawOp.Encryption.Challenges["CIPHER"] == "CHALLENGE_CIPHER" &&
		rawOp.Encryption.Nonce == "NO_ONCE") {
		t.Error("Transaction encryption fields not decoded properly")
	}

	if rawOp.Payload != "BASE64_CIPHER" {
		t.Error("Payload fields not decoded properly")
	}
}

func TestTransactionDecodeMalformedRawOperation(t *testing.T) {
	malformed := []byte(`{
		"version": 0.1,

		"encryption": {
			"encrypted": "HI"
		}
	}`)

	var rawOp Transaction
	err := rawOp.Decode(malformed)

	if err == nil {
		t.Error("Decoding should fail if type doesn't match")
	}
}

func TestTransactionDecodeMissingAttributes(t *testing.T) {
	valid := []byte(`{
		"version": 0.1
	}`)

	var rawOp Transaction
	err := rawOp.Decode(valid)

	if err != nil {
		t.Error("Decoding should not fail with missing parameters")
	}
}

func TestTransactionDecodeEncodeCycle(t *testing.T) {
	valid := []byte(`{
		"version": 0.1,

		"encryption": {
			"encrypted": false,
			"challenges": {"CIPHER": "CHALLENGE_CIPHER"},
			"nonce": "NO_ONCE"
		},

		"transmission": {},

		"payload": {}
	}`)

	var rawOp Transaction
	var rawOp2 Transaction
	rawOp.Decode(valid)
	encoded, _ := rawOp.Encode()
	rawOp2.Decode(encoded)

	if !reflect.DeepEqual(rawOp, rawOp2) {
		t.Error("Re-encoding should produce same value")
	}
}
