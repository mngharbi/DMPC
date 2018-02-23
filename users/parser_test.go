package users

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"reflect"
	"testing"
	"time"
)

/*
	Helpers
*/

func generatePrivateKey() *rsa.PrivateKey {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	return priv
}

func generatePublicKey() *rsa.PublicKey {
	priv := generatePrivateKey()
	return &priv.PublicKey
}

func pemEncodeKey(key *rsa.PublicKey) string {
	keyBytes, _ := x509.MarshalPKIXPublicKey(key)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}
	buf := new(bytes.Buffer)
	pem.Encode(buf, block)
	return string(pem.EncodeToMemory(block))
}

func jsonPemEncodeKey(key *rsa.PublicKey) string {
	res, _ := json.Marshal(pemEncodeKey(key))
	return string(res)
}

/*
	Decoding
*/
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
		"fields": [],
		"data": {
			"id": "NEW_USER",
			"encKey": ` + encKeyStringEncoded + `,
			"signKey": ` + signKeyStringEncoded + `,
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
	rq.skipPermissions = false
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}
	expectedTimestamp, _ := time.Parse(time.RFC3339, "2018-01-13T23:53:00Z")
	expected := UserRequest{
		Type:        CreateRequest,
		IssuerId:    "USER",
		CertifierId: "ADMIN",
		Fields:      []string{},
		Data: UserObject{
			Id:            "NEW_USER",
			EncKey:        encKeyStringDecoded,
			encKeyObject:  encKey,
			SignKey:       signKeyStringDecoded,
			signKeyObject: signKey,
			Permissions: PermissionsObject{
				Channel: ChannelPermissionsObject{
					Add: true,
				},
				User: UserPermissionsObject{
					Add:               true,
					Remove:            false,
					EncKeyUpdate:      false,
					SignKeyUpdate:     false,
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
		"fields": ["encKey","signKey"],
		"data": {
			"id": "NEW_USER",
			"encKey": ` + encKeyStringEncoded + `,
			"signKey": ` + signKeyStringEncoded + `,
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
	rq.skipPermissions = false
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}
	expectedTimestamp, _ := time.Parse(time.RFC3339, "2018-01-13T23:53:00Z")
	expected := UserRequest{
		Type:        UpdateRequest,
		IssuerId:    "USER",
		CertifierId: "ADMIN",
		Fields:      []string{"encKey", "signKey"},
		Data: UserObject{
			Id:            "NEW_USER",
			EncKey:        encKeyStringDecoded,
			encKeyObject:  encKey,
			SignKey:       signKeyStringDecoded,
			signKeyObject: signKey,
			Permissions: PermissionsObject{
				Channel: ChannelPermissionsObject{
					Add: true,
				},
				User: UserPermissionsObject{
					Add:               true,
					Remove:            false,
					EncKeyUpdate:      false,
					SignKeyUpdate:     false,
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

func TestDecodeOneFieldUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fields": ["active"],
		"data": {
			"active": false
		}
	}`)

	var rq UserRequest
	rq.skipPermissions = false
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
		"fields": []
	}`)

	var rq UserRequest
	rq.skipPermissions = false
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
		"fields": ["active","randomParam"]
	}`)

	var rq UserRequest
	rq.skipPermissions = false
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}

	expected := UserRequest{
		Type:        UpdateRequest,
		IssuerId:    "USER",
		CertifierId: "ADMIN",
		Fields:      []string{"active"},
	}

	if !reflect.DeepEqual(rq, expected) {
		t.Errorf("Decoding produced different results:\n result: %v\n expected: %v\n", rq, expected)
	}
}

func TestDecodeMissingIssuerUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"certifierId": "ADMIN",
		"fields": ["active"]
	}`)

	var rq UserRequest
	rq.skipPermissions = false
	errs := rq.Decode(valid)

	if !(len(errs) == 1 && errs[0].Error() == "Issuer id missing") {
		t.Errorf("Decoding update with missing issuer should fail with one error, errors: %v", errs)
	}

	var rq2 UserRequest
	rq2.skipPermissions = true
	errs = rq2.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding update with missing issuer should not fail if checkPermissions is false, errors: %v", errs)
	}
}

func TestDecodeMissingCertifierUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"fields": ["active"]
	}`)

	var rq UserRequest
	rq.skipPermissions = false
	errs := rq.Decode(valid)

	if !(len(errs) == 1 && errs[0].Error() == "Certifier id missing") {
		t.Errorf("Decoding update with missing certifier should fail with one error, errors: %v", errs)
	}

	var rq2 UserRequest
	rq2.skipPermissions = true
	errs = rq2.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding update with missing certifier should not fail if skipPermissions is false, errors: %v", errs)
	}
}

func TestDecodeEncode(t *testing.T) {
	encKey := generatePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := generatePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	requestBaseStr := `
		"type": 0,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fields": [],
		"timestamp": "2018-01-13T23:53:00Z",
	`

	valid := []byte(`{
		` + requestBaseStr + `
		"data": {
			"id": "NEW_USER",
			"encKey": ` + encKeyStringEncoded + `,
			"signKey": ` + signKeyStringEncoded + `,
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
		}
	}`)

	var rq UserRequest
	rq.skipPermissions = false
	errs := rq.Decode(valid)

	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}

	userDataBytes, err := rq.Data.Encode()
	userDataJson := string(userDataBytes)

	if err != nil {
		t.Errorf("Object encoding failed, error: %v", err)
		return
	}

	reEncodedString := []byte(`{
		` + requestBaseStr + `
		"data": ` + userDataJson + `
	}`)

	var secondRq UserRequest
	secondDecodeErrs := secondRq.Decode(reEncodedString)

	if len(secondDecodeErrs) != 0 {
		t.Errorf("Re-encoded json failed while decoding, errors: %v", secondDecodeErrs)
		return
	}

	if !reflect.DeepEqual(rq, secondRq) {
		t.Errorf("%v\n", userDataJson)
		t.Errorf("Re-encoding produced different results:\n result: %v\n expected: %v\n", rq, secondRq)
	}
}
