/*
	Integration test
	for adding listeners and status updates
*/

package status

import (
	"sync"
	"testing"
)

func TestIsolatedListeners(t *testing.T) {
	// Test listeners for different tickets
	numIsolatedTickets := 100
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	for i := 0; i < numIsolatedTickets; i++ {
		_, _ = AddListener(RequestNewTicket())
	}
	ShutdownServers()
	if listenersStore.Len() != numIsolatedTickets || statusStore.Len() != numIsolatedTickets {
		t.Errorf("Creating listeners for different tickets should create different status and listener records")
	}

	// Test listeners for the same ticket
	numListeners := 100
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	ticket := RequestNewTicket()
	for i := 0; i < numListeners; i++ {
		_, _ = AddListener(ticket)
	}
	ShutdownServers()
	if listenersStore.Len() != 1 || statusStore.Len() != 1 {
		t.Errorf("Creating listeners for the same ticket should create one status and listener record")
	}
}

func TestEarlyListeners(t *testing.T) {
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	numListeners := 100
	ticket := RequestNewTicket()
	channels := []UpdateChannel{}
	for i := 0; i < numListeners; i++ {
		channel, _ := AddListener(ticket)
		channels = append(channels, channel)
	}
	ShutdownServers()
	if !startBothServersAndTest(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	expectedStatusUpdate := &StatusRecord{
		Id:         ticket,
		Status:     QueuedStatus,
		FailReason: NoReason,
		Payload:    nil,
		Errs:       nil,
	}
	_ = UpdateStatus(ticket, expectedStatusUpdate.Status, expectedStatusUpdate.FailReason, expectedStatusUpdate.Payload, expectedStatusUpdate.Errs)
	for _, channel := range channels {
		statusUpdate, isOpen := <-channel
		if isOpen && !statusUpdate.isSame(expectedStatusUpdate) {
			t.Errorf("Early listeners should get first update. \n found=%+v\n expected=%+v", statusUpdate, expectedStatusUpdate)
		}
	}
	ShutdownServers()
}

func TestLateListeners(t *testing.T) {
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	ticket := RequestNewTicket()
	expectedStatusUpdate := &StatusRecord{
		Id:         ticket,
		Status:     QueuedStatus,
		FailReason: NoReason,
		Payload:    nil,
		Errs:       nil,
	}
	_ = UpdateStatus(ticket, expectedStatusUpdate.Status, expectedStatusUpdate.FailReason, expectedStatusUpdate.Payload, expectedStatusUpdate.Errs)
	ShutdownServers()
	if !startBothServersAndTest(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	numListeners := 100
	channels := []UpdateChannel{}
	for i := 0; i < numListeners; i++ {
		channel, _ := AddListener(ticket)
		channels = append(channels, channel)
	}
	ShutdownServers()
	if !startBothServersAndTest(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	expectedStatusUpdate = &StatusRecord{
		Id:         ticket,
		Status:     RunningStatus,
		FailReason: NoReason,
		Payload:    nil,
		Errs:       nil,
	}
	_ = UpdateStatus(ticket, expectedStatusUpdate.Status, expectedStatusUpdate.FailReason, expectedStatusUpdate.Payload, expectedStatusUpdate.Errs)
	ShutdownServers()

	for _, channel := range channels {
		statusUpdate, isOpen := <-channel
		if !isOpen || !statusUpdate.isSame(expectedStatusUpdate) {
			t.Errorf("Late listeners should only get the second update. \n found=%+v\n expected=%+v", statusUpdate, expectedStatusUpdate)
		}
	}
}

func TestFinalUpdate(t *testing.T) {
	// Multiple listeners for single ticket
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	ticket := RequestNewTicket()
	numListeners := 100
	channels := []UpdateChannel{}
	for i := 0; i < numListeners; i++ {
		channel, _ := AddListener(ticket)
		channels = append(channels, channel)
	}
	ShutdownServers()
	if listenersStore.Len() != 1 || statusStore.Len() != 1 {
		t.Errorf("Creating listeners for the same ticket should create one status and listener record")
	}

	// Final status update comes in
	if !startBothServersAndTest(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	expectedStatusUpdate := &StatusRecord{
		Id:         ticket,
		Status:     SuccessStatus,
		FailReason: NoReason,
		Payload:    nil,
		Errs:       nil,
	}
	_ = UpdateStatus(ticket, expectedStatusUpdate.Status, expectedStatusUpdate.FailReason, expectedStatusUpdate.Payload, expectedStatusUpdate.Errs)
	ShutdownServers()
	for _, channel := range channels {
		statusUpdate, isOpen := <-channel
		if !isOpen || !statusUpdate.isSame(expectedStatusUpdate) {
			t.Errorf("Early listeners should get the final update. \n found=%+v\n expected=%+v", statusUpdate, expectedStatusUpdate)
		}
		_, isOpen = <-channel
		if isOpen {
			t.Errorf("After final update, listeners channels should be closed")
		}
	}
	if listenersStore.Len() != 0 || statusStore.Len() != 1 {
		t.Errorf("After final update, listeners record should be deleted")
	}

	// Another listener comes in after final update
	if !startBothServersAndTest(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	channel, _ := AddListener(ticket)
	statusUpdate, isOpen := <-channel
	if !isOpen || !statusUpdate.isSame(expectedStatusUpdate) {
		t.Errorf("Listener after final update should get the final update. \n found=%+v\n expected=%+v", statusUpdate, expectedStatusUpdate)
	}
	_, isOpen = <-channel
	if isOpen {
		t.Errorf("After final update, new listeners channels should be closed immediately")
	}
	ShutdownServers()
}

func TestConcurrentUpdates(t *testing.T) {
	if !resetAndStartBothServers(t, multipleWorkersStatusConfig(), multipleWorkersListenersConfig(), false) {
		return
	}
	group := sync.WaitGroup{}
	numTickets := 10
	numListeners := 100
	numListenersTotal := numTickets * numListeners
	numStatusUpdates := numTickets * (SuccessStatus - QueuedStatus + 1)
	group.Add(numListenersTotal + numStatusUpdates)

	// Generate tickets
	tickets := []Ticket{}
	for i := 0; i < numTickets; i++ {
		tickets = append(tickets, RequestNewTicket())
	}

	// Generate all possible status updates
	statusUpdates := []*StatusRecord{}
	for _, ticket := range tickets {
		for status := QueuedStatus; status <= SuccessStatus; status++ {
			statusCode := StatusCode(status)
			statusUpdates = append(statusUpdates, &StatusRecord{
				Id:         ticket,
				Status:     statusCode,
				FailReason: NoReason,
				Payload:    nil,
				Errs:       nil,
			})
		}
	}
	shuffleStatusRecords(statusUpdates)

	// Add listeners concurrently
	channelsLock := &sync.Mutex{}
	channels := []UpdateChannel{}
	go (func() {
		for i := 0; i < numListeners; i++ {
			for ticketIndex := 0; ticketIndex < numTickets; ticketIndex++ {
				ticket := tickets[ticketIndex]
				go (func() {
					waitForRandomDuration()
					channel, _ := AddListener(ticket)
					channelsLock.Lock()
					channels = append(channels, channel)
					channelsLock.Unlock()
					group.Done()
				})()
			}
		}
	})()

	// Make status updates concurrently
	for updateIndex := range statusUpdates {
		updateCached := statusUpdates[updateIndex]
		go (func() {
			waitForRandomDuration()
			_ = UpdateStatus(updateCached.Id, updateCached.Status, updateCached.FailReason, updateCached.Payload, updateCached.Errs)
			group.Done()
		})()
	}

	group.Wait()

	ShutdownServers()

	if len(channels) != numListenersTotal {
		t.Errorf("Total number of channels doesn't match listeners.")
	}
	if len(statusUpdates) != numStatusUpdates {
		t.Errorf("Total number of status updates doesn't match expected number.")
	}
	if listenersStore.Len() != 0 {
		t.Errorf("All listeners records should be deleted after final update.")
	}
	if statusStore.Len() != numTickets {
		t.Errorf("There should be as many status records as tickets.")
	}
	for _, channel := range channels {
		var lastUpdate *StatusRecord
		for update := range channel {
			lastUpdate = update
		}
		if lastUpdate.Status != SuccessStatus {
			t.Errorf("Final update should always be sent to listener channel.")
		}
	}
}
