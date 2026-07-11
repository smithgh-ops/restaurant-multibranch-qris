package kds

import (
	"encoding/json"
	"testing"
)

func TestHubBroadcastSubscribeAndUnsubscribe(t *testing.T) {
	hub := NewHub()
	branchOneClientA := &realtimeClient{send: make(chan []byte, 1)}
	branchOneClientB := &realtimeClient{send: make(chan []byte, 1)}
	branchTwoClient := &realtimeClient{send: make(chan []byte, 1)}

	hub.addClient(1, branchOneClientA)
	hub.addClient(1, branchOneClientB)
	hub.addClient(2, branchTwoClient)

	event := RealtimeEvent{Type: "ticket.created", Ticket: &KitchenTicket{ID: 42}}
	if err := hub.Broadcast(1, event); err != nil {
		t.Fatalf("broadcast returned error: %v", err)
	}

	assertEventTicketID(t, branchOneClientA.send, 42)
	assertEventTicketID(t, branchOneClientB.send, 42)

	select {
	case <-branchTwoClient.send:
		t.Fatal("branch 2 client unexpectedly received event")
	default:
	}

	hub.removeClient(1, branchOneClientA)
	if len(hub.clients[1]) != 1 {
		t.Fatalf("expected 1 branch 1 client after unsubscribe, got %d", len(hub.clients[1]))
	}
}

func assertEventTicketID(t *testing.T, ch <-chan []byte, expectedID uint64) {
	t.Helper()
	select {
	case payload := <-ch:
		var event RealtimeEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatalf("unmarshal broadcast payload: %v", err)
		}
		if event.Ticket == nil || event.Ticket.ID != expectedID {
			t.Fatalf("expected ticket id %d, got %+v", expectedID, event.Ticket)
		}
	default:
		t.Fatalf("expected event with ticket id %d", expectedID)
	}
}
