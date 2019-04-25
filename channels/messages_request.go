package channels

import (
	"errors"
	"github.com/mngharbi/DMPC/core"
	"time"
)

/*
	Structure for messages daemon response
*/
type MessagesStatusCode int

const (
	MessagesSuccess MessagesStatusCode = iota
	MessagesDropped
	MessagesBufferError
)

type MessagesResponse struct {
	Result MessagesStatusCode
}

/*
	Structure for add message request
*/
type AddMessageRequest struct {
	ChannelId string
	Timestamp time.Time
	Signers   *core.VerifiedSigners
	Message   []byte
}

/*
	Validates and sanitizes request
*/
func (rq *AddMessageRequest) sanitizeAndValidate() error {
	valid := rq.Signers != nil &&
		len(rq.Message) > 0 &&
		len(rq.ChannelId) > 0
	if !valid {
		return errors.New("Add message request is invalid.")
	}
	return nil
}

/*
	Structure for buffer operation request
*/
type BufferOperationRequest struct {
	Operation *core.Operation
}

/*
	Validates and sanitizes request
*/
func (rq *BufferOperationRequest) sanitizeAndValidate() error {
	if rq.Operation == nil {
		return errors.New("Buffer operation request is invalid.")
	} else {
		rq.Operation.Meta.Buffered = true
		return nil
	}
}
