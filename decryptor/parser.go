package decryptor

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
	Result int    `json:"result"`
	Ticket string `json:"ticket"`
}
