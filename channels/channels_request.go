package channels

import (
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
	Result ChannelsStatusCode
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
	Permissions *ChannelPermissionsObject `json:"permissions"`
	Timestamp   time.Time                 `json:"timestamp"`
}

/*
	Validates and sanitizes request
*/
func (rq *OpenChannelRequest) sanitizeAndValidate() error {
	if len(rq.Id) == 0 ||
		len(rq.KeyId) == 0 ||
		rq.Permissions == nil ||
		len(rq.Permissions.Users) == 0 {
		return errors.New("Open channel request is invalid.")
	}
	return nil
}

/*
	Structure for close channel request
*/
type CloseChannelRequest struct {
	Id        string    `json:"id"`
	KeyId     string    `json:"keyId"`
	Timestamp time.Time `json:"timestamp"`
}

/*
	Validates and sanitizes request
*/
func (rq *CloseChannelRequest) sanitizeAndValidate() error {
	if len(rq.Id) == 0 ||
		len(rq.KeyId) == 0 {
		return errors.New("Close channel request is invalid.")
	}
	return nil
}
