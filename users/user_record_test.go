package users

import (
	"time"
	"testing"
	"reflect"
)

func testReqPastTime() time.Time {
	return time.Date(2016, time.January, 0, 0, 0, 0, 0, time.UTC)
}

func testRecordTime() time.Time {
	return time.Date(2017, time.January, 0, 0, 0, 0, 0, time.UTC)
}

func testReqTime() time.Time {
	return time.Date(2018, time.January, 0, 0, 0, 0, 0, time.UTC)
}

func generateKeyRecord() keyRecord {
	return keyRecord{
		Key: *generatePublicKey(),
		UpdatedAt: testRecordTime(),
	}
}

func generateBoolRecord() booleanRecord {
	return booleanRecord{
		Ok: true,
		UpdatedAt: testRecordTime(),
	}
}

func testRecord() userRecord {
	return userRecord{
		Id: "id",
		EncKey: generateKeyRecord(),
		SignKey: generateKeyRecord(),
		Permissions: permissionsRecord{
			Channel: channelPermissionsRecord{
				Add: generateBoolRecord(),
				UpdatedAt: testRecordTime(),
			},
			User: userPermissionsRecord{
				Add: generateBoolRecord(),
				Remove: generateBoolRecord(),
				EncKeyUpdate: generateBoolRecord(),
				SignKeyUpdate: generateBoolRecord(),
				PermissionsUpdate: generateBoolRecord(),
				UpdatedAt: testRecordTime(),
			},
		},
		Active: generateBoolRecord(),
		CreatedAt: testRecordTime(),
		UpdatedAt: testRecordTime(),
	}
}

func testRequest(reqType int, late bool) UserRequest {
	var reqTime time.Time
	if late {
		reqTime = testReqPastTime()
	} else {
		reqTime = testReqTime()
	}

	return UserRequest{
		Type: reqType,
		IssuerId: "issuer",
		CertifierId: "certifier",
		Timestamp: reqTime,
	}
}

func TestUpdateRequestActive(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.Active.Ok = false
	expected.Active.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Active = false
	req.FieldsUpdated = []string{"active"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating active field failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestActiveSkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Active = false
	req.FieldsUpdated = []string{"active"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating active field with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestEncKey(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.EncKey.Key = *generatePublicKey()
	expected.EncKey.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.encKeyObject = &expected.EncKey.Key
	req.FieldsUpdated = []string{"encKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating encryption key failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestEncKeySkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.encKeyObject = generatePublicKey()
	req.FieldsUpdated = []string{"encKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating encryption key with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestSignKey(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.SignKey.Key = *generatePublicKey()
	expected.SignKey.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.signKeyObject = &expected.SignKey.Key
	req.FieldsUpdated = []string{"signKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating signing key failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestSignKeySkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.signKeyObject = generatePublicKey()
	req.FieldsUpdated = []string{"signKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating signing key with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsChannelAdd(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.Permissions.Channel.Add.Ok = false
	expected.Permissions.Channel.Add.UpdatedAt = testReqTime()
	expected.Permissions.Channel.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.Channel.Add = false
	req.FieldsUpdated = []string{"permissions.channel.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating channel add permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsChannelAddSkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.Channel.Add = false
	req.FieldsUpdated = []string{"permissions.channel.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating channel add permission with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserAdd(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.Permissions.User.Add.Ok = false
	expected.Permissions.User.Add.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.Add = false
	req.FieldsUpdated = []string{"permissions.user.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user add permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserAddSkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.Add = false
	req.FieldsUpdated = []string{"permissions.user.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user add permission with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserRemove(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.Permissions.User.Remove.Ok = false
	expected.Permissions.User.Remove.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.Remove = false
	req.FieldsUpdated = []string{"permissions.user.remove"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user remove permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserRemoveSkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.Remove = false
	req.FieldsUpdated = []string{"permissions.user.remove"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user remove permission with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserEncKeyUpdate(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.Permissions.User.EncKeyUpdate.Ok = false
	expected.Permissions.User.EncKeyUpdate.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.EncKeyUpdate = false
	req.FieldsUpdated = []string{"permissions.user.encKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user encryption key update permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserEncKeyUpdateSkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.EncKeyUpdate = false
	req.FieldsUpdated = []string{"permissions.user.encKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user encryption key update with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserSignKeyUpdate(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.Permissions.User.SignKeyUpdate.Ok = false
	expected.Permissions.User.SignKeyUpdate.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.SignKeyUpdate = false
	req.FieldsUpdated = []string{"permissions.user.signKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserSignKeyUpdateSkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.SignKeyUpdate = false
	req.FieldsUpdated = []string{"permissions.user.signKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserPermissionsUpdate(t *testing.T) {
	obj := testRecord()

	expected := obj
	expected.Permissions.User.PermissionsUpdate.Ok = false
	expected.Permissions.User.PermissionsUpdate.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.PermissionsUpdate = false
	req.FieldsUpdated = []string{"permissions.user.permissionsUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserPermissionsUpdateSkipped(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.PermissionsUpdate = false
	req.FieldsUpdated = []string{"permissions.user.permissionsUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestInvalidUpdate(t *testing.T) {
	obj := testRecord()

	expected := obj

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.PermissionsUpdate = false
	req.FieldsUpdated = []string{"random"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Update succeeded despite fields updated being invalid.\n result: %v\n expected: %v\n", obj, expected)
	}
}
