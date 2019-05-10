package channels

import (
	"encoding/json"
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
	Result MessagesStatusCode `json:"result"`
}

// *MessagesResponse -> Json
func (resp *MessagesResponse) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(resp)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}

/*
	Structure for add message request
*/
type AddMessageRequest struct {
	ChannelId  string
	Timestamp  time.Time
	Signers    *core.VerifiedSigners
	Message    []byte
	rawMessage []byte
}

func (rq *AddMessageRequest) decodeMessage() error {
	var base64encodedMessage string
	err := json.Unmarshal(rq.Message, &base64encodedMessage)
	if err != nil {
		return err
	}
	rq.rawMessage, err = core.Base64DecodeString(base64encodedMessage)
	return err
}

func EncodeMessage(message []byte) []byte {
	return core.CiphertextEncode(message)
}

/*
	Validates and sanitizes request
*/
func (rq *AddMessageRequest) sanitizeAndValidate() error {
	valid := rq.Signers != nil &&
		len(rq.Message) > 0 &&
		len(rq.ChannelId) > 0 &&
		rq.decodeMessage() == nil
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
