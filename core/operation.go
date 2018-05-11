package core

import (
	"encoding/json"
)

/*
	Request types
*/
type RequestType int

const (
	UsersRequestType RequestType = iota
	AddMessageType
)

/*
	Structure of an operation before permanent encryption
*/
type OperationEncryptionFields struct {
	Encrypted bool   `json:"encrypted"`
	KeyId     string `json:"keyId"`
	Nonce     string `json:"nonce"`
}
type OperationAuthenticationFields struct {
	Id        string `json:"id"`
	Signature string `json:"signature"`
}
type OperationMetaFields struct {
	RequestType RequestType `json:"requestType"`
	Buffered    bool
}
type Operation struct {
	Encryption    OperationEncryptionFields     `json:"encryption"`
	Issue         OperationAuthenticationFields `json:"issue"`
	Certification OperationAuthenticationFields `json:"certification"`
	Meta          OperationMetaFields           `json:"meta"`
	Payload       string                        `json:"payload"`
}

/*
	Determines if the request should be dropped if decryption/signature verification fails
*/
func (op *Operation) ShouldDrop() bool {
	return !(op.Meta.RequestType == AddMessageType && !op.Meta.Buffered)
}

/*
	Decodes an operation
*/
func (op *Operation) Decode(stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op); err != nil {
		return err
	}

	return nil
}

/*
	Encodes an operation
*/
func (op *Operation) Encode() ([]byte, error) {
	jsonStream, _ := json.Marshal(op)

	return jsonStream, nil
}
