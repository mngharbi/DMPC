package core

import (
	"testing"
	"reflect"
)

func TestDecodeValid(t *testing.T) {
	valid := []byte(`{
		"version": 0.1,

		"encryption": {
			"temp": {
				"encrypted": false,
				"keys": {"mngharbi": "CIPHER"},
				"nonce": "NO_ONCE"
			},
			"perm": {
				"encrypted": true,
				"keyId": "XXX",
				"nonce": "NO_ONCE"
			}
		},

		"issue": {
			"id": "XXX",
			"signature": "SIGN_USER"
		},

		"certification": {
			"id": "XXX",
			"signature": "SIGN_CERT"
		},

		"transmission": {},

		"payload": {}
	}`)

	op, err := Decode(valid)

	if err != nil {
		t.Error("Decoding Failed")
	}

	rawOp := op.Raw

	if rawOp.Version != 0.1 {
		t.Error("Version not decoded properly")
	}

	if !(rawOp.Encryption.Temp.Encrypted == false &&
		rawOp.Encryption.Temp.Keys != nil && rawOp.Encryption.Temp.Keys["mngharbi"] == "CIPHER" &&
		rawOp.Encryption.Temp.Nonce == "NO_ONCE") {
		t.Error("Temporary encryption fields not decoded properly")
	}

	if !(rawOp.Encryption.Perm.Encrypted == true &&
		rawOp.Encryption.Perm.KeyId == "XXX" &&
		rawOp.Encryption.Perm.Nonce == "NO_ONCE") {
		t.Error("Permanent encryption fields not decoded properly")
	}

	if rawOp.Issue.Id != "XXX" || rawOp.Issue.Signature != "SIGN_USER" {
		t.Error("Issue fields not decoded properly")
	}

	if rawOp.Certification.Id != "XXX" || rawOp.Certification.Signature != "SIGN_CERT" {
		t.Error("Certification fields not decoded properly")
	}
}

func TestDecodeMalformedOperation(t *testing.T) {
	malformed := []byte(`{
		"version": 0.1,

		"encryption": {
			"temp": {
				"encrypted": "HI"
			}
		}
	}`)

	_, err := Decode(malformed)

	if err == nil {
		t.Error("Decoding should fail if type doesn't match")
	}
}

func TestDecodeMissingAttributes(t *testing.T) {
	valid := []byte(`{
		"version": 0.1
	}`)

	_, err := Decode(valid)

	if err != nil {
		t.Error("Decoding should not fail with missing parameters")
	}
}

func TestDecodeEncodeCycle(t *testing.T) {
	valid := []byte(`{
		"version": 0.1,

		"encryption": {
			"temp": {
				"encrypted": false,
				"keys": {"mngharbi": "CIPHER"},
				"nonce": "NO_ONCE"
			},
			"perm": {
				"encrypted": true,
				"keyId": "XXX",
				"nonce": "NO_ONCE"
			}
		},

		"issue": {
			"id": "XXX",
			"signature": "SIGN_USER"
		},

		"certification": {
			"id": "XXX",
			"signature": "SIGN_CERT"
		},

		"transmission": {},

		"payload": {}
	}`)

	op, _ := Decode(valid)
	encoded, _ := op.Encode()
	op2, _ := Decode(encoded)

	if !reflect.DeepEqual(op, op2) {
		t.Error(op)
		t.Error("Re-encoding should produce same value")
	}
}
