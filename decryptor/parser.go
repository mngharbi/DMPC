package decryptor

/*
	Request response structure
*/
const (
	Success = iota
	TemporaryDecryptionError
	PermanentDecryptionError
	ExecutorError
)

type DecryptorResponse struct {
	Result int    `json:"result"`
	Ticket string `json:"ticket"`
}
