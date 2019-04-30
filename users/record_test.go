package users

import (
	"github.com/mngharbi/DMPC/core"
	"reflect"
	"testing"
	"time"
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
		Key:       *core.GeneratePublicKey(),
		UpdatedAt: testRecordTime(),
	}
}

func generateBoolRecord(permissionDefault bool) booleanRecord {
	return booleanRecord{
		Ok:        permissionDefault,
		UpdatedAt: testRecordTime(),
	}
}

func testRecord(permissionDefault bool) userRecord {
	return userRecord{
		Id:      "id",
		EncKey:  generateKeyRecord(),
		SignKey: generateKeyRecord(),
		Permissions: permissionsRecord{
			Channel: channelPermissionsRecord{
				Add:       generateBoolRecord(permissionDefault),
				Read:      generateBoolRecord(permissionDefault),
				UpdatedAt: testRecordTime(),
			},
			User: userPermissionsRecord{
				Add:               generateBoolRecord(permissionDefault),
				Remove:            generateBoolRecord(permissionDefault),
				EncKeyUpdate:      generateBoolRecord(permissionDefault),
				SignKeyUpdate:     generateBoolRecord(permissionDefault),
				PermissionsUpdate: generateBoolRecord(permissionDefault),
				UpdatedAt:         testRecordTime(),
			},
			UpdatedAt: testRecordTime(),
		},
		Active:    generateBoolRecord(true),
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
		Type:      reqType,
		Timestamp: reqTime,
	}
}

