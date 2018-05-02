package users

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"time"
)

/*
	Error messages
*/

const (
	unknownRequestTypeErrorMsg string = "Unknown request type"
	signersMissingErrorMsg     string = "Signers missing"
	issuerIdMissingErrorMsg    string = "Issuer id missing"
	certifierIdMissingErrorMsg string = "Certifier id missing"
	noFieldsUpdatedErrorMsg    string = "No fields updated"
	noSubjectsErrorMsg         string = "No users requested"
)

/*
	External structure of a user
*/
type ChannelPermissionsObject struct {
	Add bool `json:"add"`
}

type UserPermissionsObject struct {
	Add               bool `json:"add"`
	Remove            bool `json:"remove"`
	EncKeyUpdate      bool `json:"encKeyUpdate"`
	SignKeyUpdate     bool `json:"signKeyUpdate"`
	PermissionsUpdate bool `json:"permissionsUpdate"`
}
type PermissionsObject struct {
	Channel ChannelPermissionsObject `json:"channel"`
	User    UserPermissionsObject    `json:"user"`
}
type UserObject struct {
	Id     string `json:"id"`
	EncKey string `json:"encKey"`
	// @TODO: Make it possible to pass this directly
	encKeyObject  *rsa.PublicKey
	SignKey       string `json:"signKey"`
	signKeyObject *rsa.PublicKey
	Permissions   PermissionsObject `json:"permissions"`
	Active        bool              `json:"active"`
	CreatedAt     time.Time         `json:"createdAt"`
	DisabledAt    time.Time         `json:"disabledAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

/*
	External structure of a user related request
*/
const (
	CreateRequest = iota
	UpdateRequest
	ReadRequest
)

// @TODO: Change Type to enumerated type
type UserRequest struct {
	Type      int        `json:"type"`
	Fields    []string   `json:"fields"`
	Data      UserObject `json:"data"`
	Timestamp time.Time  `json:"timestamp"`
	signers   *core.VerifiedSigners

	// Private settings
	skipPermissions bool
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
	Result int `json:"result"`
	// @TODO: Consider returning pointers after benchmarking
	Data []UserObject `json:"data"`
}

/*
	User request creation/checking
*/

// Json -> *UserRequest
func (rq *UserRequest) Decode(stream []byte) error {
	return json.Unmarshal(stream, rq)
}

func (rq *UserRequest) addSigners(signers *core.VerifiedSigners) {
	rq.signers = signers
}

// Used to correct request and return errors if irreparable
func (rq *UserRequest) sanitizeAndCheckParams() []error {
	res := []error{}

	// Verify type, issuer, and certifier
	if !(CreateRequest <= rq.Type && rq.Type <= ReadRequest) {
		res = append(res, errors.New(unknownRequestTypeErrorMsg))
	}

	if !rq.skipPermissions {
		if rq.signers == nil {
			res = append(res, errors.New(signersMissingErrorMsg))
		} else if len(rq.signers.IssuerId) == 0 {
			res = append(res, errors.New(issuerIdMissingErrorMsg))
		} else if len(rq.signers.CertifierId) == 0 {
			res = append(res, errors.New(certifierIdMissingErrorMsg))
		}
	}

	switch rq.Type {

	// For create requests, clear fields updated, and parse public keys
	case CreateRequest:
		rq.Fields = []string{}

		if parsedKey, err := core.PublicStringToAsymKey(rq.Data.EncKey); err == nil {
			rq.Data.encKeyObject = parsedKey
		} else {
			res = append(res, err)
		}
		if parsedKey, err := core.PublicStringToAsymKey(rq.Data.SignKey); err == nil {
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
			if parsedKey, err := core.PublicStringToAsymKey(rq.Data.EncKey); err == nil {
				rq.Data.encKeyObject = parsedKey
			} else {
				res = append(res, err)
			}
		}
		if contains(rq.Fields, "signKey") {
			if parsedKey, err := core.PublicStringToAsymKey(rq.Data.SignKey); err == nil {
				rq.Data.signKeyObject = parsedKey
			} else {
				res = append(res, err)
			}
		}

		if len(rq.Fields) == 0 {
			res = append(res, errors.New(noFieldsUpdatedErrorMsg))
		}

	/*
		For read requests:
			* Check there are user ids requested
	*/
	case ReadRequest:
		if len(rq.Fields) == 0 {
			res = append(res, errors.New(noSubjectsErrorMsg))
		}
	}

	return res
}

var sanitizeFieldsUpdatedAllowed map[string]bool = map[string]bool{
	"encKey":                             true,
	"signKey":                            true,
	"permissions.channel.add":            true,
	"permissions.user.add":               true,
	"permissions.user.remove":            true,
	"permissions.user.encKeyUpdate":      true,
	"permissions.user.signKeyUpdate":     true,
	"permissions.user.permissionsUpdate": true,
	"active": true,
}

func (rq *UserRequest) sanitizeFieldsUpdated() {
	newSlice := make([]string, 0)
	for _, field := range rq.Fields {
		if sanitizeFieldsUpdatedAllowed[field] {
			newSlice = append(newSlice, field)
		}
	}
	rq.Fields = newSlice
}

/*
	User request encoding
*/
func (usr *UserRequest) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(usr)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
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
	User response encoding
*/
func (resp *UserResponse) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(resp)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}
