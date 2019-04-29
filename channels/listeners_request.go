package channels

import (
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
