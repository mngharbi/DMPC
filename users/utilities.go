/*
	Utilities
*/

package users

import (
	"github.com/mngharbi/DMPC/core"
)

// Make a user object from a user record
func (usr *UserObject) createFromRecord(rec *userRecord) {
	usr.Id = rec.Id
	usr.EncKey = core.AsymKeyToString(&rec.EncKey.Key)
	usr.SignKey = core.AsymKeyToString(&rec.SignKey.Key)
	usr.Permissions.Channel.Add = rec.Permissions.Channel.Add.Ok
	usr.Permissions.User.Add = rec.Permissions.User.Add.Ok
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
func makeSearchByIdRecord(usr *UserObject) *userRecord {
	return &userRecord{
		Id: usr.Id,
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
