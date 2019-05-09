package channels

import (
	"encoding/json"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"time"
)

/*
	Structure for channel actions response
*/
type ChannelsStatusCode int

const (
	ChannelsSuccess ChannelsStatusCode = iota
	ChannelsFailure
	BufferError
)

type ChannelsResponse struct {
	Result  ChannelsStatusCode `json:"result"`
	Channel *ChannelObject     `json:"channel"`
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
	Structure for read channel request
*/
type ReadChannelRequest struct {
	Id string
}

// *ReadChannelRequest -> Json
func (rq *ReadChannelRequest) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(rq)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

// Json -> *ReadChannelRequest
func (rq *ReadChannelRequest) Decode(stream []byte) error {
	return json.Unmarshal(stream, rq)
}

/*
	Validates and sanitizes request
*/
func (rq *ReadChannelRequest) sanitizeAndValidate() error {
	if len(rq.Id) == 0 {
		return errors.New("Read channel request is invalid.")
	}
	return nil
}

/*
	Structure for open channel request
*/
type OpenChannelRequest struct {
	Channel   *ChannelObject `json:"channel"`
	Signers   *core.VerifiedSigners
	Key       []byte    `json:"key"`
	Timestamp time.Time `json:"timestamp"`
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
	if rq.Channel == nil ||
		len(rq.Channel.Id) == 0 ||
		len(rq.Channel.KeyId) == 0 ||
		len(rq.Channel.Permissions.Users) == 0 ||
		len(rq.Key) == 0 {
		return errors.New("Open channel request is invalid.")
	}
	return nil
}

/*
	Structure for close channel request
*/
type CloseChannelRequest struct {
	Id        string
	Signers   *core.VerifiedSigners
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
