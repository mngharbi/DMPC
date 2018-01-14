package users

import (
	"testing"
	"time"
	"reflect"
)

func TestDecodeCreateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 0,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": [],
		"data": {
			"id": "NEW_USER",
			"encKey": "encKey",
			"signKey": "signKey",
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
			EncKey: "encKey",
			SignKey: "signKey",
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
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": ["encKey"],
		"data": {
			"id": "NEW_USER",
			"encKey": "encKey",
			"signKey": "signKey",
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
		FieldsUpdated: []string{"encKey"},
		Data: UserObject{
			Id: "NEW_USER",
			EncKey: "encKey",
			SignKey: "signKey",
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

func TestDecodeEmptyUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": []
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) == 0 {
		t.Errorf("Decoding update with no fields updated should fail")
		return
	}
}

func TestDecodeInvalidFieldsUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"certifierId": "ADMIN",
		"fieldsUpdated": ["encKey","signKey","randomParam"]
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
		FieldsUpdated: []string{"encKey","signKey"},
	}

	if !reflect.DeepEqual(rq, expected) {
		t.Errorf("Decoding produced different results:\n result: %v\n expected: %v\n", rq, expected)
	}
}

func TestDecodeMissingIssuerUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"certifierId": "ADMIN",
		"fieldsUpdated": ["encKey","signKey","randomParam"]
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) == 0 {
		t.Errorf("Decoding should fail because of missing issuer")
	}
}

func TestDecodeMissingCertifierUpdateRequest(t *testing.T) {
	valid := []byte(`{
		"type": 1,
		"issuerId": "USER",
		"fieldsUpdated": ["encKey","signKey","randomParam"]
	}`)

	var rq UserRequest
	errs := rq.Decode(valid)

	if len(errs) == 0 {
		t.Errorf("Decoding should fail because of missing issuer")
	}
}

