package core

import (
	"reflect"
	"testing"
	"time"
)

/*
	Operation parsing
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
			"requestType": 1,
			"timestamp": "2018-01-01T00:00:00.000Z"
		},

		"payload": "BASE64_CIPHER"
	}`)

	var rawOp Operation
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
		t.Error("Meta request type not decoded properly")
	}

	if !(rawOp.Meta.Timestamp.Unix() == time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC).Unix()) {
		t.Error("Meta timestamp not decoded properly")
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

	var rawOp Operation
	err := rawOp.Decode(malformed)

	if err == nil {
		t.Error("Decoding should fail if type doesn't match")
	}
}

func TestPermDecodeMissingAttributes(t *testing.T) {
	valid := []byte(`{
		"payload": "BASE64_CIPHER"
	}`)

	var rawOp Operation
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
			"requestType": 1,
			"timestamp": "2018-01-01T00:00:00.000Z"
		},

		"payload": "BASE64_CIPHER"
	}`)

	var rawOp Operation
	var rawOp2 Operation
	rawOp.Decode(valid)
	encoded, _ := rawOp.Encode()
	rawOp2.Decode(encoded)

	if !reflect.DeepEqual(rawOp, rawOp2) {
		t.Error("Re-encoding should produce same value")
	}
}

func TestOperationDrop(t *testing.T) {
	op := &Operation{}
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
