package core

import (
	"encoding/json"
)

/*
	Structure of an operation before temporary encryption
*/
type OperationTemporaryEncrypted struct {
	Version    float64 `json:"version"`

	Encryption struct {
		Encrypted       bool              `json:"encrypted"`
		Challenges      map[string]string `json:"challenges"`
		Nonce           string            `json:"nonce"`
	} `json:"encryption"`

	Transmission json.RawMessage `json:"transmission"`

	Payload      string `json:"payload"`
}

/*
	Structure of an operation before permanent encryption
*/
type OperationPermanentEncrypted struct {
	Encryption struct {
		Encrypted       bool              `json:"encrypted"`
		KeyId           string            `json:"keyId"`
		Nonce           string            `json:"nonce"`
	} `json:"encryption"`

	Issue struct {
		Signature string `json:"signature"`
	} `json:"issue"`

	Certification struct {
		Signature string `json:"signature"`
	} `json:"certification"`

	Meta struct {
		RequestType int `json:"requestType"`
	} `json:"meta"`

	Payload      string `json:"payload"`
}

func (op *OperationTemporaryEncrypted) Decode(stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op); err != nil {
		return err
	}

	return nil
}

func (op *OperationTemporaryEncrypted) Encode() ([]byte, error) {
	jsonStream, _ := json.Marshal(op)

	return jsonStream, nil
}

func (op *OperationPermanentEncrypted) Decode(stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op); err != nil {
		return err
	}

	return nil
}

func (op *OperationPermanentEncrypted) Encode() ([]byte, error) {
	jsonStream, _ := json.Marshal(op)

	return jsonStream, nil
}
