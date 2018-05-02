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
