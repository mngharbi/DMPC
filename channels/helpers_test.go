/*
	Test helpers
*/

package channels

import (
	"sync"
	"testing"
)

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
