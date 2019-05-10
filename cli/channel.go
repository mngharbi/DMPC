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
	Encrypt channel operation
*/
func EncryptChannelOperation(op *core.Operation) {
	opEncoded, _ := op.Encode()
	signedEncyrptionOperation := WrapPayloadInRootSignedGenericOperation(opEncoded, core.ChannelEncryptType)
	signedEncyrptionOperation.Meta.ChannelId = op.Meta.ChannelId
	encryptionTs := WrapOperationInResultOnlyTransaction(signedEncyrptionOperation)
	encryptionTsEncoded, _ := encryptionTs.Encode()

	// Run transaction and write output
	runOneTransactionAndWrite(encryptionTsEncoded)
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
func GenerateChannelOpenOperation(issue bool, certify bool) {
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

	// Sign operation
	if issue || certify {
		RootSignOperation(op, issue, certify)
	}

	// Write operation to stdout
	WriteOperation(op)
}

/*
	Generic channel action (except open)
*/
func generateGenericChannelCloseOperation(channelId string, issue bool, certify bool, encrypt bool, rqEncoded []byte, requestType core.RequestType, currentTime time.Time) {
	// Read channel id if none passed
	if channelId == "" {
		channelId = cliGetString("Enter channel id:")
	}

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
		requestType,
		rqEncoded,
		false,
	)

	// Set Channel from channel object
	op.Meta.ChannelId = channelId

	// Set timestamp
	op.Meta.Timestamp = currentTime

	// Sign operation
	if issue || certify {
		RootSignOperation(op, issue, certify)
	}

	if encrypt {
		EncryptChannelOperation(op)
	} else {
		WriteOperation(op)
	}
}

/*
	Generate channel close operation
*/
func GenerateChannelCloseOperation(channelId string, issue bool, certify bool, encrypt bool) {
	// Make request
	currentTime := time.Now()
	rq := &channels.CloseChannelRequest{
		Timestamp: currentTime,
	}
	rqEncoded, _ := rq.Encode()

	generateGenericChannelCloseOperation(channelId, issue, certify, encrypt, rqEncoded, core.CloseChannelType, currentTime)
}
