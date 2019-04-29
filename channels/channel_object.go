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
	Users map[string]*ChannelPermissionObject `json:"users"`
}
type ChannelObject struct {
	Id          string                    `json:"id"`
	KeyId       string                    `json:"keyId"`
	Permissions *ChannelPermissionsObject `json:"permissions"`
	State       ChannelObjectState        `json:"state"`
}

// *ChannelObject -> Json
func (rq *ChannelObject) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(rq)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

// Json -> *ChannelObject
func (rq *ChannelObject) Decode(stream []byte) error {
	return json.Unmarshal(stream, rq)
}
