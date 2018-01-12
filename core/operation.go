package core

import (
	"encoding/json"
)

/*
	Structure of an operation as it comes in
*/
type RawOperation struct {
	Version float64 `json:"version"`

	Encryption struct {

		// Temporary encryption
		Temp struct {

			Encrypted bool `json:"encrypted"`

			Keys map[string]string `json:"keys"`

			Nonce string `json:"nonce"`

		} `json:"temp"`

		// Permanent encryption
		Perm struct {

			Encrypted bool `json:"encrypted"`

			KeyId string `json:"keyId"`

			Nonce string `json:"nonce"`

		}  `json:"perm"`

	}  `json:"encryption"`

	Issue struct {

		Id string `json:"id"`

		Signature string `json:"signature"`

	} `json:"issue"`

	Certification struct {

		Id string `json:"id"`

		Signature string `json:"signature"`

	} `json:"certification"`

	Transmission json.RawMessage `json:"transmission"`

	Payload json.RawMessage `json:"payload"`
}

/*
	Structure of an operation augmented
	with information on whether decryption/verification occured
*/
type Operation struct {
	Raw *RawOperation

	TempDecrypted bool

	PermDecrypted bool

	IssueVerified bool

	CertificationVerified bool
}


func (op *Operation) Decode (stream []byte) error {
	// Try to decode json into raw operation
	if err := json.Unmarshal(stream, &op.Raw); err != nil {
		return err
	}

	return nil
}

func (op *Operation) Encode () ([]byte, error) {
	jsonStream, err := json.Marshal(op.Raw)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}