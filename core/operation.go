package core

import (
	"encoding/json"
)

/*
	Structure of a transaction (before temporary decryption)
*/
type TransactionEncryptionFields struct {
	Encrypted  bool              `json:"encrypted"`
	Challenges map[string]string `json:"challenges"`
	Nonce      string            `json:"nonce"`
}
type Transaction struct {
	Version float64 `json:"version"`

	Encryption TransactionEncryptionFields `json:"encryption"`

	Transmission json.RawMessage `json:"transmission"`

	Payload string `json:"payload"`
}

/*
	Decodes a transaction
*/
func (op *Transaction) Decode(stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op); err != nil {
		return err
	}

	return nil
}

/*
	Encodes a transaction
*/
func (op *Transaction) Encode() ([]byte, error) {
	jsonStream, _ := json.Marshal(op)

	return jsonStream, nil
}

/*
	Structure of an operation before permanent encryption
*/
type RequestType int

const (
	UsersRequestType RequestType = iota
	AddMessageType
)

type PermanentEncryptionFields struct {
	Encrypted bool   `json:"encrypted"`
	KeyId     string `json:"keyId"`
	Nonce     string `json:"nonce"`
}
type PermanentAuthenticationFields struct {
	Id        string `json:"id"`
	Signature string `json:"signature"`
}
type PermanentMetaFields struct {
	RequestType RequestType `json:"requestType"`
	Buffered    bool
}
type PermanentEncryptedOperation struct {
	Encryption PermanentEncryptionFields `json:"encryption"`

	Issue PermanentAuthenticationFields `json:"issue"`

	Certification PermanentAuthenticationFields `json:"certification"`

	Meta PermanentMetaFields `json:"meta"`

	Payload string `json:"payload"`
}

/*
	Determines if the request should be dropped if decryption/signature verification fails
*/
func (op *PermanentEncryptedOperation) ShouldDrop() bool {
	return !(op.Meta.RequestType == AddMessageType && !op.Meta.Buffered)
}

/*
	Decodes an operation
*/
func (op *PermanentEncryptedOperation) Decode(stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op); err != nil {
		return err
	}

	return nil
}

/*
	Encodes an operation
*/
func (op *PermanentEncryptedOperation) Encode() ([]byte, error) {
	jsonStream, _ := json.Marshal(op)

	return jsonStream, nil
}
