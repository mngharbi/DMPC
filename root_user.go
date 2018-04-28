package main

/*
	Utilities for getting the root user from config
*/

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/users"
	"io/ioutil"
)

/*
   Error messages
*/
const (
	parseUserWithoutKeysError string = "Could not find user object file"
	parseEncryptionError      string = "Invalid signing key file"
	parseSigningError         string = "Invalid encryption key file"
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

func (config *Config) getRootUserObject() *users.UserObject {
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
