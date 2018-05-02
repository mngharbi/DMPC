package core

import (
	"encoding/json"
)

/*
	Structure of an operation before temporary encryption
*/
type TemporaryEncryptionFields struct {
	Encrypted  bool              `json:"encrypted"`
	Challenges map[string]string `json:"challenges"`
	Nonce      string            `json:"nonce"`
}
type TemporaryEncryptedOperation struct {
	Version float64 `json:"version"`

	Encryption TemporaryEncryptionFields `json:"encryption"`

	Transmission json.RawMessage `json:"transmission"`

	Payload string `json:"payload"`
}

/*
	Structure of an operation before permanent encryption
*/
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
	// @TODO: Change to its own type
	RequestType int `json:"requestType"`
}
type PermanentEncryptedOperation struct {
	Encryption PermanentEncryptionFields `json:"encryption"`

	Issue PermanentAuthenticationFields `json:"issue"`

	Certification PermanentAuthenticationFields `json:"certification"`

	Meta PermanentMetaFields `json:"meta"`

	Payload string `json:"payload"`
}

func (op *TemporaryEncryptedOperation) Decode(stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op); err != nil {
		return err
	}

	return nil
}

func (op *TemporaryEncryptedOperation) Encode() ([]byte, error) {
	jsonStream, _ := json.Marshal(op)

	return jsonStream, nil
}

func (op *PermanentEncryptedOperation) Decode(stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op); err != nil {
		return err
	}

	return nil
}

func (op *PermanentEncryptedOperation) Encode() ([]byte, error) {
	jsonStream, _ := json.Marshal(op)

	return jsonStream, nil
}
