package config

/*
	Utilities for getting the root user from conf
*/

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/users"
	"io/ioutil"
	"log"
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
func (conf *Config) GetRootUserFilePath() string {
	return conf.Paths.RootUserFilePath
}

func (conf *Config) getRootUserObjectWithoutKeys() (*users.UserObject, error) {
	raw, err := ioutil.ReadFile(conf.GetRootUserFilePath())
	if err != nil {
		return nil, err
	}

	var userObj users.UserObject
	json.Unmarshal(raw, &userObj)
	return &userObj, nil
}

func (conf *Config) GetRootUserObject() *users.UserObject {
	// Get user object without keys
	userObj, err := conf.getRootUserObjectWithoutKeys()
	if err != nil {
		log.Fatalf(parseUserWithoutKeysError)
	}

	// Get public keys for root user from conf
	userObj.SignKey, err = conf.GetEncodedPublicSigningKey()
	if err != nil {
		log.Fatalf(parseSigningError)
	}
	userObj.EncKey, err = conf.GetEncodedPublicEncryptionKey()
	if err != nil {
		log.Fatalf(parseEncryptionError)
	}

	return userObj
}
