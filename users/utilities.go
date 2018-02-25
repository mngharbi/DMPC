/*
	Utilities
*/

package users

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// Make a user object from a user record
func (usr *UserObject) createFromRecord(rec *userRecord) {
	usr.Id = rec.Id
	usr.EncKey = convertKeyToString(&rec.EncKey.Key)
	usr.SignKey = convertKeyToString(&rec.SignKey.Key)
	usr.Permissions.Channel.Add = rec.Permissions.Channel.Add.Ok
	usr.Permissions.User.Add = rec.Permissions.User.Add.Ok
	usr.Permissions.User.Remove = rec.Permissions.User.Remove.Ok
	usr.Permissions.User.EncKeyUpdate = rec.Permissions.User.EncKeyUpdate.Ok
	usr.Permissions.User.SignKeyUpdate = rec.Permissions.User.SignKeyUpdate.Ok
	usr.Permissions.User.PermissionsUpdate = rec.Permissions.User.PermissionsUpdate.Ok
	usr.Active = rec.Active.Ok
	if usr.Active {
		usr.DisabledAt = rec.Active.UpdatedAt
	}
	usr.CreatedAt = rec.UpdatedAt
	usr.UpdatedAt = rec.UpdatedAt
}

// Make a dummy user record pointer for search from a user object
func makeSearchByIdRecord(usr *UserObject) *userRecord {
	return &userRecord{
		Id: usr.Id,
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func convertKeyToString(key *rsa.PublicKey) string {
	// Break into bytes
	keyBytes, _ := x509.MarshalPKIXPublicKey(key)

	// Build pem block containing public key
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}

	// PEM encode block
	buf := new(bytes.Buffer)
	pem.Encode(buf, block)

	// Return string representing bytes
	return string(pem.EncodeToMemory(block))
}

func convertRsaStringToKey(rsaString string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(rsaString))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse DER encoded public key: " + err.Error())
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("unknown type of public key" + err.Error())
	}
}
