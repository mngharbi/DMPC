package users

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"github.com/mngharbi/DMPC/core"
	"time"
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
	Id            string `json:"id"`
	EncKey        string `json:"encKey"`
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

type UserRequest struct {
	Type        int        `json:"type"`
	IssuerId    string     `json:"issuerId"`
	CertifierId string     `json:"certifierId"`
	Fields      []string   `json:"fields"`
	Data        UserObject `json:"data"`
	Timestamp   time.Time  `json:"timestamp"`

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
	Result int
	// @TODO: Consider returning pointers after benchmarking
	Data []UserObject
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

	// Verify type, issuer, and certifier
	if !(CreateRequest <= rq.Type && rq.Type <= ReadRequest) {
		res = append(res, errors.New("Unknown request type"))
	}
	if !rq.skipPermissions && len(rq.IssuerId) == 0 {
		res = append(res, errors.New("Issuer id missing"))
	}
	if !rq.skipPermissions && len(rq.CertifierId) == 0 {
		res = append(res, errors.New("Certifier id missing"))
	}

	switch rq.Type {

	// For create requests, clear fields updated, and parse public keys
	case CreateRequest:
		rq.Fields = []string{}

		if parsedKey, err := core.StringToAsymKey(rq.Data.EncKey); err == nil {
			rq.Data.encKeyObject = parsedKey
		} else {
			res = append(res, err)
		}
		if parsedKey, err := core.StringToAsymKey(rq.Data.SignKey); err == nil {
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
			if parsedKey, err := core.StringToAsymKey(rq.Data.EncKey); err == nil {
				rq.Data.encKeyObject = parsedKey
			} else {
				res = append(res, err)
			}
		}
		if contains(rq.Fields, "signKey") {
			if parsedKey, err := core.StringToAsymKey(rq.Data.SignKey); err == nil {
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
	User object encoding
*/
func (usr *UserObject) Encode() ([]byte, error) {
	jsonStream, err := json.Marshal(usr)

	if err != nil {
		return nil, err
	}

	return jsonStream, nil
}
