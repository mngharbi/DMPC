/*
	Testing for user structure
	(en/de)coding/decoding and verification
	@TODO: Separate decoding and verification
*/
package users

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode"
)

/*
	Decoding
*/
func TestDecodeCreateRequest(t *testing.T) {
	encKey := core.GeneratePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := core.GeneratePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	valid := []byte(`{
		"type": 0,
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
					"read": false, 
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
	err := rq.Decode(valid)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	signers := generateGenericSigners()
	rq.addSigners(signers)
	errs := rq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("Sanitization failed, errors: %v", errs)
		return
	}

	expectedTimestamp, _ := time.Parse(time.RFC3339, "2018-01-13T23:53:00Z")
	expected := UserRequest{
		Type:    CreateRequest,
		Fields:  []string{},
		signers: signers,
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
					Read:              false,
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
		t.Errorf("Decoding produced different results:\n result: %+v\n expected: %+v\n", rq, expected)
	}

}

func TestDecodeUpdateRequest(t *testing.T) {
	encKey := core.GeneratePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := core.GeneratePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	valid := []byte(`{
		"type": 1,
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
					"read": false,
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
	err := rq.Decode(valid)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	signers := generateGenericSigners()
	rq.addSigners(signers)
	errs := rq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("Sanitization failed, errors: %v", errs)
		return
	}

	expectedTimestamp, _ := time.Parse(time.RFC3339, "2018-01-13T23:53:00Z")
	expected := UserRequest{
		Type:    UpdateRequest,
		Fields:  []string{"encKey", "signKey"},
		signers: signers,
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
					Read:              false,
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

func TestDecodeAndVerifyOneFieldUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"fields": ["active"],
		"data": {
			"active": false
		}
	}`)

	var rq UserRequest
	rq.skipPermissions = false
	err := rq.Decode(valid)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	signers := generateGenericSigners()
	rq.addSigners(signers)
	errs := rq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("Sanitization failed, errors: %v", errs)
		return
	}
}

func TestDecodeAndVerifyEmptyUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"fields": []
	}`)

	var rq UserRequest
	rq.skipPermissions = false
	err := rq.Decode(valid)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	signers := generateGenericSigners()
	rq.addSigners(signers)
	errs := rq.sanitizeAndCheckParams()
	if !(len(errs) == 1 && errs[0].Error() == noFieldsUpdatedErrorMsg) {
		t.Errorf("Decoding update with no fields updated should fail with one error, errors: %v", errs)
	}
}

func TestDecodeAndVerifyInvalidFieldsUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"fields": ["active","randomParam"]
	}`)

	var rq UserRequest
	rq.skipPermissions = false
	err := rq.Decode(valid)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	signers := generateGenericSigners()
	rq.addSigners(signers)
	errs := rq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("Decoding Failed, errors: %v", errs)
		return
	}

	expected := UserRequest{
		Type:    UpdateRequest,
		Fields:  []string{"active"},
		signers: signers,
	}

	if !reflect.DeepEqual(rq, expected) {
		t.Errorf("Decoding produced different results:\n result: %v\n expected: %v\n", rq, expected)
	}
}

func TestDecodeAndVerifyNoSigners(t *testing.T) {
	// Create valid user create request, and decode it
	valid, _ := generateUserCreateRequest("user", false, false, false, false, false, false, false, false)
	var rq UserRequest
	err := rq.Decode(valid)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	/*
		Without signers structure
	*/
	rq.addSigners(nil)

	// Non verified
	rq.skipPermissions = true
	errs := rq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("No signers, non verified failed. Errors: %v", errs)
	}

	// Verified
	rq.skipPermissions = false
	errs = rq.sanitizeAndCheckParams()
	if !(len(errs) == 1 && errs[0].Error() == signersMissingErrorMsg) {
		t.Errorf("No signers, verified failed. Errors: %v", errs)
	}

	/*
		No issuer
	*/
	signers := generateGenericSigners()
	signers.IssuerId = ""
	rq.addSigners(signers)

	// Non verified
	rq.skipPermissions = true
	errs = rq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("No issuer, non verified failed. Errors: %v", errs)
	}

	// Verified
	rq.skipPermissions = false
	errs = rq.sanitizeAndCheckParams()
	if !(len(errs) == 1 && errs[0].Error() == issuerIdMissingErrorMsg) {
		t.Errorf("No issuer, verified failed. Errors: %v", errs)
	}

	/*
		No certifier
	*/
	signers = generateGenericSigners()
	signers.CertifierId = ""
	rq.addSigners(signers)

	// Non verified
	rq.skipPermissions = true
	errs = rq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("No certifier, non verified failed. Errors: %v", errs)
	}

	// Verified
	rq.skipPermissions = false
	errs = rq.sanitizeAndCheckParams()
	if !(len(errs) == 1 && errs[0].Error() == certifierIdMissingErrorMsg) {
		t.Errorf("No certifier, verified failed. Errors: %v", errs)
	}
}

func TestDecodeAndVerifyEncode(t *testing.T) {
	encKey := core.GeneratePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := core.GeneratePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	requestBaseStr := `
		"type": 0,
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
					"read": false,
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
	err := rq.Decode(valid)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	signers := generateGenericSigners()
	rq.addSigners(signers)
	errs := rq.sanitizeAndCheckParams()
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
	err = secondRq.Decode(reEncodedString)
	if err != nil {
		t.Errorf("Decoding Failed, error: %v", err)
		return
	}

	secondRq.addSigners(signers)
	errs = secondRq.sanitizeAndCheckParams()
	if len(errs) != 0 {
		t.Errorf("Re-encoded json failed while decoding, errors: %v", errs)
		return
	}

	if !reflect.DeepEqual(rq, secondRq) {
		t.Errorf("%v\n", userDataJson)
		t.Errorf("Re-encoding produced different results:\n result: %v\n expected: %v\n", rq, secondRq)
	}
}

func stripAllWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func TestEncodeResponse(t *testing.T) {
	response := &UserResponse{
		Result: 0,
	}

	responseEncodedExpected := []byte(`{
		"result": 0,
		"data": null
	}`)

	responseEncoded, err := response.Encode()
	if err != nil {
		t.Errorf("Response encoding failed.")
		return
	}

	trimmedEncoded := []byte(stripAllWhitespace(string(responseEncoded)))
	trimmedEncodedExpected := []byte(stripAllWhitespace(string(responseEncodedExpected)))

	if !reflect.DeepEqual(trimmedEncoded, trimmedEncodedExpected) {
		t.Errorf("%v != expected: %v", string(trimmedEncoded), string(trimmedEncodedExpected))
		t.Errorf("Response encoding doesn't match expected encoding.")
		return
	}
}
