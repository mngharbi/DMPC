package core

import (
	"encoding/json"
)

/*
	Structure of a transaction (before decryption)
*/

type PipelineConfig struct {
	ReadStatusUpdates bool `json:"read_status_updates"`
	ReadResult        bool `json:"read_result"`
}

type TransactionEncryptionFields struct {
	Encrypted  bool              `json:"encrypted"`
	Challenges map[string]string `json:"challenges"`
	Nonce      string            `json:"nonce"`
}

type Transaction struct {
	// DMPC version
	Version float64 `json:"version"`

	// In transit encryption
	Encryption TransactionEncryptionFields `json:"encryption"`

	// Generic structure that can be used for  for transmission
	Transmission json.RawMessage `json:"transmission"`

	// Ingestion options
	Pipeline PipelineConfig `json:"pipeline"`

	// Encoded and possibly encrypted payload/operation
	Payload json.RawMessage `json:"payload"`
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
