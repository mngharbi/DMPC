/*
	Integration testing
	(includes all operations)
*/

package channels

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/status"
	"reflect"
	"testing"
	"time"
)

func TestStartShutdownServers(t *testing.T) {
	operationQueuerDummy, _ := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartBothServers(t, multipleWorkersChannelsConfig(), multipleWorkersMessagesConfig(), multipleWorkersListenersConfig(), operationQueuerDummy) {
		return
	}
	ShutdownServers()
}

/*
	Helpers
*/

func makeGenericCloseRequest(channelId string, userId string, timestamp time.Time) *CloseChannelRequest {
	return &CloseChannelRequest{
		Id: channelId,
		Signers: &core.VerifiedSigners{
			IssuerId:    userId,
			CertifierId: userId,
		},
		Timestamp: timestamp,
	}
}

func makeGenericSubscribeRequest(channelId string, userId string) *SubscribeRequest {
	return &SubscribeRequest{
		ChannelId: channelId,
		Signers: &core.VerifiedSigners{
			IssuerId:    userId,
			CertifierId: userId,
		},
	}
}

func makeGenericUnsubscribeRequest(channelId string, subscriberId string) *UnsubscribeRequest {
	return &UnsubscribeRequest{
		ChannelId:    channelId,
		SubscriberId: subscriberId,
	}
}

func makeGenericAddMessageRequest(channelId string, time time.Time, userId string, msg []byte) *AddMessageRequest {
	return &AddMessageRequest{
		ChannelId: channelId,
		Timestamp: time,
		Signers: &core.VerifiedSigners{
			IssuerId:    userId,
			CertifierId: userId,
		},
		Message: msg,
	}
}

func makeAddMessageRequestAndWait(t *testing.T, request *AddMessageRequest) *MessagesResponse {
	channel, err := AddMessage(request)
	if err != nil {
		t.Errorf("Valid add message request should not be rejected. err=%+v", err)
	}
	msg := <-channel
	return msg
}

func makeBufferOperationRequestAndWait(t *testing.T, request *BufferOperationRequest) *MessagesResponse {
	channel, err := BufferOperation(request)
	if err != nil {
		t.Errorf("Valid buffer operation request should not be rejected. err=%+v", err)
	}
	msg := <-channel
	return msg
}

func makeListenersRequestAndWait(t *testing.T, request interface{}) *ListenersResponse {
	channel, err := ListenerAction(request)
	if err != nil {
		t.Errorf("Valid listeners action request should not be rejected. err=%+v", err)
	}
	msg := <-channel
	return msg
}

func makeChannelsRequestAndWait(t *testing.T, request interface{}) *ChannelsResponse {
	channel, err := ChannelAction(request)
	if err != nil {
		t.Errorf("Valid channels action request should not be rejected. err=%+v", err)
	}
	msg := <-channel
	return msg
}