func TestUpdateRequestActive(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Active.Ok = false
	expected.Active.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Active = false
	req.Fields = []string{"active"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating active field failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestActiveSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Active = false
	req.Fields = []string{"active"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating active field with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestEncKey(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.EncKey.Key = *core.GeneratePublicKey()
	expected.EncKey.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.encKeyObject = &expected.EncKey.Key
	req.Fields = []string{"encKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating encryption key failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestEncKeySkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.encKeyObject = core.GeneratePublicKey()
	req.Fields = []string{"encKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating encryption key with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestSignKey(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.SignKey.Key = *core.GeneratePublicKey()
	expected.SignKey.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.signKeyObject = &expected.SignKey.Key
	req.Fields = []string{"signKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating signing key failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestSignKeySkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.signKeyObject = core.GeneratePublicKey()
	req.Fields = []string{"signKey"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating signing key with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsChannelAdd(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Permissions.Channel.Add.Ok = false
	expected.Permissions.Channel.Add.UpdatedAt = testReqTime()
	expected.Permissions.Channel.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.Channel.Add = false
	req.Fields = []string{"permissions.channel.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating channel add permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsChannelAddSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.Channel.Add = false
	req.Fields = []string{"permissions.channel.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating channel add permission with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsChannelRead(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Permissions.Channel.Read.Ok = false
	expected.Permissions.Channel.Read.UpdatedAt = testReqTime()
	expected.Permissions.Channel.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.Channel.Read = false
	req.Fields = []string{"permissions.channel.read"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating channel read permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsChannelReadSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.Channel.Read = false
	req.Fields = []string{"permissions.channel.read"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating channel read permission with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserAdd(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Permissions.User.Add.Ok = false
	expected.Permissions.User.Add.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.Add = false
	req.Fields = []string{"permissions.user.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user add permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserAddSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.Add = false
	req.Fields = []string{"permissions.user.add"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user add permission with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserRemove(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Permissions.User.Remove.Ok = false
	expected.Permissions.User.Remove.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.Remove = false
	req.Fields = []string{"permissions.user.remove"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user remove permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserRemoveSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.Remove = false
	req.Fields = []string{"permissions.user.remove"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user remove permission with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserEncKeyUpdate(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Permissions.User.EncKeyUpdate.Ok = false
	expected.Permissions.User.EncKeyUpdate.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.EncKeyUpdate = false
	req.Fields = []string{"permissions.user.encKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user encryption key update permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserEncKeyUpdateSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.EncKeyUpdate = false
	req.Fields = []string{"permissions.user.encKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user encryption key update with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserSignKeyUpdate(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Permissions.User.SignKeyUpdate.Ok = false
	expected.Permissions.User.SignKeyUpdate.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.SignKeyUpdate = false
	req.Fields = []string{"permissions.user.signKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserSignKeyUpdateSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.SignKeyUpdate = false
	req.Fields = []string{"permissions.user.signKeyUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserPermissionsUpdate(t *testing.T) {
	obj := testRecord(true)

	expected := obj
	expected.Permissions.User.PermissionsUpdate.Ok = false
	expected.Permissions.User.PermissionsUpdate.UpdatedAt = testReqTime()
	expected.Permissions.User.UpdatedAt = testReqTime()
	expected.Permissions.UpdatedAt = testReqTime()
	expected.UpdatedAt = testReqTime()

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.PermissionsUpdate = false
	req.Fields = []string{"permissions.user.permissionsUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update permission failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestPermissionsUserPermissionsUpdateSkipped(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, true)
	req.Data.Permissions.User.PermissionsUpdate = false
	req.Fields = []string{"permissions.user.permissionsUpdate"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Updating user signature key update with stale request didn't fail.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestUpdateRequestInvalidUpdate(t *testing.T) {
	obj := testRecord(true)

	expected := obj

	req := testRequest(UpdateRequest, false)
	req.Data.Permissions.User.PermissionsUpdate = false
	req.Fields = []string{"random"}

	obj.applyUpdateRequest(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Update succeeded despite fields updated being invalid.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestCreateRequest(t *testing.T) {
	obj := userRecord{}

	expected := testRecord(true)

	req := testRequest(CreateRequest, false)
	req.Data.Id = "id"
	encKeyCopy := expected.EncKey.Key
	req.Data.encKeyObject = &encKeyCopy
	signKeyCopy := expected.SignKey.Key
	req.Data.signKeyObject = &signKeyCopy
	req.Data.Permissions.Channel.Add = true
	req.Data.Permissions.Channel.Read = true
	req.Data.Permissions.User.Add = true
	req.Data.Permissions.User.Remove = true
	req.Data.Permissions.User.EncKeyUpdate = true
	req.Data.Permissions.User.SignKeyUpdate = true
	req.Data.Permissions.User.PermissionsUpdate = true
	req.Data.Active = true
	req.Timestamp = testRecordTime()

	obj.create(&req)

	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("Creation using empty record failed.\n result: %v\n expected: %v\n", obj, expected)
	}
}

func TestAuthorizationCreate(t *testing.T) {
	obj := testRecord(true)
	obj.Permissions.User.Add.Ok = false

	req := testRequest(CreateRequest, false)
	req.Data.Id = "notId"
	encKeyCopy := obj.EncKey.Key
	req.Data.encKeyObject = &encKeyCopy
	signKeyCopy := obj.SignKey.Key
	req.Data.signKeyObject = &signKeyCopy
	req.Data.Permissions.Channel.Add = true
	req.Data.Permissions.Channel.Read = true
	req.Data.Permissions.User.Add = true
	req.Data.Permissions.User.Remove = true
	req.Data.Permissions.User.EncKeyUpdate = true
	req.Data.Permissions.User.SignKeyUpdate = true
	req.Data.Permissions.User.PermissionsUpdate = true
	req.Data.Active = true
	req.Timestamp = testRecordTime()

	if obj.isAuthorized(&req) {
		t.Errorf("Creation should require user add permissions.\n")
	}

	obj.Permissions.User.Add.Ok = true

	if !obj.isAuthorized(&req) {
		t.Errorf("Creation should only require user add permissions.\n")
	}
}

func TestAuthorizationUpdate(t *testing.T) {
	obj := testRecord(false)

	req := testRequest(UpdateRequest, false)
	req.Data.Id = "notId"
	encKeyCopy := obj.EncKey.Key
	req.Data.encKeyObject = &encKeyCopy
	signKeyCopy := obj.SignKey.Key
	req.Data.signKeyObject = &signKeyCopy
	req.Data.Permissions.Channel.Add = true
	req.Data.Permissions.Channel.Read = true
	req.Data.Permissions.User.Add = true
	req.Data.Permissions.User.Remove = true
	req.Data.Permissions.User.EncKeyUpdate = true
	req.Data.Permissions.User.SignKeyUpdate = true
	req.Data.Permissions.User.PermissionsUpdate = true
	req.Data.Active = true
	req.Timestamp = testRecordTime()
	req.Fields = []string{}

	if !obj.isAuthorized(&req) {
		t.Errorf("No update is authorized.\n")
	}

	req.Fields = []string{"active"}

	if obj.isAuthorized(&req) {
		t.Errorf("Update to active needs user delete permission.\n")
	}

	obj.Permissions.User.Remove.Ok = true

	if !obj.isAuthorized(&req) {
		t.Errorf("Update to active only needs user delete permission.\n")
	}

	req.Fields = []string{"active", "encKey"}

	if obj.isAuthorized(&req) {
		t.Errorf("Update to encryption key needs same user or encryption key permission.\n")
	}

	obj.Permissions.User.EncKeyUpdate.Ok = true

	if !obj.isAuthorized(&req) {
		t.Errorf("Update to encryption key needs same user or encryption key permission.\n")
	}

	obj.Permissions.User.EncKeyUpdate.Ok = false
	req.Data.Id = "id"

	if !obj.isAuthorized(&req) {
		t.Errorf("Update to encryption key needs same user or encryption key permission.\n")
	}

	req.Fields = []string{"active", "signKey"}
	req.Data.Id = "notId"

	if obj.isAuthorized(&req) {
		t.Errorf("Update to signature key needs same user or signature key permission.\n")
	}

	obj.Permissions.User.SignKeyUpdate.Ok = true

	if !obj.isAuthorized(&req) {
		t.Errorf("Update to signature key needs same user or signature key permission.\n")
	}

	obj.Permissions.User.SignKeyUpdate.Ok = false
	req.Data.Id = "id"

	if !obj.isAuthorized(&req) {
		t.Errorf("Update to signature key needs same user or signature key permission.\n")
	}

	req.Data.Id = "notId"

	permissionsFields := []string{"permissions.channel.add", "permissions.channel.read", "permissions.user.add",
		"permissions.user.remove", "permissions.user.encKeyUpdate",
		"permissions.user.signKeyUpdate", "permissions.user.permissionsUpdate"}

	for _, field := range permissionsFields {
		req.Fields = []string{field}
		obj.Permissions.User.PermissionsUpdate.Ok = false

		if obj.isAuthorized(&req) {
			t.Errorf("Update to %s needs permission update permission.\n", field)
		}

		obj.Permissions.User.PermissionsUpdate.Ok = true

		if !obj.isAuthorized(&req) {
			t.Errorf("Update to %s needs permission update permission.\n", field)
		}
	}
}
