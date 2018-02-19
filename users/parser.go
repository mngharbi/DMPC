package users

import (
	"encoding/json"
	"encoding/pem"
	"crypto/rsa"
	"crypto/x509"
	"bytes"
	"errors"
	"time"
)

/*
	External structure of a user
*/
type ChannelPermissionsObject struct {
	Add		bool	`json:"add"`
}

type UserPermissionsObject struct {
	Add					bool	`json:"add"`
	Remove				bool	`json:"remove"`
	EncKeyUpdate		bool	`json:"encKeyUpdate"`
	SignKeyUpdate		bool	`json:"signKeyUpdate"`
	PermissionsUpdate	bool	`json:"permissionsUpdate"`
}
type PermissionsObject struct {
	Channel	ChannelPermissionsObject	`json:"channel"`
	User	UserPermissionsObject		`json:"user"`
}
type UserObject struct {
	Id				string				`json:"id"`
	EncKey			string				`json:"encKey"`
	encKeyObject	*rsa.PublicKey
	SignKey			string				`json:"signKey"`
	signKeyObject	*rsa.PublicKey
	Permissions		PermissionsObject	`json:"permissions"`
	Active			bool				`json:"active"`
	CreatedAt		time.Time			`json:"createdAt"`
	DisabledAt		time.Time			`json:"disabledAt"`
	UpdatedAt		time.Time			`json:"updatedAt"`
}

/*
	External structure of a user related request
*/
const (
	CreateRequest = iota
	UpdateRequest
	ReadRequest
)
type UserRequest struct {
	Type			int			`json:"type"`
	IssuerId		string		`json:"issuerId"`
	CertifierId		string		`json:"certifierId"`
	Fields			[]string	`json:"fields"`
	Data			UserObject	`json:"data"`
	Timestamp		time.Time	`json:"timestamp"`
}

/*
	External structure of a user related response
*/
const (
	Success = iota
	IssuerUnknownError
	CertifierUnknownError
	SubjectUnknownError
	CertifierPermissionsError
	UnlockingFailedError
)
type UserResponse struct {
	Result			int
	Data 			[]UserObject
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

// Used to correct request and return errors if irreparable
func (rq *UserRequest) sanitizeAndCheckParams() []error {
	res := []error{}

	// Verifies type, issuer and certifier
	if !(CreateRequest <= rq.Type && rq.Type <= ReadRequest) {
		res = append(res, errors.New("Unknown request type"))
	}
	if len(rq.IssuerId) == 0 {
		res = append(res, errors.New("Issuer id missing"))
	}
	if len(rq.CertifierId) == 0 {
		res = append(res, errors.New("Certifier id missing"))
	}

	switch rq.Type {

		// For create requests, clear fields updated, and parse public keys
		case CreateRequest:
			rq.Fields = []string{}

			if parsedKey,err := convertRsaStringToKey(rq.Data.EncKey); err == nil {
				rq.Data.encKeyObject = parsedKey
			} else {
				res = append(res, err)
			}
			if parsedKey,err := convertRsaStringToKey(rq.Data.SignKey); err == nil {
				rq.Data.signKeyObject = parsedKey
			} else {
				res = append(res, err)
			}

		/*
		For update requests:
			* Only leave valid fields updated
			* Check there are updates
			* Parse public keys if any
		*/
		case UpdateRequest:
			rq.sanitizeFieldsUpdated()

			if contains(rq.Fields, "encKey") {
				if parsedKey,err := convertRsaStringToKey(rq.Data.EncKey); err == nil {
					rq.Data.encKeyObject = parsedKey
				} else {
					res = append(res, err)
				}
			}
			if contains(rq.Fields, "signKey") {
				if parsedKey,err := convertRsaStringToKey(rq.Data.SignKey); err == nil {
					rq.Data.signKeyObject = parsedKey
				} else {
					res = append(res, err)
				}
			}

			if len(rq.Fields) == 0 {
				res = append(res, errors.New("No fields updated"))
			}

		/*
		For read requests:
			* Check there are user ids requested
		*/
		case ReadRequest:
			if len(rq.Fields) == 0 {
				res = append(res, errors.New("No users requested"))
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
	for _,field := range rq.Fields {
		if sanitizeFieldsUpdatedAllowed[field] {
			newSlice = append(newSlice, field)
		}
	}
	rq.Fields = newSlice
}

/*
	User object encoding
*/
func (usr *UserObject) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(usr)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}


/*
	Utilities
*/

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
}

func makeSearchByIdRecord(usr *UserObject) *userRecord {
	return &userRecord {
		Id: usr.Id,
	}
}

func convertKeyToString(key *rsa.PublicKey) string {
	// Break into bytes
	keyBytes,_ := x509.MarshalPKIXPublicKey(key)

	// Build pem block containing public key
	block := &pem.Block{
		Type: "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}

	// PEM encode block
	buf := new(bytes.Buffer)
	pem.Encode(buf, block)

	// Return string representing bytes
	return string(pem.EncodeToMemory(block))
}
