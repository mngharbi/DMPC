package main

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
func (config *Config) getPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	encodedKey, err := config.getEncodedPrivateKey(filePath)
	if err != nil {
		return nil, err
	}
	return core.PrivateStringToAsymKey(encodedKey)
}

func (config *Config) getPublicKey(filePath string) (*rsa.PublicKey, error) {
	encodedKey, err := config.getEncodedPublicKey(filePath)
	if err != nil {
		return nil, err
	}
	return core.PublicStringToAsymKey(encodedKey)
}

func (config *Config) getEncodedPublicKey(filePath string) (string, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (config *Config) getEncodedPrivateKey(filePath string) (string, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

/*
   Non encoded
*/
func (config *Config) getPublicEncryptionKey() (*rsa.PublicKey, error) {
	return config.getPublicKey(config.Paths.PublicEncryptionKeyPath)
}

func (config *Config) getPublicSigningKey() (*rsa.PublicKey, error) {
	return config.getPublicKey(config.Paths.PublicSigningKeyPath)
}

func (config *Config) getPrivateEncryptionKey() (*rsa.PrivateKey, error) {
	return config.getPrivateKey(config.Paths.PrivateEncryptionKeyPath)
}

func (config *Config) getPrivateSigningKey() (*rsa.PrivateKey, error) {
	return config.getPrivateKey(config.Paths.PrivateSigningKeyPath)
}

/*
   Encoded
*/
func (config *Config) getEncodedPublicEncryptionKey() (string, error) {
	return config.getEncodedPublicKey(config.Paths.PublicEncryptionKeyPath)
}

func (config *Config) getEncodedPublicSigningKey() (string, error) {
	return config.getEncodedPublicKey(config.Paths.PublicSigningKeyPath)
}

func (config *Config) getEncodedPrivateEncryptionKey() (string, error) {
	return config.getEncodedPrivateKey(config.Paths.PrivateEncryptionKeyPath)
}

func (config *Config) getEncodedPrivateSigningKey() (string, error) {
	return config.getEncodedPrivateKey(config.Paths.PrivateSigningKeyPath)
}
