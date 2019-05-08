package cli

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"log"
	"os"
)

/*
	Read operation
*/
func ReadOperation() (op *core.Operation) {
	op = &core.Operation{}
	err := json.NewDecoder(os.Stdin).Decode(op)
	if err != nil {
		log.Fatal(err.Error())
	}
	return
}

/*
	Write operation
*/
func WriteOperation(op *core.Operation) {
	json.NewEncoder(os.Stdout).Encode(op)
}

/*
	Sign operation using root user
*/
func RootSignOperation(op *core.Operation, issue bool, certify bool) {
	// Get configuration structure
	conf := GetConfig()

	// Get root user object
	userObj, err := conf.getRootUserObjectWithoutKeys()
	if err != nil {
		log.Fatalf(parseUserWithoutKeysError)
	}

	// Get root user signing key
	key, err := conf.GetPrivateSigningKey()
	if err != nil {
		log.Fatalf(parseSigningError)
	}

	// Sign operation
	if issue {
		if err = op.IssuerSign(key, userObj.Id); err != nil {
			log.Fatalf(err.Error())
		}
	}
	if certify {
		if err = op.CertifierSign(key, userObj.Id); err != nil {
			log.Fatalf(err.Error())
		}
	}
}

/*
	Main sign operation function
*/
func ReadAndSignOperation(issue bool, certify bool) {
	if !IsFunctional() {
		return
	}

	// Read operation from stdin
	op := ReadOperation()

	// Do sign
	RootSignOperation(op, issue, certify)

	// Write operation to stdout
	WriteOperation(op)
}
