package config

/*
   Key parsing
*/

import (
	"crypto/rsa"
	"github.com/mngharbi/DMPC/core"
	"io/ioutil"
)

/*
   Generic public/private key parsing from file
*/
func GetPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	encodedKey, err := GetEncodedPrivateKey(filePath)
	if err != nil {
		return nil, err
	}
	return core.PrivateStringToAsymKey(encodedKey)
}

func GetPublicKey(filePath string) (*rsa.PublicKey, error) {
	encodedKey, err := GetEncodedPublicKey(filePath)
	if err != nil {
		return nil, err
	}
	return core.PublicStringToAsymKey(encodedKey)
}

func GetEncodedPublicKey(filePath string) (string, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func GetEncodedPrivateKey(filePath string) (string, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

/*
   Non encoded
*/
func (conf *Config) GetPublicEncryptionKey() (*rsa.PublicKey, error) {
	return GetPublicKey(conf.Paths.PublicEncryptionKeyPath)
}

func (conf *Config) GetPublicSigningKey() (*rsa.PublicKey, error) {
	return GetPublicKey(conf.Paths.PublicSigningKeyPath)
}

func (conf *Config) GetPrivateEncryptionKey() (*rsa.PrivateKey, error) {
	return GetPrivateKey(conf.Paths.PrivateEncryptionKeyPath)
}

func (conf *Config) GetPrivateSigningKey() (*rsa.PrivateKey, error) {
	return GetPrivateKey(conf.Paths.PrivateSigningKeyPath)
}

/*
   Encoded
*/
func (conf *Config) GetEncodedPublicEncryptionKey() (string, error) {
	return GetEncodedPublicKey(conf.Paths.PublicEncryptionKeyPath)
}

func (conf *Config) GetEncodedPublicSigningKey() (string, error) {
	return GetEncodedPublicKey(conf.Paths.PublicSigningKeyPath)
}

func (conf *Config) GetEncodedPrivateEncryptionKey() (string, error) {
	return GetEncodedPrivateKey(conf.Paths.PrivateEncryptionKeyPath)
}

func (conf *Config) GetEncodedPrivateSigningKey() (string, error) {
	return GetEncodedPrivateKey(conf.Paths.PrivateSigningKeyPath)
}
