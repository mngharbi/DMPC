package decryptor

import (
	"github.com/mngharbi/DMPC/status"
)

/*
	Request response structure
*/
const (
	Success = iota
	TransactionDecryptionError
	PermanentDecryptionError
	VerificationError
	ExecutorError
)

type DecryptorResponse struct {
	// @TODO: Result should be typed
	Result int           `json:"result"`
	Ticket status.Ticket `json:"ticket"`
}
