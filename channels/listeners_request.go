package channels

import (
	"encoding/json"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"time"
)

/*
	Structure for listener actions response
*/
type ListenersStatusCode int

const (
	ListenersSuccess ListenersStatusCode = iota
	ListenersFailure
	ListenersUnauthorized
)

type ListenersResponse struct {
	Result       ListenersStatusCode
	Channel      EventChannel
	SubscriberId string
}

/*
	Structure for listen request
*/
type SubscribeRequest struct {
	ChannelId string
	Signers   *core.VerifiedSigners
	Timestamp time.Time `json:"timestamp"`
}

// *SubscribeRequest -> Json
func (rq *SubscribeRequest) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(rq)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

// Json -> *SubscribeRequest
func (rq *SubscribeRequest) Decode(stream []byte) error {
	return json.Unmarshal(stream, rq)
}

/*
	Validates and sanitizes request
*/
func (rq *SubscribeRequest) sanitizeAndValidate() error {
	if len(rq.ChannelId) == 0 ||
		rq.Signers == nil {
		return errors.New("Listen request is invalid.")
	}
	return nil
}

/*
	Structure for unsubscribe request
*/
type UnsubscribeRequest struct {
	ChannelId    string
	SubscriberId string
	Timestamp    time.Time
}
