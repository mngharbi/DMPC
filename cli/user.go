package cli

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
	"log"
	"os"
	"time"
)

/*
	Read user object from stdin
*/
func ReadUserObject() (obj *users.UserObject) {
	obj = &users.UserObject{}
	err := json.NewDecoder(os.Stdin).Decode(obj)
	if err != nil {
		log.Fatal(err.Error())
	}
	return
}

/*
	Write user object to stdout
*/
func WriteUserObject(obj *users.UserObject) {
	json.NewEncoder(os.Stdout).Encode(obj)
}

/*
	Generate user object
*/
func GenerateUserObject() {
	userId := cliGetString("Enter the user id:")

	// Get permissions
	perms := users.PermissionsObject{}
	perms.Channel.Add = cliConfirm("Can add channels?")
	perms.Channel.Read = cliConfirm("Can read channels info?")
	perms.User.Add = cliConfirm("Can add new users?")
	perms.User.Remove = cliConfirm("Can remove users?")
	perms.User.EncKeyUpdate = cliConfirm("Can update users encryption key?")
	perms.User.SignKeyUpdate = cliConfirm("Can update users signing key?")
	perms.User.PermissionsUpdate = cliConfirm("Can update users permissions?")

	// Read keys from paths
	encKeyPath := cliGetFilePath("Path to encryption key file:")
	encKeyEncoded, err := GetEncodedPublicKey(encKeyPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	signKeyPath := cliGetFilePath("Path to signing key file:")
	signKeyEncoded, err := GetEncodedPublicKey(signKeyPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create object and write it to stdout
	currentTime := time.Now()
	obj := &users.UserObject{
		Id:          userId,
		EncKey:      encKeyEncoded,
		SignKey:     signKeyEncoded,
		Permissions: perms,
		Active:      true,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}
	WriteUserObject(obj)
}

/*
	Generate user create opreation
*/
func GenerateUserCreateOperation() {
	// Read object from stdin
	obj := ReadUserObject()

	currentTime := time.Now()

	// Generate request
	requestEncoded, _ := users.GenerateCreateRequest(obj, currentTime).Encode()

	// Generate non-encrypted/unsigned operation
	op := core.GenerateOperation(
		false,
		"",
		nil,
		false,
		"",
		nil,
		false,
		"",
		nil,
		false,
		core.UsersRequestType,
		requestEncoded,
		false,
	)

	// Set timestamp
	op.Meta.Timestamp = currentTime

	// Write operation to stdout
	WriteOperation(op)
}
