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
	Order     int       `json:"order"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

func makeOpenEvent(timestamp time.Time) *Event {
	return &Event{
		Type:      Open,
		Order:     0,
		Timestamp: timestamp,
		Data:      nil,
	}
}

func makeCloseEvent(timestamp time.Time, validMessages int) *Event {
	return &Event{
		Type:      Close,
		Order:     validMessages,
		Timestamp: timestamp,
		Data:      nil,
	}
}

func makeMessageEvent(timestamp time.Time, order int, message []byte) *Event {
	return &Event{
		Type:      Message,
		Order:     order,
		Timestamp: timestamp,
		Data:      message,
	}
}
