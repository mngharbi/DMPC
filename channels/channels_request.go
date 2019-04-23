package channels

import (
	"encoding/json"
	"errors"
	"time"
)

/*
	Structure for channel actions response
*/
type ChannelsStatusCode int

const (
	ChannelsSuccess ChannelsStatusCode = iota
)

type ChannelsResponse struct {
	Result ChannelsStatusCode `json:"result"`
}

// *ChannelsResponse -> Json
func (resp *ChannelsResponse) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(resp)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

/*
	Structure for open channel request
*/
type ChannelPermissionObject struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Close bool `json:"close"`
}
type ChannelPermissionsObject struct {
	Users map[string]*ChannelPermissionObject `json:"users"`
}
type OpenChannelRequest struct {
	Id          string                    `json:"id"`
	KeyId       string                    `json:"keyId"`
	Key         []byte                    `json:"key"`
	Permissions *ChannelPermissionsObject `json:"permissions"`
	Timestamp   time.Time                 `json:"timestamp"`
}

// *OpenChannelRequest -> Json
func (rq *OpenChannelRequest) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(rq)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

// Json -> *OpenChannelRequest
func (rq *OpenChannelRequest) Decode(stream []byte) error {
	return json.Unmarshal(stream, rq)
}

/*
	Validates and sanitizes request
*/
func (rq *OpenChannelRequest) sanitizeAndValidate() error {
	if len(rq.Id) == 0 ||
		len(rq.KeyId) == 0 ||
		rq.Permissions == nil ||
		len(rq.Permissions.Users) == 0 ||
		len(rq.Key) == 0 {
		return errors.New("Open channel request is invalid.")
	}
	return nil
}

/*
	Structure for close channel request
*/
type CloseChannelRequest struct {
	Id        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

// *CloseChannelRequest -> Json
func (rq *CloseChannelRequest) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(rq)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

// Json -> *CloseChannelRequest
func (rq *CloseChannelRequest) Decode(stream []byte) error {
	return json.Unmarshal(stream, rq)
}

/*
	Validates and sanitizes request
*/
func (rq *CloseChannelRequest) sanitizeAndValidate() error {
	if len(rq.Id) == 0 {
		return errors.New("Close channel request is invalid.")
	}
	return nil
}
