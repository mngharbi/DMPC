package channels

import (
	"github.com/mngharbi/DMPC/status"
	"testing"
)

func TestMessagesStartShutdown(t *testing.T) {
	operationQueuerDummy, _ := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartMessagesServer(t, multipleWorkersMessagesConfig(), operationQueuerDummy) {
		return
	}
	shutdownMessagesServer()
}

/*
	Add message
*/

func TestAddMessageServerDown(t *testing.T) {
	if ch, err := AddMessage(makeValidAddMessageRequest()); ch != nil || err == nil {
		t.Error("Adding message while server is down should fail.")
	}
}

func TestAddMessageValid(t *testing.T) {
	operationQueuerDummy, _ := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartMessagesServer(t, multipleWorkersMessagesConfig(), operationQueuerDummy) {
		return
	}
	defer shutdownMessagesServer()
	if ch, err := AddMessage(makeValidAddMessageRequest()); ch == nil || err != nil {
		t.Error("Adding valid message should not fail.")
	}
}

func TestAddMessageInvalid(t *testing.T) {
	operationQueuerDummy, _ := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartMessagesServer(t, multipleWorkersMessagesConfig(), operationQueuerDummy) {
		return
	}
	defer shutdownMessagesServer()

	// Invalid signers
	invalid := makeValidAddMessageRequest()
	invalid.Signers = nil
	ch, err := AddMessage(invalid)
	if ch != nil || err == nil {
		t.Error("Adding message with invalid signers should fail.")
	}

	// Empty message
	invalid = makeValidAddMessageRequest()
	invalid.Message = []byte{}
	ch, err = AddMessage(invalid)
	if ch != nil || err == nil {
		t.Error("Adding empty message should fail.")
	}

	// Empty channel id
	invalid = makeValidAddMessageRequest()
	invalid.ChannelId = ""
	ch, err = AddMessage(invalid)
	if ch != nil || err == nil {
		t.Error("Adding message with empty channel id should fail.")
	}

	// Invalid message
	invalid = makeValidAddMessageRequest()
	invalid.Message = []byte("{}")
	ch, err = AddMessage(invalid)
	if ch != nil || err == nil {
		t.Error("Adding message with invalid message should fail.")
	}
}

/*
	Buffer operation
*/

func TestBufferOperationServerDown(t *testing.T) {
	if ch, err := BufferOperation(makeValidBufferOperationRequest()); ch != nil || err == nil {
		t.Error("Buffer operation while server is down should fail.")
	}
}

func TestBufferOperationValid(t *testing.T) {
	operationQueuerDummy, _ := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartMessagesServer(t, multipleWorkersMessagesConfig(), operationQueuerDummy) {
		return
	}
	defer shutdownMessagesServer()
	if ch, err := BufferOperation(makeValidBufferOperationRequest()); ch == nil || err != nil {
		t.Error("Buffering valid operation should not fail.")
	}
}

func TestBufferOperationInvalid(t *testing.T) {
	operationQueuerDummy, _ := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartMessagesServer(t, multipleWorkersMessagesConfig(), operationQueuerDummy) {
		return
	}
	defer shutdownMessagesServer()

	// Nil operation
	invalid := makeValidBufferOperationRequest()
	invalid.Operation = nil
	ch, err := BufferOperation(invalid)
	if ch != nil || err == nil {
		t.Error("Buffer operation with nil operation should fail.")
	}

	// Mark as buffered
	invalid = makeValidBufferOperationRequest()
	invalid.Operation.Meta.Buffered = false
	ch, err = BufferOperation(invalid)
	if ch == nil || err != nil || invalid.Operation.Meta.Buffered == false {
		t.Error("Buffer operation with not buffered operation should mark it as buffered.")
	}
}
