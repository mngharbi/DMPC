package decryptor

/*
	Request response structure
*/
const (
	Success = iota
	TemporaryDecryptionError
	PermanentDecryptionError
)

type DecryptorResponse struct {
	Result int `json:"result"`
	Ticket int `json:"ticket"`
}
