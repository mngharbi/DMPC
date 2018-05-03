package decryptor

import (
	"github.com/mngharbi/DMPC/status"
)

/*
	Request response structure
*/
const (
	Success = iota
	TemporaryDecryptionError
	PermanentDecryptionError
	VerificationError
	ExecutorError
)

type DecryptorResponse struct {
	Result int           `json:"result"`
	Ticket status.Ticket `json:"ticket"`
}
