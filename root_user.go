package main

/*
	Utilities for getting the root user from config
*/

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/users"
	"io/ioutil"
	"time"
)

/*
   Error messages
*/
const (
	parseUserWithoutKeysError    string = "Could not find user object file"
	parseEncryptionError         string = "Invalid signing key file"
	parseSigningError            string = "Invalid encryption key file"
	encodeRootUserOperationError string = "Unable to encode root user operation"
)

/*
   Utilities
*/
func (config *Config) getRootUserFilePath() string {
	return config.Paths.RootUserFilePath
}

func (config *Config) getRootUserObjectWithoutKeys() (*users.UserObject, error) {
	raw, err := ioutil.ReadFile(config.getRootUserFilePath())
	if err != nil {
		return nil, err
	}

	var userObj users.UserObject
	json.Unmarshal(raw, &userObj)
	return &userObj, nil
}

func (config *Config) getRootUserObjectWithKeys() *users.UserObject {
	// Get user object without keys
	userObj, err := config.getRootUserObjectWithoutKeys()
	if err != nil {
		log.Fatalf(parseUserWithoutKeysError)
	}

	// Get public keys for root user from config
	userObj.SignKey, err = config.getEncodedPublicSigningKey()
	if err != nil {
		log.Fatalf(parseSigningError)
	}
	userObj.EncKey, err = config.getEncodedPublicEncryptionKey()
	if err != nil {
		log.Fatalf(parseEncryptionError)
	}

	return userObj
}

func buildRootUserOperation(config *Config) *core.TemporaryEncryptedOperation {
	// Get root user object from configuration
	log.Debugf("Parsing root user object from configuration")
	rootUserObject := config.getRootUserObjectWithKeys()

	// Build user request
	encodedCreateRequest, err := users.GenerateCreateRequest(rootUserObject, time.Now()).Encode()
	if err != nil {
		log.Fatalf(encodeRootUserOperationError)
	}

	// Build operation
	permanentEncryptedEncoded, err := core.GeneratePermanentEncryptedOperation(
		// Non encrypted
		false, "", nil, true,
		// No issuer
		"", nil, true,
		// No certifier
		"", nil, true,
		// non base64 encoded payload and meta
		executor.UsersRequest, encodedCreateRequest, false,
	).Encode()
	if err != nil {
		log.Fatalf(encodeRootUserOperationError)
	}
	return core.GenerateTemporaryEncryptedOperation(
		// Non encrypted
		false, nil, nil, true,
		// non base64 encoded payload
		permanentEncryptedEncoded, false,
	)
}

func createRootUser(operation *core.TemporaryEncryptedOperation) string {
	operationEncoded, err := operation.Encode()
	if err != nil {
		log.Fatalf("Error encoding temporary encrypted operation for root user.")
	}
	rootUserChannel, errs := decryptor.MakeUnverifiedRequest(operationEncoded)
	if errs != nil {
		log.Fatalf("Error making root user creation request.")
	}
	rootUserNativeResp := <-rootUserChannel
	rootUserResp := (*rootUserNativeResp).(*decryptor.DecryptorResponse)
	if rootUserResp.Result != decryptor.Success {
		log.Fatalf("Error making root user creation request.")
	}
	return rootUserResp.Ticket
}
