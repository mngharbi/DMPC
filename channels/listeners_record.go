package channels

import (
	"errors"
	"github.com/mngharbi/eventqueue"
	"sync"
)

/*
	Errors
*/

var (
	unrecognizedChannelId error = errors.New("Unrecognized channel id")
)

/*
	Definitions
*/

var (
	listenersStore *sync.Map
)

type listenersRecord struct {
	eventQueue        *eventqueue.EventQueue
	listenerCertifier map[string]string
}

// Alias for a channel of events
type EventChannel chan *Event

/*
	Pubsub operations
*/

func getOrMakeListenersRecord(channelId string) *listenersRecord {
	listenersRecInterface, _ := listenersStore.LoadOrStore(channelId, &listenersRecord{
		eventQueue:        eventqueue.New(),
		listenerCertifier: make(map[string]string),
	})
	return listenersRecInterface.(*listenersRecord)
}

func subscribe(channelId string, certifierId string) (EventChannel, string) {
	// Make channel to be passed to daemon and back to the caller
	channel := make(EventChannel, 0)

	// Get or make listeners record
	listenersRec := getOrMakeListenersRecord(channelId)

	// Subscribe and track subscriber certifier id
	genericChannel, subscriberId := listenersRec.eventQueue.Subscribe()
	listenersRec.listenerCertifier[subscriberId] = certifierId

	// Pass through result
	go func() {
		for event := range genericChannel {
			channel <- event.(*Event)
		}
		close(channel)
	}()

	return channel, subscriberId
}

func unsubscribe(channelId string, subscriberId string) error {
	// Get listeners record
	listenersRecInterface, ok := listenersStore.Load(channelId)
	if !ok {
		return unrecognizedChannelId
	}
	listenersRec := listenersRecInterface.(*listenersRecord)

	// Unsubscribe and delete certifier
	delete(listenersRec.listenerCertifier, subscriberId)
	return listenersRec.eventQueue.Unsubscribe(subscriberId)
}

func unsubscribeUnauthorized(channelId string, permissions *channelPermissionsRecord) {
	// Get or make listeners record
	listenersRec := getOrMakeListenersRecord(channelId)

	// Search for unauthorized subscribers
	unauthorizedSubscriberIds := []string{}
	for subscriberId, certifierId := range listenersRec.listenerCertifier {
		certifierPermissions, certifierFound := permissions.users[certifierId]
		if !certifierFound || !certifierPermissions.read {
			unauthorizedSubscriberIds = append(unauthorizedSubscriberIds, subscriberId)
		}
	}

	for _, subscriberId := range unauthorizedSubscriberIds {
		delete(listenersRec.listenerCertifier, subscriberId)
		listenersRec.eventQueue.Unsubscribe(subscriberId)
	}
}

func publish(id string, event *Event) error {
	channelEventQueue, _ := listenersStore.Load(id)
	return channelEventQueue.(*eventqueue.EventQueue).Publish(event)
}
