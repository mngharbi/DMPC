package channels

import (
	"encoding/json"
)

/*
	Readable state
*/
type ChannelObjectState string

const (
	ChannelObjectBufferedState     ChannelObjectState = "buffered"
	ChannelObjectOpenState         ChannelObjectState = "open"
	ChannelObjectClosedState       ChannelObjectState = "closed"
	ChannelObjectInconsistentState ChannelObjectState = "inconsistent"
)

/*
	Structure for channel object
*/
type ChannelPermissionObject struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Close bool `json:"close"`
}
type ChannelPermissionsObject struct {
	Users map[string]ChannelPermissionObject `json:"users"`
}
type ChannelObject struct {
	Id          string                   `json:"id"`
	KeyId       string                   `json:"keyId"`
	Permissions ChannelPermissionsObject `json:"permissions"`
	State       ChannelObjectState       `json:"state"`
}

// *ChannelObject -> Json
func (obj *ChannelObject) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(obj)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

// Json -> *ChannelObject
func (obj *ChannelObject) Decode(stream []byte) error {
	return json.Unmarshal(stream, obj)
}

/*
	Utilities
*/

func (obj *ChannelObject) buildFromRecord(rec *channelRecord) {
	obj.Id = rec.id
	obj.KeyId = rec.keyId
	if rec.permissions != nil {
		obj.Permissions = ChannelPermissionsObject{}
		obj.Permissions.buildFromRecord(rec.permissions)
	}
	obj.State = objectStateMapping[rec.state]
}

func (obj *ChannelPermissionsObject) buildFromRecord(rec *channelPermissionsRecord) {
	obj.Users = map[string]ChannelPermissionObject{}
	for userId, userPermissionsRec := range rec.users {
		obj.Users[userId] = ChannelPermissionObject{
			Read:  userPermissionsRec.read,
			Write: userPermissionsRec.write,
			Close: userPermissionsRec.close,
		}
	}
}
