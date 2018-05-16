package channels

import (
	"github.com/mngharbi/DMPC/core"
	"time"
)

/*
	Structure for messages daemon response
*/
type MessagesStatusCode int

const (
	MessagesSuccess MessagesStatusCode = iota
)

type MessagesResponse struct {
	Result MessagesStatusCode
}

/*
	Structure for add message request
*/
type AddMessageRequest struct {
	Timestamp time.Time
	Signers   *core.VerifiedSigners
	ChannelId string
	Message   []byte
}

/*
	Validates and sanitizes request
*/
func (rq *AddMessageRequest) sanitizeAndValidate() error {
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
	return nil
}
