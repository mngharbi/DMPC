package channels

/*
	Logging messages
*/
const (
	// Channels daemon
	channelsDaemonStartLogMsg    string = "Channels daemon started"
	channelsDaemonShutdownLogMsg string = "Channels daemon shutdown"
	channelsRunningRequestLogMsg string = "Channels running request"
	channelsRequestDoneLogMsg    string = "Channels request done"

	// Messages daemon
	messagesDaemonStartLogMsg    string = "Channel messages daemon started"
	messagesDaemonShutdownLogMsg string = "Channel messages daemon shutdown"
	messagesRunningRequestLogMsg string = "Channel messages running request"
	messagesRequestDoneLogMsg    string = "Channel messages request done"

	// Listeners daemon
	listenersDaemonStartLogMsg    string = "Channel listeners daemon started"
	listenersDaemonShutdownLogMsg string = "Channel listeners daemon shutdown"
	listenersRunningRequestLogMsg string = "Channel listeners running request"
	listenersRequestDoneLogMsg    string = "Channel listeners request done"
)
