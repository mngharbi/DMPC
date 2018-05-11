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
