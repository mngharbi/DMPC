/*
	Communication structures
*/

package core

import (
	"crypto/rsa"
)

/*
	Function called to shutdown main daemon
*/
type ShutdownLambda func()

/*
	Function to get a signing key for a user given its id
*/
type UsersSignKeyRequester func([]string) ([]*rsa.PublicKey, error)

/*
	Function to get a symmetric key given its id
*/
type KeyRequester func(string) []byte

/*
	Function to add key to keys subsystem
*/
type KeyAdder func(keyId string, key []byte) error

/*
	Function to decrypt by key id
*/
type Decryptor func(keyId string, nonce []byte, ciphertext []byte) ([]byte, error)
