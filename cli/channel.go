package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mngharbi/DMPC/channels"
	"github.com/mngharbi/DMPC/core"
	"log"
	"os"
	"time"
)

/*
	Read channel object from stdin
*/
func ReadChannelObject() (obj *channels.ChannelObject) {
	obj = &channels.ChannelObject{}
	err := json.NewDecoder(os.Stdin).Decode(obj)
	if err != nil {
		log.Fatal(err.Error())
	}
	return
}

/*
	Write channel object to stdout
*/
func WriteChannelObject(obj *channels.ChannelObject) {
	json.NewEncoder(os.Stdout).Encode(obj)
}

/*
	Generate channel object
*/
func GenerateChannelObject() {
	ch := &channels.ChannelObject{}

	// Make channel id
	channelPrefix := cliGetString("Enter the channel prefix:")
	ch.Id = fmt.Sprintf("%v_%v", channelPrefix, core.GenerateUniqueId())

	// Read users and their permissions
	ch.Permissions.Users = map[string]channels.ChannelPermissionObject{}
	for {
		userId := cliGetString("Enter channel member's id:")
		ch.Permissions.Users[userId] = channels.ChannelPermissionObject{
			Read:  cliConfirm("Read permission?"),
			Write: cliConfirm("Write permission?"),
			Close: cliConfirm("Close permission?"),
		}
		if !cliConfirm("Add another member?") {
			break
		}
	}

	// Make channel key id
	ch.KeyId = core.GenerateUniqueId()

	ch.State = channels.ChannelObjectOpenState

	WriteChannelObject(ch)
}

/*
	Generate channel open operation
*/
func GenerateChannelOpenOperation() {
	ch := ReadChannelObject()

	currentTime := time.Now()

	// Make request
	rq := &channels.OpenChannelRequest{
		Channel:   ch,
		Key:       core.GenerateSymmetricKey(),
		Timestamp: currentTime,
	}
	rqEncoded, _ := rq.Encode()

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
		core.AddChannelType,
		rqEncoded,
		false,
	)

	// Set Channel from channel object
	op.Meta.ChannelId = ch.Id

	// Set timestamp
	op.Meta.Timestamp = currentTime

	// Write operation to stdout
	WriteOperation(op)
}
