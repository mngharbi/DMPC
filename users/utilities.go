/*
	Utilities
*/

package users

import (
	"crypto/rsa"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/memstore"
	"time"
)

/*
	Make creation record
*/
func GenerateCreateRequest(userObject *UserObject, timestamp time.Time) *UserRequest {
	return &UserRequest{
		Type:      CreateRequest,
		Data:      *userObject,
		Timestamp: timestamp,
	}
}

/*
	Generic function to read multiple users' attribute
*/
const (
	genericRequestFailureErrorMsg string = "Unable to make request to retrieve user records"
	genericUserNotFoundErrorMsg   string = "Unable to find at least one of the users"
)

func getGenericUserAttributeByIds(ids []string, handleAttribute func(*UserObject)) error {
	// Make unverified request for user
	rq := &UserRequest{
		Type:       ReadRequest,
		Fields:     ids,
		ReadLock:   true,
		ReadUnlock: true,
	}
	rq.skipPermissions = true
	channel, errs := makeRequest(rq)
	if len(errs) != 0 {
		return errors.New(genericRequestFailureErrorMsg)
	}

	// Wait for response
	resp := <-channel
	if resp == nil || resp.Result != Success {
		return errors.New(genericRequestFailureErrorMsg)
	} else if resp.Data == nil || len(resp.Data) != len(ids) {
		return errors.New(genericUserNotFoundErrorMsg)
	} else {
		for _, userObject := range resp.Data {
			handleAttribute(&userObject)
		}
		return nil
	}
}

/*
	Gets signing keys by user ids
*/
func GetSigningKeysById(ids []string) ([]*rsa.PublicKey, error) {
	var keys []*rsa.PublicKey
	handleSingingKeyLambda := func(obj *UserObject) {
		keys = append(keys, obj.signKeyObject)
	}
	if err := getGenericUserAttributeByIds(ids, handleSingingKeyLambda); err != nil {
		return nil, err
	}
	return keys, nil
}

/*
	Gets global channel permissions by user ids
*/
func GetChannelPermissionsByIds(ids []string) ([]*ChannelPermissionsObject, error) {
	var permissions []*ChannelPermissionsObject
	handlePermissionsLambda := func(obj *UserObject) {
		permissions = append(permissions, &obj.Permissions.Channel)
	}
	if err := getGenericUserAttributeByIds(ids, handlePermissionsLambda); err != nil {
		return nil, err
	}
	return permissions, nil
}

/*
	Read records from store
*/
func readUserRecordsByIds(store *memstore.Memstore, ids []string) []*userRecord {
	result := []*userRecord{}

	for _, id := range ids {
		searchRecord := makeSearchByIdRecord(id)
		itemResult := store.Get(searchRecord, idIndexStr)
		if itemResult != nil {
			result = append(result, itemResult.(*userRecord))
		} else {
			result = append(result, nil)
		}
	}

	return result
}

// Make a user object from a user record
func (usr *UserObject) createFromRecord(rec *userRecord) {
	usr.Id = rec.Id
	usr.encKeyObject = &rec.EncKey.Key
	usr.EncKey = core.PublicAsymKeyToString(&rec.EncKey.Key)
	usr.signKeyObject = &rec.SignKey.Key
	usr.SignKey = core.PublicAsymKeyToString(&rec.SignKey.Key)
	usr.Permissions.Channel.Add = rec.Permissions.Channel.Add.Ok
	usr.Permissions.Channel.Read = rec.Permissions.Channel.Read.Ok
	usr.Permissions.User.Add = rec.Permissions.User.Add.Ok
	usr.Permissions.User.Read = rec.Permissions.User.Read.Ok
	usr.Permissions.User.Remove = rec.Permissions.User.Remove.Ok
	usr.Permissions.User.EncKeyUpdate = rec.Permissions.User.EncKeyUpdate.Ok
	usr.Permissions.User.SignKeyUpdate = rec.Permissions.User.SignKeyUpdate.Ok
	usr.Permissions.User.PermissionsUpdate = rec.Permissions.User.PermissionsUpdate.Ok
	usr.Active = rec.Active.Ok
	if usr.Active {
		usr.DisabledAt = rec.Active.UpdatedAt
	}
	usr.CreatedAt = rec.CreatedAt
	usr.UpdatedAt = rec.UpdatedAt
}

// Make a dummy user record pointer for search from a user object
func (usr *UserObject) makeSearchByIdRecord() memstore.Item {
	return &userRecord{
		Id: usr.Id,
	}
}

// Make a dummy user record pointer for search from an id
func makeSearchByIdRecord(id string) memstore.Item {
	return &userRecord{
		Id: id,
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
