/*
	Test helpers
*/

package channels

import (
	"crypto/rand"
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/memstore"
	"sync"
	"testing"
	"time"
)

/*
	Shared Constants and variables
*/

const (
	genericChannelId   string = "channelId"
	genericKeyId       string = "keyId"
	genericUserId      string = "USER_1"
	genericIssuerId    string = "IssuerId"
	genericCertifierId string = "CertifierId"
)

/*
	Utilities for channel state
*/

func (rec *channelRecord) isDurationSet() bool {
	return rec.duration != nil
}

func (rec *channelRecord) isOpenedSet() bool {
	return rec.isDurationSet() && !rec.duration.opened.IsZero()
}

func (rec *channelRecord) isClosedSet() bool {
	return rec.isDurationSet() && !rec.duration.closed.IsZero()
}

func (rec *channelRecord) isValidDuration() bool {
	return rec.isDurationSet() && rec.duration.closed.After(rec.duration.opened)
}

func (rec *channelRecord) hasClosureAttempts() bool {
	return len(rec.closureAttempts) > 0
}

func (rec *channelRecord) hasClosureRecord() bool {
	return rec.closure != nil
}

func (rec *channelRecord) computeState() channelState {
	if !rec.isOpenedSet() && !rec.isClosedSet() && !rec.hasClosureRecord() {
		return channelBufferedState
	}
	if rec.isOpenedSet() && !rec.hasClosureAttempts() && !rec.isClosedSet() && !rec.hasClosureRecord() {
		return channelOpenState
	}
	if rec.isOpenedSet() && !rec.hasClosureAttempts() && rec.isClosedSet() && rec.hasClosureRecord() && rec.isValidDuration() {
		return channelClosedState
	}

	return channelInconsistentState
}

/*
	Generic utilities
*/

func generateRandomBytes(nbBytes int) (bytes []byte) {
	bytes = make([]byte, nbBytes)
	rand.Read(bytes)
	return
}

/*
	Channels server utilities
*/

func startChannelsServerAndTest(t *testing.T, conf ChannelsServerConfig) bool {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	if err := startChannelsServer(conf, wg); err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func resetAndStartChannelsServer(t *testing.T, conf ChannelsServerConfig) bool {
	channelsServerSingleton = channelsServer{}
	return startChannelsServerAndTest(t, conf)
}

func multipleWorkersChannelsConfig() ChannelsServerConfig {
	return ChannelsServerConfig{
		NumWorkers: 6,
	}
}

/*
	Messages server utilities
*/

func startMessagesServerAndTest(t *testing.T, conf MessagesServerConfig) bool {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	if err := startMessagesServer(conf, wg); err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func resetAndStartMessagesServer(t *testing.T, conf MessagesServerConfig) bool {
	messagesServerSingleton = messagesServer{}
	return startMessagesServerAndTest(t, conf)
}

func multipleWorkersMessagesConfig() MessagesServerConfig {
	return MessagesServerConfig{
		NumWorkers: 6,
	}
}

func makeValidAddMessageRequest() *AddMessageRequest {
	return &AddMessageRequest{
		Timestamp: time.Now(),
		Signers: &core.VerifiedSigners{
			IssuerId:    genericIssuerId,
			CertifierId: genericCertifierId,
		},
		Message: generateRandomBytes(100),
	}
}

func makeValidBufferOperationRequest() *BufferOperationRequest {
	return &BufferOperationRequest{
		Operation: core.GenerateOperation(
			true,
			genericKeyId,
			nil,
			true,
			genericIssuerId,
			nil,
			true,
			genericCertifierId,
			nil,
			true,
			core.AddMessageType,
			generateRandomBytes(100),
			false,
		),
	}
}

/*
	Listeners server utilities
*/

func startListenersServerAndTest(t *testing.T, conf ListenersServerConfig) bool {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	if err := startListenersServer(conf, wg); err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func resetAndStartListenersServer(t *testing.T, conf ListenersServerConfig) bool {
	listenersServerSingleton = listenersServer{}
	return startListenersServerAndTest(t, conf)
}

func multipleWorkersListenersConfig() ListenersServerConfig {
	return ListenersServerConfig{
		NumWorkers: 6,
	}
}

func getListenersRecordById(mem *memstore.Memstore, id string) *listenersRecord {
	item := mem.Get(makeSearchListenersRecord(id), listenersIndexId)
	if item != nil {
		return item.(*listenersRecord)
	}
	return nil
}

/*
	Server utilities
*/

func startBothServersAndTest(t *testing.T, channelsConf ChannelsServerConfig, messagesConf MessagesServerConfig, listenersConf ListenersServerConfig) bool {
	if err := StartServers(channelsConf, messagesConf, listenersConf, log, shutdownProgram); err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func resetAndStartBothServers(t *testing.T, channelsConf ChannelsServerConfig, messagesConf MessagesServerConfig, listenersConf ListenersServerConfig) bool {
	channelsServerSingleton = channelsServer{}
	messagesServerSingleton = messagesServer{}
	listenersServerSingleton = listenersServer{}
	return startBothServersAndTest(t, channelsConf, messagesConf, listenersConf)
}
