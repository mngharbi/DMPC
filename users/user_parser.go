package users

import (
	"encoding/json"
	"encoding/pem"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"time"
)

/*
	External structure of a user
*/
type ChannelPermissionsObject struct {
	Add bool `json:"add"`
}

type UserPermissionsObject struct {
	Add 				bool `json:"add"`
	Remove 				bool `json:"remove"`
	EncKeyUpdate 		bool `json:"encKeyUpdate"`
	SignKeyUpdate 		bool `json:"signKeyUpdate"`
	PermissionsUpdate 	bool `json:"permissionsUpdate"`
}
type PermissionsObject struct {
	Channel ChannelPermissionsObject 	`json:"channel"`
	User UserPermissionsObject 			`json:"user"`

}
type UserObject struct {
	Id 			 string 				`json:"id"`
	EncKey		 string 				`json:"encKey"`
	encKeyObject *rsa.PublicKey
	SignKey 	 string 				`json:"signKey"`
	signKeyObject *rsa.PublicKey
	Permissions  PermissionsObject 	`json:"permissions"`
	Active 		 bool 				`json:"active"`
	CreatedAt 	 time.Time 			`json:"createdAt"`
	DisabledAt 	 time.Time 			`json:"disabledAt"`
	UpdatedAt 	 time.Time 			`json:"updatedAt"`
}

/*
	External structure of a user related request
*/
const (
    CreateRequest = iota
    UpdateRequest
    DeleteRequest
)
type UserRequest struct {
	Type 			int  		`json:"type"`
	IssuerId 		string 		`json:"issuerId"`
	CertifierId 	string 		`json:"certifierId"`
	FieldsUpdated 	[]string 	`json:"fieldsUpdated"`
	Data 			UserObject 	`json:"data"`
	Timestamp 		time.Time 	`json:"timestamp"`
}

/*
	User request creation/checking
*/

// Json -> *UserRequest
func (rq *UserRequest) Decode(stream []byte) []error {
	// Parse json
	if err := json.Unmarshal(stream, rq); err != nil {
		return []error{err}
	}

	return rq.sanitizeAndCheckParams()
}

func (rq *UserRequest) sanitizeAndCheckParams() []error {
	res := []error{}

	if !(CreateRequest <= rq.Type && rq.Type <= DeleteRequest) {
		res = append(res, errors.New("Unknown request type"))
	}

	if len(rq.IssuerId) == 0 {
		res = append(res, errors.New("Issuer id missing"))
	}

	if len(rq.CertifierId) == 0 {
		res = append(res, errors.New("Certifier id missing"))
	}

	switch rq.Type {
		case CreateRequest, DeleteRequest:
			rq.FieldsUpdated = []string{}
			parseKey(rq.Data.encKeyObject, rq.Data.EncKey, res)
			parseKey(rq.Data.signKeyObject, rq.Data.SignKey, res)

		case UpdateRequest:
			rq.sanitizeFieldsUpdated()

			if contains(rq.FieldsUpdated, "encKey") {
				parseKey(rq.Data.encKeyObject, rq.Data.EncKey, res)
			}
			if contains(rq.FieldsUpdated, "signKey") {
				parseKey(rq.Data.signKeyObject, rq.Data.SignKey, res)
			}

			if len(rq.FieldsUpdated) == 0 {
				res = append(res, errors.New("No fields updated"))
			}
	}

	return res
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func parseKey(key *rsa.PublicKey, str string, errs []error) {
	key, err := convertRsaStringToKey(str)
	if err != nil {
		errs = append(errs, err)
	}
}

var sanitizeFieldsUpdatedAllowed map[string]bool = map[string]bool{
	"encKey": true,
	"signKey": true,
	"permissions.channel.add": true,
	"permissions.user.add": true,
	"permissions.user.remove": true,
	"permissions.user.encKeyUpdate": true,
	"permissions.user.signKeyUpdate": true,
	"permissions.user.permissionsUpdate": true,
	"active": true,
}

func (rq *UserRequest) sanitizeFieldsUpdated() {
	newSlice := make([]string, 0)
	for _,field := range rq.FieldsUpdated {
		if sanitizeFieldsUpdatedAllowed[field] {
			newSlice = append(newSlice, field)
		}
	}
	rq.FieldsUpdated = newSlice
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
