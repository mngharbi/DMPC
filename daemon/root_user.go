package daemon

/*
	Utilities for getting the root user from conf
*/

import (
	"github.com/mngharbi/DMPC/config"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/decryptor"
	"github.com/mngharbi/DMPC/executor"
	"github.com/mngharbi/DMPC/status"
	"github.com/mngharbi/DMPC/users"
	"time"
)

/*
   Error messages
*/
const (
	encodeRootUserOperationError string = "Unable to encode root user operation"
	createRootUserRequestError   string = "Error making root user creation request"
	listenOnRootUserRequestError string = "Error setting up listener on root user creation request"
	createRootUserFailedError    string = "User creation request failed"
)

/*
   Utilities
*/
func buildRootUserOperation(conf *config.Config) *core.Transaction {
	// Get root user object from confuration
	log.Debugf("Parsing root user object from confuration")
	rootUserObject := conf.GetRootUserObject()

	// Build user request
	encodedCreateRequest, err := users.GenerateCreateRequest(rootUserObject, time.Now()).Encode()
	if err != nil {
		log.Fatalf(encodeRootUserOperationError)
	}

	// Build transaction
	permanentEncryptedEncoded, err := core.GeneratePermanentEncryptedOperation(
		// Non encrypted
		false, "", nil, true,
		// No issuer
		"", nil, true,
		// No certifier
		"", nil, true,
		// non base64 encoded payload and meta
		executor.UsersRequest, encodedCreateRequest, false,
	).Encode()
	if err != nil {
		log.Fatalf(encodeRootUserOperationError)
	}
	return core.GenerateTransaction(
		// Non encrypted
		false, nil, nil, true,
		// non base64 encoded payload
		permanentEncryptedEncoded, false,
	)
}

func createRootUser(transaction *core.Transaction) {
	// Make unverified request
	log.Debugf("Requesting to add root user")
	rootUserChannel, errs := decryptor.MakeUnverifiedTransactionRequest(transaction)
	if len(errs) != 0 {
		log.Fatalf(createRootUserRequestError)
	}

	// Wait for decryptor to return ticket
	log.Debugf("Root user request made. Waiting for ticket")
	rootUserNativeResp := <-rootUserChannel
	rootUserResp := (*rootUserNativeResp).(*decryptor.DecryptorResponse)
	if rootUserResp.Result != decryptor.Success {
		log.Fatalf(createRootUserRequestError)
	}

	// Wait until ticket status is success
	log.Debugf("Adding listener on user creation ticket")
	updateChannel, err := status.AddListener(rootUserResp.Ticket)
	if err != nil {
		log.Fatalf(listenOnRootUserRequestError)
	}

	log.Debugf("Waiting for user creation to be executed")
	var statusUpdate *status.StatusRecord
	for statusUpdate = range updateChannel {
	}
	if statusUpdate.Status != status.SuccessStatus {
		log.Fatalf(createRootUserFailedError)
	}

	log.Debugf("Root user successfully created")
}
