package users

import (
	"testing"
	"time"
	"reflect"
	"bytes"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"encoding/json"
)

func generatePublicKey() *rsa.PublicKey {
	Priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	return &Priv.PublicKey
}


func jsonPemEncodeKey(key *rsa.PublicKey) string {
	keyBytes, _ := x509.MarshalPKIXPublicKey(key)
	block := &pem.Block{
		Type: "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}
	buf := new(bytes.Buffer)
	_ = pem.Encode(buf, block)
	res,_ := json.Marshal(string(pem.EncodeToMemory(block)))
	return string(res)
}

func TestDecodeCreateRequest(t *testing.T) {
	encKey := generatePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := generatePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	valid := []byte(`{
		"type": 0,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": [],
		"data": {
			"id": "NEW_USER",
			"encKey": ` +encKeyStringEncoded+ `,
			"signKey": ` +signKeyStringEncoded+ `,
			"permissions": {
				"channel": {
					"add": true
				},
				"user": {
					"add": true,
					"remove": false,
					"encKeyUpdate": false,
					"signKeyUpdate": false,
					"permissionsUpdate": false
				}
			},
			"active": true
		},
		"timestamp": "2018-01-13T23:53:00Z"
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}
	expectedTimestamp, _ := time.Parse(time.RFC3339, "2018-01-13T23:53:00Z")
	expected := UserRequest{
		Type: CreateRequest,
		IssuerId: "USER",
		CertifierId: "ADMIN",
		FieldsUpdated: []string{},
		Data: UserObject{
			Id: "NEW_USER",
			EncKey: encKeyStringDecoded,
			encKeyObject: encKey,
			SignKey: signKeyStringDecoded,
			signKeyObject: signKey,
			Permissions: PermissionsObject{
				Channel: ChannelPermissionsObject{
					Add: true,
				},
				User: UserPermissionsObject{
					Add: true,
					Remove: false,
					EncKeyUpdate: false,
					SignKeyUpdate: false,
					PermissionsUpdate: false,
				},
			},
			Active: true,
		},
		Timestamp: expectedTimestamp,
	}

	if !reflect.DeepEqual(rq, expected) {
		t.Errorf("Decoding produced different results:\n result: %v\n expected: %v\n", rq, expected)
	}

}

func TestDecodeUpdateRequest(t *testing.T) {
	encKey := generatePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := generatePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": ["encKey","signKey"],
		"data": {
			"id": "NEW_USER",
			"encKey": ` +encKeyStringEncoded+ `,
			"signKey": ` +signKeyStringEncoded+ `,
			"permissions": {
				"channel": {
					"add": true
				},
				"user": {
					"add": true,
					"remove": false,
					"encKeyUpdate": false,
					"signKeyUpdate": false,
					"permissionsUpdate": false
				}
			},
			"active": true
		},
		"timestamp": "2018-01-13T23:53:00Z"
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}
	expectedTimestamp, _ := time.Parse(time.RFC3339, "2018-01-13T23:53:00Z")
	expected := UserRequest{
		Type: UpdateRequest,
		IssuerId: "USER",
		CertifierId: "ADMIN",
		FieldsUpdated: []string{"encKey","signKey"},
		Data: UserObject{
			Id: "NEW_USER",
			EncKey: encKeyStringDecoded,
			encKeyObject: encKey,
			SignKey: signKeyStringDecoded,
			signKeyObject: signKey,
			Permissions: PermissionsObject{
				Channel: ChannelPermissionsObject{
					Add: true,
				},
				User: UserPermissionsObject{
					Add: true,
					Remove: false,
					EncKeyUpdate: false,
					SignKeyUpdate: false,
					PermissionsUpdate: false,
				},
			},
			Active: true,
		},
		Timestamp: expectedTimestamp,
	}

	if !reflect.DeepEqual(rq, expected) {
		t.Errorf("Decoding produced different results:\n result: %v\n expected: %v\n", rq, expected)
	}

}

func TestDecodeDeleteRequest(t *testing.T) {
	encKey := generatePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := generatePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	valid := []byte(`{
		"type": 2,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": [],
		"data": {
			"id": "NEW_USER"
		},
		"timestamp": "2018-01-13T23:53:00Z"
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}
	expectedTimestamp, _ := time.Parse(time.RFC3339, "2018-01-13T23:53:00Z")
	expected := UserRequest{
		Type: DeleteRequest,
		IssuerId: "USER",
		CertifierId: "ADMIN",
		FieldsUpdated: []string{},
		Data: UserObject{
			Id: "NEW_USER",
		},
		Timestamp: expectedTimestamp,
	}

	if !reflect.DeepEqual(rq, expected) {
		t.Errorf("Decoding produced different results:\n result: %v\n expected: %v\n", rq, expected)
	}
}

func TestDecodeOneFieldUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": ["active"],
		"data": {
			"active": false
		}
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding with one modified value should be correct, errors: %v", errs)
	}
}

func TestDecodeEmptyUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": []
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if !(len(errs) == 1 && errs[0].Error() == "No fields updated") {
		t.Errorf("Decoding update with no fields updated should fail with one error, errors: %v", errs)
	}
}

func TestDecodeInvalidFieldsUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": ["active","randomParam"]
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}

	expected := UserRequest{
		Type: UpdateRequest,
		IssuerId: "USER",
		CertifierId: "ADMIN",
		FieldsUpdated: []string{"active"},
	}

	if !reflect.DeepEqual(rq, expected) {
		t.Errorf("Decoding produced different results:\n result: %v\n expected: %v\n", rq, expected)
	}
}

func TestDecodeMissingIssuerUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"certifierId": "ADMIN",
		"fieldsUpdated": ["active"]
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if !(len(errs) == 1 && errs[0].Error() == "Issuer id missing") {
		t.Errorf("Decoding update with missing issuer should fail with one error, errors: %v", errs)
	}
}

func TestDecodeMissingCertifierUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"fieldsUpdated": ["active"]
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if !(len(errs) == 1 && errs[0].Error() == "Certifier id missing") {
		t.Errorf("Decoding update with missing certifier should fail with one error, errors: %v", errs)
	}
}
