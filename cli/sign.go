package cli

import (
	"log"
)

/*
	Main sign operation function
*/
func SignOperation(issue bool, certify bool) {
	if !IsFunctional() {
		return
	}

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

	// Read operation from stdin
	op := ReadOperation()

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

	// Write operation to stdout
	WriteOperation(op)
}
