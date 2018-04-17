package users

import (
	"crypto/rsa"
	"sync"
	"time"
)

/*
	Record of a user
	Keeps track of granual timestamps for changes
*/
type keyRecord struct {
	Key       rsa.PublicKey
	UpdatedAt time.Time
}
type booleanRecord struct {
	Ok        bool
	UpdatedAt time.Time
}

type permissionsRecord struct {
	Channel   channelPermissionsRecord
	User      userPermissionsRecord
	UpdatedAt time.Time
}

type channelPermissionsRecord struct {
	Add       booleanRecord
	UpdatedAt time.Time
}

type userPermissionsRecord struct {
	Add               booleanRecord
	Remove            booleanRecord
	EncKeyUpdate      booleanRecord
	SignKeyUpdate     booleanRecord
	PermissionsUpdate booleanRecord
	UpdatedAt         time.Time
}

type userRecord struct {
	Id          string
	EncKey      keyRecord
	SignKey     keyRecord
	Permissions permissionsRecord
	Active      booleanRecord
	CreatedAt   time.Time
	UpdatedAt   time.Time
	lock        *sync.RWMutex
}

func (rec *userRecord) Less(index string, than interface{}) bool {
	switch index {
	case "id":
		return rec.Id < than.(*userRecord).Id
	}
	return false
}

/*
	Record update (run in a mutex context)
*/
func (record *userRecord) applyUpdateRequest(req *UserRequest) {
	for _, field := range req.Fields {
		switch field {
		case "active":
			if record.Active.update(req.Data.Active, req.Timestamp) {
				record.UpdatedAt = req.Timestamp
			}
		case "encKey":
			if record.EncKey.update(*req.Data.encKeyObject, req.Timestamp) {
				record.UpdatedAt = req.Timestamp
			}
		case "signKey":
			if record.SignKey.update(*req.Data.signKeyObject, req.Timestamp) {
				record.UpdatedAt = req.Timestamp
			}
		case "permissions.channel.add":
			if record.Permissions.Channel.Add.update(req.Data.Permissions.Channel.Add, req.Timestamp) {
				record.UpdatedAt = req.Timestamp
				record.Permissions.UpdatedAt = req.Timestamp
				record.Permissions.Channel.UpdatedAt = req.Timestamp
			}

		case "permissions.user.add", "permissions.user.remove", "permissions.user.encKeyUpdate", "permissions.user.signKeyUpdate", "permissions.user.permissionsUpdate":
			var perm *booleanRecord
			var reqVal bool
			switch field {
			case "permissions.user.add":
				perm = &record.Permissions.User.Add
				reqVal = req.Data.Permissions.User.Add
			case "permissions.user.remove":
				perm = &record.Permissions.User.Remove
				reqVal = req.Data.Permissions.User.Remove
			case "permissions.user.encKeyUpdate":
				perm = &record.Permissions.User.EncKeyUpdate
				reqVal = req.Data.Permissions.User.EncKeyUpdate
			case "permissions.user.signKeyUpdate":
				perm = &record.Permissions.User.SignKeyUpdate
				reqVal = req.Data.Permissions.User.SignKeyUpdate
			case "permissions.user.permissionsUpdate":
				perm = &record.Permissions.User.PermissionsUpdate
				reqVal = req.Data.Permissions.User.PermissionsUpdate
			}

			if perm.update(reqVal, req.Timestamp) {
				record.UpdatedAt = req.Timestamp
				record.Permissions.UpdatedAt = req.Timestamp
				record.Permissions.User.UpdatedAt = req.Timestamp
			}
		}
	}
}

func (perm *booleanRecord) update(val bool, time time.Time) bool {
	if time.After(perm.UpdatedAt) {
		perm.Ok = val
		perm.UpdatedAt = time
		return true
	}
	return false
}

func (keyRec *keyRecord) update(val rsa.PublicKey, time time.Time) bool {
	if time.After(keyRec.UpdatedAt) {
		keyRec.Key = val
		keyRec.UpdatedAt = time
		return true
	}
	return false
}

/*
	Create user record from creation request
*/
func (record *userRecord) create(req *UserRequest) {
	// Id
	record.Id = req.Data.Id

	// Active
	record.Active.update(req.Data.Active, req.Timestamp)

	/*
		Keys
	*/

	// Encryption key
	record.EncKey.update(*req.Data.encKeyObject, req.Timestamp)

	// Signature key
	record.SignKey.update(*req.Data.signKeyObject, req.Timestamp)

	/*
		Permissions
	*/

	// Permissions: Channel add
	record.Permissions.Channel.Add.update(req.Data.Permissions.Channel.Add, req.Timestamp)

	// Permissions: User add
	record.Permissions.User.Add.update(req.Data.Permissions.User.Add, req.Timestamp)

	// Permissions: User remove
	record.Permissions.User.Remove.update(req.Data.Permissions.User.Remove, req.Timestamp)

	// Permissions: User Encryption Key Update
	record.Permissions.User.EncKeyUpdate.update(req.Data.Permissions.User.EncKeyUpdate, req.Timestamp)

	// Permissions: User Signature Key Update
	record.Permissions.User.SignKeyUpdate.update(req.Data.Permissions.User.SignKeyUpdate, req.Timestamp)

	// Permissions: User Permissions Update
	record.Permissions.User.PermissionsUpdate.update(req.Data.Permissions.User.PermissionsUpdate, req.Timestamp)

	/*
		Timestamps
	*/
	record.Permissions.UpdatedAt = req.Timestamp
	record.Permissions.Channel.UpdatedAt = req.Timestamp
	record.Permissions.User.UpdatedAt = req.Timestamp
	record.UpdatedAt = req.Timestamp
	record.CreatedAt = req.Timestamp
}

/*
	Check permissions on request
*/
func (record *userRecord) isAuthorized(req *UserRequest) bool {
	result := true

	switch req.Type {
	case CreateRequest:
		// For creation we need to check user add permission
		result = record.Permissions.User.Add.Ok

	case UpdateRequest:
		isSameUser := req.Data.Id == record.Id

		for _, field := range req.Fields {
			if !result {
				break
			}
			switch field {
			case "active":
				result = record.Permissions.User.Remove.Ok
			case "encKey":
				result = record.Permissions.User.EncKeyUpdate.Ok || isSameUser
			case "signKey":
				result = record.Permissions.User.SignKeyUpdate.Ok || isSameUser
			case "permissions.channel.add", "permissions.user.add",
				"permissions.user.remove", "permissions.user.encKeyUpdate",
				"permissions.user.signKeyUpdate", "permissions.user.permissionsUpdate":
				result = record.Permissions.User.PermissionsUpdate.Ok
			}
		}
	}

	return result
}

/*
	Record locking
*/

// Read lock
func (record *userRecord) RLock() {
	record.lock.RLock()
}

// Read unlock
func (record *userRecord) RUnlock() {
	record.lock.RUnlock()
}

// Write lock
func (record *userRecord) Lock() {
	record.lock.Lock()
}

// Write unlock
func (record *userRecord) Unlock() {
	record.lock.Unlock()
}
