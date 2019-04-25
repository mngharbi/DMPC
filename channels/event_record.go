package channels

import (
	"time"
)

/*
	Event types
*/

type EventType string

const (
	Open    EventType = "channel_open"
	Message EventType = "new_message"
	Close   EventType = "channel_close"
)

/*
	Event definition
*/

type Event struct {
	Type      EventType `json:"type"`
	Position  int       `json:"position"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

func makeOpenEvent(timestamp time.Time) *Event {
	return &Event{
		Type:      Open,
		Position:  0,
		Timestamp: timestamp,
		Data:      nil,
	}
}

func makeCloseEvent(timestamp time.Time, validMessages int) *Event {
	return &Event{
		Type:      Close,
		Position:  validMessages,
		Timestamp: timestamp,
		Data:      nil,
	}
}

func makeMessageEvent(timestamp time.Time, position int, message []byte) *Event {
	return &Event{
		Type:      Message,
		Position:  position,
		Timestamp: timestamp,
		Data:      message,
	}
}
