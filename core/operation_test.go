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

/*
	Permanent encrypted operation parsing
*/
func TestPermDecodeValid(t *testing.T) {
	valid := []byte(`{
		"encryption": {
			"encrypted": false,
			"keyId": "KEY_ID",
			"nonce": "NO_ONCE"
		},

		"issue": {
			"signature":"ISSUER_SIGNATURE"
		},

		"certification": {
			"signature":"CERTIFIER_SIGNATURE"
		},

		"meta": {
			"requestType": 1
		},

		"payload": "BASE64_CIPHER"
	}`)

	var rawOp PermanentEncryptedOperation
	err := rawOp.Decode(valid)

	if err != nil {
		t.Error("Decoding Failed")
	}

	if !(rawOp.Encryption.Encrypted == false &&
		rawOp.Encryption.KeyId == "KEY_ID" &&
		rawOp.Encryption.Nonce == "NO_ONCE") {
		t.Error("Encryption fields not decoded properly")
	}

	if !(rawOp.Issue.Signature == "ISSUER_SIGNATURE") {
		t.Error("Issuer signature not decoded properly")
	}

	if !(rawOp.Certification.Signature == "CERTIFIER_SIGNATURE") {
		t.Error("Certification signature not decoded properly")
	}

	if !(rawOp.Meta.RequestType == 1) {
		t.Error("Meta fields not decoded properly")
	}

	if rawOp.Payload != "BASE64_CIPHER" {
		t.Error("Payload fields not decoded properly")
	}
}

func TestPermDecodeMalformedRawOperation(t *testing.T) {
	malformed := []byte(`{
		"encryption": {
			"encrypted": "HI"
		}
	}`)

	var rawOp PermanentEncryptedOperation
	err := rawOp.Decode(malformed)

	if err == nil {
		t.Error("Decoding should fail if type doesn't match")
	}
}

func TestPermDecodeMissingAttributes(t *testing.T) {
	valid := []byte(`{
		"payload": "BASE64_CIPHER"
	}`)

	var rawOp PermanentEncryptedOperation
	err := rawOp.Decode(valid)

	if err != nil {
		t.Error("Decoding should not fail with missing parameters")
	}
}

func TestPermDecodeEncodeCycle(t *testing.T) {
	valid := []byte(`{
		"encryption": {
			"encrypted": false,
			"keyId": "KEY_ID",
			"nonce": "NO_ONCE"
		},

		"issue": {
			"signature":"ISSUER_SIGNATURE"
		},

		"certification": {
			"signature":"CERTIFIER_SIGNATURE"
		},

		"meta": {
			"requestType": 1
		},

		"payload": "BASE64_CIPHER"
	}`)

	var rawOp PermanentEncryptedOperation
	var rawOp2 PermanentEncryptedOperation
	rawOp.Decode(valid)
	encoded, _ := rawOp.Encode()
	rawOp2.Decode(encoded)

	if !reflect.DeepEqual(rawOp, rawOp2) {
		t.Error("Re-encoding should produce same value")
	}
}

func TestOperationDrop(t *testing.T) {
	op := &PermanentEncryptedOperation{}
	op.Meta.RequestType = UsersRequestType
	op.Meta.Buffered = false
	if !op.ShouldDrop() {
		t.Error("Anything except messages should be dropped if decryption failed")
	}

	op.Meta.RequestType = AddMessageType
	op.Meta.Buffered = false
	if op.ShouldDrop() {
		t.Error("Messages should not be dropped the first time decryption fails")
	}

	op.Meta.RequestType = AddMessageType
	op.Meta.Buffered = true
	if !op.ShouldDrop() {
		t.Error("Messages should be dropped if decryption fails after buffering")
	}
}
