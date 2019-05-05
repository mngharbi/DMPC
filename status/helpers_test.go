/*
	Test helpers
*/

package status

import (
	"encoding/json"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func waitForRandomDuration() {
	duration := time.Duration(rand.Intn(100)) * time.Millisecond
	timer := time.NewTimer(duration)
	<-timer.C
}

func shuffleStatusRecords(src []*StatusRecord) {
	for i := range src {
		j := rand.Intn(i + 1)
		src[i], src[j] = src[j], src[i]
	}
}

func startStatusServerAndTest(t *testing.T, conf StatusServerConfig) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	serversStartWaitGroup.Add(1)
	if err := startStatusServer(conf); err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func resetAndStartStatusServer(t *testing.T, conf StatusServerConfig) bool {
	statusServerSingleton = statusServer{}
	return startStatusServerAndTest(t, conf)
}

func multipleWorkersStatusConfig() StatusServerConfig {
	return StatusServerConfig{
		NumWorkers: 6,
	}
}

func startListenersServerAndTest(t *testing.T, conf ListenersServerConfig) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	serversStartWaitGroup.Add(1)
	if err := startListenersServer(conf); err != nil {
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

func startBothServersAndTest(t *testing.T, statusConf StatusServerConfig, listenersConf ListenersServerConfig, ignoreError bool) bool {
	serversStartWaitGroup = sync.WaitGroup{}
	if err := StartServers(statusConf, listenersConf, log, shutdownProgram); err != nil {
		if !ignoreError {
			t.Errorf(err.Error())
		}
		return false
	}
	return true
}

func resetAndStartBothServers(t *testing.T, statusConf StatusServerConfig, listenersConf ListenersServerConfig, ignoreError bool) bool {
	statusServerSingleton = statusServer{}
	listenersServerSingleton = listenersServer{}
	return startBothServersAndTest(t, statusConf, listenersConf, ignoreError)
}

/*
	Encodable type
*/

type encodableTestStruct struct {
	Id int `json:"id"`
}

func (rec *encodableTestStruct) Encode() ([]byte, error) {
	return json.Marshal(rec)
}

/*
	Channel response type
*/

type channelTestStruct struct {
	Channel      chan []byte
	ChannelId    string
	SubscriberId string
}

func (ch *channelTestStruct) GetResponse() ([]byte, bool) {
	resp, ok := <-ch.Channel
	return resp, ok
}

func (ch *channelTestStruct) GetSubscriberId() string {
	return ch.SubscriberId
}

func (ch *channelTestStruct) GetChannelId() string {
	return ch.ChannelId
}