func TestSync(t *testing.T) {
	openingTime := time.Now()
	beforeOpeningTime := openingTime.Add(-1 * time.Hour)
	secondAfterOpeningTime := openingTime.Add(time.Second)
	twoSecondsAfterOpeningTime := openingTime.Add(2 * time.Second)
	minuteAfterOpeningTime := openingTime.Add(time.Minute)
	twoMinutesAfterOpeningTime := openingTime.Add(2 * time.Minute)
	hourAfterOpeningTime := openingTime.Add(time.Hour)
	twoHoursAfterOpeningTime := openingTime.Add(2 * time.Hour)
	threeHoursAfterOpeningTime := openingTime.Add(3 * time.Hour)
	genericChannelId := "CHANNEL_ID"
	inexistentChannelId := "NOT_FOUNT_CHANNEL_ID"
	genericKeyId := "CHANNEL_KEY_ID"
	genericNoopId := "NOOP_USER_ID"
	genericReaderId := "READER_USER_ID"
	genericWriterId := "WRITER_USER_ID"
	genericCloserId := "CLOSER_USER_ID"
	genericMessages := [][]byte{
		[]byte("message_0"),
		[]byte("message_1"),
		[]byte("message_2"),
		[]byte("message_3"),
		[]byte("message_4"),
		[]byte("message_5"),
	}

	// Early invalid time closure
	earlyInvalidCloseReq := makeGenericCloseRequest(genericChannelId, genericCloserId, beforeOpeningTime)

	// Early valid closure
	earlyValidCloseReq := makeGenericCloseRequest(genericChannelId, genericCloserId, threeHoursAfterOpeningTime)

	// Early (earlier) valid closure
	earlierValidCloseReq := makeGenericCloseRequest(genericChannelId, genericCloserId, twoHoursAfterOpeningTime)

	// Early operation
	earlyOp := &core.Operation{
		Meta: core.OperationMetaFields{
			ChannelId: genericChannelId,
			Timestamp: beforeOpeningTime,
		},
	}
	earlyOpReq := &BufferOperationRequest{
		Operation: earlyOp,
	}

	// Early valid subscribe request
	earlyValidSubReq := makeGenericSubscribeRequest(genericChannelId, genericReaderId)

	// Early unauthorized subscribe request
	earlyUnauthorizedSubReq := makeGenericSubscribeRequest(genericChannelId, genericWriterId)

	// Early add message
	earlyAddMessageReq := makeGenericAddMessageRequest(genericChannelId, minuteAfterOpeningTime, genericWriterId, genericMessages[0])

	// Opening request
	openReq := &OpenChannelRequest{
		Id: genericChannelId,
		Signers: &core.VerifiedSigners{
			IssuerId:    genericNoopId,
			CertifierId: genericNoopId,
		},
		KeyId: genericKeyId,
		Key:   generateRandomBytes(core.SymmetricKeySize),
		Permissions: &ChannelPermissionsObject{
			Users: map[string]*ChannelPermissionObject{
				genericNoopId: {
					Read:  false,
					Write: false,
					Close: false,
				},
				genericReaderId: {
					Read:  true,
					Write: false,
					Close: false,
				},
				genericWriterId: {
					Read:  false,
					Write: true,
					Close: false,
				},
				genericCloserId: {
					Read:  false,
					Write: false,
					Close: true,
				},
			},
		},
		Timestamp: openingTime,
	}

	// Second valid closure attempt
	afterOpeningValidCloseReq := makeGenericCloseRequest(genericChannelId, genericCloserId, hourAfterOpeningTime)

	// Unauthorized closure attempt
	afterOpeningUnauthorizedCloseReq := makeGenericCloseRequest(genericChannelId, genericNoopId, twoSecondsAfterOpeningTime)

	// Late operation request
	lateOp := &core.Operation{
		Meta: core.OperationMetaFields{
			ChannelId: genericChannelId,
			Timestamp: minuteAfterOpeningTime,
		},
	}
	lateOpReq := &BufferOperationRequest{
		Operation: lateOp,
	}

	// After opening, early message
	afterOpeningEarlyAddMessageReq := makeGenericAddMessageRequest(genericChannelId, beforeOpeningTime, genericWriterId, genericMessages[1])

	// After opening, good range message, pos 1
	afterOpeningPos1AddMessageReq := makeGenericAddMessageRequest(genericChannelId, minuteAfterOpeningTime, genericWriterId, genericMessages[2])

	// After opening, good range message, pos 2
	afterOpeningPos2AddMessageReq := makeGenericAddMessageRequest(genericChannelId, twoMinutesAfterOpeningTime, genericWriterId, genericMessages[3])

	// After opening, late message
	afterOpeningLateAddMessageReq := makeGenericAddMessageRequest(genericChannelId, twoHoursAfterOpeningTime, genericWriterId, genericMessages[4])

	// Early valid subscribe request
	afterClosureValidSubReq := makeGenericSubscribeRequest(genericChannelId, genericReaderId)

	// Early unauthorized subscribe request
	afterClosureUnauthorizedSubReq := makeGenericSubscribeRequest(genericChannelId, genericWriterId)

	// After opening, good range message, pos 0
	afterOpeningPos0AddMessageReq := makeGenericAddMessageRequest(genericChannelId, secondAfterOpeningTime, genericWriterId, genericMessages[5])

	// Third closure attempt
	thirdCloseReq := makeGenericCloseRequest(genericChannelId, genericCloserId, twoSecondsAfterOpeningTime)

	/*
		Test in order
	*/

	operationQueuerDummy, operationsQueued := createDummyOperationQueuerFunctor(status.RequestNewTicket(), nil, false)
	if !resetAndStartBothServers(t, multipleWorkersChannelsConfig(), multipleWorkersMessagesConfig(), multipleWorkersListenersConfig(), operationQueuerDummy) {
		return
	}

	// Early invalid time closure
	earlyInvalidCloseResp := makeChannelsRequestAndWait(t, earlyInvalidCloseReq)
	if earlyInvalidCloseResp.Result != ChannelsSuccess {
		t.Errorf("Early invalid close request should succeed. response=%+v", earlyInvalidCloseResp)
	}

	// Early valid closure
	earlyValidCloseResp := makeChannelsRequestAndWait(t, earlyValidCloseReq)
	if earlyValidCloseResp.Result != ChannelsSuccess {
		t.Errorf("Early valid close request should succeed. response=%+v", earlyValidCloseResp)
	}

	// Early (earlier) valid closure
	earlierValidCloseResp := makeChannelsRequestAndWait(t, earlierValidCloseReq)
	if earlierValidCloseResp.Result != ChannelsSuccess {
		t.Errorf("Earlier close request should succeed. response=%+v", earlierValidCloseResp)
	}

	// Early operation
	earlyOpResp := makeBufferOperationRequestAndWait(t, earlyOpReq)
	if earlyOpResp.Result != MessagesSuccess {
		t.Errorf("Early buffer operation should succeed. response=%+v", earlyOpResp)
	}

	// Early valid subscribe request
	earlyValidSubResp := makeListenersRequestAndWait(t, earlyValidSubReq)
	if earlyValidSubResp.Result != ListenersSuccess {
		t.Errorf("Early subscribe request should succeed. response=%+v", earlyValidSubResp)
	}

	// Early unauthorized subscribe request
	earlyUnauthorizedSubResp := makeListenersRequestAndWait(t, earlyUnauthorizedSubReq)
	if earlyUnauthorizedSubResp.Result != ListenersSuccess {
		t.Errorf("Early invalid subscribe request should succeed. response=%+v", earlyUnauthorizedSubResp)
	}

	// Early add message
	earlyAddMessageResp := makeAddMessageRequestAndWait(t, earlyAddMessageReq)
	if earlyAddMessageResp.Result != MessagesDropped {
		t.Errorf("Early add messages request should be dropped. response=%+v", earlyAddMessageResp)
	}

	// Expect operation to not be buffered yet
	select {
	case <-operationsQueued:
		t.Errorf("Operation should not be buffered early.")
	default:
		break
	}

	// Opening request
	openResp := makeChannelsRequestAndWait(t, openReq)
	if openResp.Result != ChannelsSuccess {
		t.Errorf("Opening request should succeed. response=%+v", openResp)
	}

	// Expect operation to be buffered
	select {
	case <-operationsQueued:
		break
	default:
		t.Errorf("Operation should be buffered after opening.")
	}

	// Expect unauthorized subscriber channel to be closed
	if _, ok := <-earlyUnauthorizedSubResp.Channel; ok {
		t.Errorf("Early unauthorized subscriber should have channel closed after opening.")
	}

	// Second valid closure attempt
	afterOpeningValidCloseResp := makeChannelsRequestAndWait(t, afterOpeningValidCloseReq)
	if afterOpeningValidCloseResp.Result != ChannelsSuccess {
		t.Errorf("Closing again after opening should succeed. response=%+v", afterOpeningValidCloseResp)
	}

	// Unauthorized closure attempt
	afterOpeningUnauthorizedCloseResp := makeChannelsRequestAndWait(t, afterOpeningUnauthorizedCloseReq)
	if afterOpeningUnauthorizedCloseResp.Result != ChannelsFailure {
		t.Errorf("Closing with unauthorized should fail. response=%+v", afterOpeningUnauthorizedCloseResp)
	}

	// Late operation request, expect it to be buffered
	select {
	case <-operationsQueued:
		t.Errorf("No more operation should be buffered.")
	default:
		break
	}
	lateOpResp := makeBufferOperationRequestAndWait(t, lateOpReq)
	if lateOpResp.Result != MessagesSuccess {
		t.Errorf("Late operation buffering after opening should succeed. response=%+v", lateOpResp)
	}
	select {
	case <-operationsQueued:
		break
	default:
		t.Errorf("Operation requests should be run after opening.")
	}

	// After opening, early message
	afterOpeningEarlyAddMessageResp := makeAddMessageRequestAndWait(t, afterOpeningEarlyAddMessageReq)
	if afterOpeningEarlyAddMessageResp.Result != MessagesDropped {
		t.Errorf("Early message after opening should be dropped. response=%+v", afterOpeningEarlyAddMessageResp)
	}

	// After opening, good range message, pos 1
	afterOpeningPos1AddMessageResp := makeAddMessageRequestAndWait(t, afterOpeningPos1AddMessageReq)
	if afterOpeningPos1AddMessageResp.Result != MessagesSuccess {
		t.Errorf("Position 1 message after opening should succeed. response=%+v", afterOpeningPos1AddMessageResp)
	}

	// After opening, good range message, pos 2
	afterOpeningPos2AddMessageResp := makeAddMessageRequestAndWait(t, afterOpeningPos2AddMessageReq)
	if afterOpeningPos2AddMessageResp.Result != MessagesSuccess {
		t.Errorf("Position 2 message after opening should succeed. response=%+v", afterOpeningPos2AddMessageResp)
	}

	// After opening, late message
	afterOpeningLateAddMessageResp := makeAddMessageRequestAndWait(t, afterOpeningLateAddMessageReq)
	if afterOpeningLateAddMessageResp.Result != MessagesDropped {
		t.Errorf("After closure timestamped message after opening should be dropped. response=%+v", afterOpeningLateAddMessageResp)
	}

	// Expect 5 events to be read in early subscriber: opening, early close, after opening closing, 2 messages
	eventsExpected := []*Event{
		makeOpenEvent(openingTime),
		makeCloseEvent(twoHoursAfterOpeningTime, 0),
		makeCloseEvent(hourAfterOpeningTime, 0),
		makeMessageEvent(minuteAfterOpeningTime, 0, genericMessages[2]),
		makeMessageEvent(twoMinutesAfterOpeningTime, 1, genericMessages[3]),
	}
	for i := 0; i < len(eventsExpected); i++ {
		event := <-earlyValidSubResp.Channel
		if !reflect.DeepEqual(event, eventsExpected[i]) {
			t.Errorf("Early subscriber event #%v does not match. event=%+v, expected=%+v", i, event, eventsExpected[i])
		}
	}
	select {
	case <-earlyValidSubResp.Channel:
		t.Errorf("Early subscriber should only read 5 events.")
	default:
		break
	}

	// Wait until events are drained
	waitUntilEmpty(genericChannelId)

	// Valid subscribe request
	afterClosureValidSubResp := makeListenersRequestAndWait(t, afterClosureValidSubReq)
	if afterClosureValidSubResp.Result != ListenersSuccess {
		t.Errorf("Valid subscribe request after closure should succeed. response=%+v", afterClosureValidSubResp)
	}

	// Early unauthorized subscribe request
	afterClosureUnauthorizedSubResp := makeListenersRequestAndWait(t, afterClosureUnauthorizedSubReq)
	if afterClosureUnauthorizedSubResp.Result != ListenersUnauthorized ||
		afterClosureUnauthorizedSubResp.Channel != nil {
		t.Errorf("Unauthorized subscribe request after closure should fail. response=%+v", afterClosureUnauthorizedSubResp)
	}

	// After opening, good range message, pos 0
	afterOpeningPos0AddMessageResp := makeAddMessageRequestAndWait(t, afterOpeningPos0AddMessageReq)
	if afterOpeningPos0AddMessageResp.Result != MessagesSuccess {
		t.Errorf("Position 0 message after opening should succeed. response=%+v", afterOpeningPos0AddMessageResp)
	}

	// Expect to only send one message event in both subscribers
	expectedPos0MessageEvent := makeMessageEvent(secondAfterOpeningTime, 0, genericMessages[5])
	subscriberChannels := []EventChannel{earlyValidSubResp.Channel, afterClosureValidSubResp.Channel}
	for subscriberIdx, subscriberChannel := range subscriberChannels {
		event, ok := <-subscriberChannel
		if !ok || !reflect.DeepEqual(event, expectedPos0MessageEvent) {
			t.Errorf("After opening pos 0 message should be read by subscriber. subscriberIdx=%v, event=%+v, expected=%+v", subscriberIdx, event, expectedPos0MessageEvent)
		}
		select {
		case <-subscriberChannel:
			t.Errorf("Subscriber should only get one message event.")
		default:
			break
		}
	}

	// Third late closure
	thirdCloseResp := makeChannelsRequestAndWait(t, thirdCloseReq)
	if thirdCloseResp.Result != ChannelsSuccess {
		t.Errorf("Third closure attempt should succeed. response=%+v", thirdCloseResp)
	}

	// Expect to only send a close event in both subscribers
	expectedThirdCloseEvent := makeCloseEvent(twoSecondsAfterOpeningTime, 1)
	for subscriberIdx, subscriberChannel := range subscriberChannels {
		event, ok := <-subscriberChannel
		if !ok || !reflect.DeepEqual(event, expectedThirdCloseEvent) {
			t.Errorf("Subscribers need to be notified of third closure. subscriberIdx=%v, event=%+v, expected=%+v", subscriberIdx, event, expectedThirdCloseEvent)
		}
		select {
		case <-subscriberChannel:
			t.Errorf("Subscriber should only get one close event.")
		default:
			break
		}
	}

	// Unsubscribe request to automatically unsubscribed
	invalidIdUnsubscribeReq := makeGenericUnsubscribeRequest(genericChannelId, earlyUnauthorizedSubResp.SubscriberId)
	invalidIdUnsubscribeResp := makeListenersRequestAndWait(t, invalidIdUnsubscribeReq)
	if invalidIdUnsubscribeResp.Result != ListenersFailure {
		t.Errorf("Unsubscribing unsubscribed id should fail. response=%+v", invalidIdUnsubscribeResp)
	}

	// Unsubscribe from invalid channel
	invalidChannelUnsubscribeReq := makeGenericUnsubscribeRequest(inexistentChannelId, earlyValidSubResp.SubscriberId)
	invalidChannelUnsubscribeResp := makeListenersRequestAndWait(t, invalidChannelUnsubscribeReq)
	if invalidChannelUnsubscribeResp.Result != ListenersFailure {
		t.Errorf("Unsubscribing from unknown channel should fail. response=%+v", invalidChannelUnsubscribeResp)
	}

	// Unsubscribe both channels
	subscriberIds := []string{earlyValidSubResp.SubscriberId, afterClosureValidSubResp.SubscriberId}
	for _, subscriberId := range subscriberIds {
		validUnsubscribeReq := makeGenericUnsubscribeRequest(genericChannelId, subscriberId)
		validUnsubscribeResp := makeListenersRequestAndWait(t, validUnsubscribeReq)
		if validUnsubscribeResp.Result != ListenersSuccess {
			t.Errorf("Unsubscribing for valid subscriber id should succeed. response=%+v", validUnsubscribeResp)
		}
	}
	for _, subscriberChannel := range subscriberChannels {
		if event, ok := <-subscriberChannel; ok {
			t.Errorf("Subscriber channels should be closed after unsubscribing. Read event=%+v", event)
		}
	}

	ShutdownServers()
}
