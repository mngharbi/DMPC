package cli

import (
	"encoding/json"
	"github.com/mngharbi/DMPC/core"
	"log"
	"os"
	"time"
)

/*
	Read transaction from stdin
*/
func ReadTransaction() (ts *core.Transaction) {
	ts = &core.Transaction{}
	err := json.NewDecoder(os.Stdin).Decode(ts)
	if err != nil {
		log.Fatal(err.Error())
	}
	return
}

/*
	Write transaction to stdout
*/
func WriteTransaction(ts *core.Transaction) {
	json.NewEncoder(os.Stdout).Encode(ts)
}

/*
	Generators
*/

func ResultOnlyPlaintextTransaction(payload []byte) *core.Transaction {
	return &core.Transaction{
		Version: core.Version,
		Encryption: core.TransactionEncryptionFields{
			Encrypted:  false,
			Challenges: nil,
			Nonce:      "",
		},
		Pipeline: core.PipelineConfig{
			ReadStatusUpdates: false,
			ReadResult:        true,
			KeepAlive:         false,
		},
		Payload: core.PlaintextEncode(payload),
	}
}

func WrapOperationInResultOnlyTransaction(op *core.Operation) *core.Transaction {
	opEncoded, _ := op.Encode()
	return ResultOnlyPlaintextTransaction(opEncoded)
}

func WrapPayloadInGenericOperation(payload []byte, requestType core.RequestType) *core.Operation {
	return &core.Operation{
		Encryption: core.OperationEncryptionFields{
			Encrypted: false,
			KeyId:     "",
			Nonce:     "",
		},
		Issue: core.OperationAuthenticationFields{
			Id:        "",
			Signature: "",
		},
		Certification: core.OperationAuthenticationFields{
			Id:        "",
			Signature: "",
		},
		Meta: core.OperationMetaFields{
			RequestType: requestType,
			Timestamp:   time.Now(),
		},
		Payload: core.PlaintextEncode(payload),
	}
}

func WrapPayloadInRootSignedGenericOperation(payload []byte, requestType core.RequestType) *core.Operation {
	op := WrapPayloadInGenericOperation(payload, requestType)
	RootSignOperation(op, true, true)
	return op
}

/*
	Wrap into transaction
*/
func GenerateTransaction(ignoreResult bool, statusUpdates bool, keepAlive bool, recepients []string) {
	op := ReadOperation()
	opEncoded, _ := op.Encode()

	// Create wrapping non-encrypted transaction
	ts := &core.Transaction{
		Version: core.Version,
		Encryption: core.TransactionEncryptionFields{
			Encrypted:  false,
			Challenges: nil,
			Nonce:      "",
		},
		Pipeline: core.PipelineConfig{
			ReadStatusUpdates: statusUpdates,
			ReadResult:        !ignoreResult,
			KeepAlive:         keepAlive,
		},
		Payload: core.PlaintextEncode(opEncoded),
	}

	// Encrypt for recepients if any
	if len(recepients) > 0 {
		// Set up challenges map for recepients
		ts.Encryption.Challenges = make(map[string]string)
		for _, recepient := range recepients {
			ts.Encryption.Challenges[recepient] = ""
		}

		// Prepare transaction encrypt request transaction
		tsEncoded, _ := ts.Encode()
		encryptionTs := WrapOperationInResultOnlyTransaction(
			WrapPayloadInRootSignedGenericOperation(tsEncoded, core.TransactionEncryptType),
		)
		encryptionTsEncoded, _ := encryptionTs.Encode()

		// Run transaction and write output
		runOneTransactionAndWrite(encryptionTsEncoded)
	} else {
		WriteTransaction(ts)
	}
}

/*
	Reads transaction from stdin
	Writes results into stdout until the connection closes
*/
func ReadAndRunOneTransaction() {
	ts := ReadTransaction()
	tsEncoded, _ := ts.Encode()

	runOneTransactionAndWrite(tsEncoded)
}
