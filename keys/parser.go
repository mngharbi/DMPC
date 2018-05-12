package keys

import (
	"github.com/mngharbi/DMPC/core"
)

/*
	Key request structure
*/
type keyRequestType int

const (
	AddKeyRequest keyRequestType = iota
	DecryptRequest
)

type keyRequest struct {
	Type    keyRequestType
	KeyId   string
	Payload []byte
	Nonce   []byte
}

/*
	Make record from creation request
*/
func (req *keyRequest) makeRecord() *keyRecord {
	return &keyRecord{
		Id:  req.KeyId,
		Key: req.Payload,
	}
}

/*
	Make search record from creation request
*/
func (req *keyRequest) makeSearchRecord() *keyRecord {
	return &keyRecord{
		Id: req.KeyId,
	}
}

/*
	Validates request format
*/
func (req *keyRequest) validate() bool {
	if len(req.KeyId) == 0 {
		return false
	}

	switch req.Type {
	case AddKeyRequest:
		return len(req.Payload) == core.SymmetricKeySize
	case DecryptRequest:
		return len(req.Nonce) == core.SymmetricNonceSize
	}

	return false
}

/*
	Key response structure
*/
type keyResponseCode int

const (
	Success keyResponseCode = iota
	DecryptionFailure
)

type keyResponse struct {
	Result    keyResponseCode
	Decrypted []byte
}
