package users

import (
	"encoding/json"
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
	Id 			string 				`json:"id"`
	EncKey		string 				`json:"encKey"`
	SignKey 	string 				`json:"signKey"`
	Permissions PermissionsObject 	`json:"permissions"`
	Active 		bool 				`json:"active"`
	CreatedAt 	time.Time 			`json:"createdAt"`
	DisabledAt 	time.Time 			`json:"disabledAt"`
	UpdatedAt 	time.Time 			`json:"updatedAt"`
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
		case CreateRequest:
			rq.FieldsUpdated = []string{}
		case DeleteRequest:
			rq.FieldsUpdated = []string{}
		case UpdateRequest:
			rq.sanitizeFieldsUpdated()

			if len(rq.FieldsUpdated) == 0 {
				res = append(res, errors.New("No fields updated"))
			}
	}

	return res
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
