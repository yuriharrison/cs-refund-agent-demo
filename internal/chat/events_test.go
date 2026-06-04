package chat_test

import (
	"testing"
	"time"

	"github.com/yuriharrison/empirical-proj/internal/chat"
)

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := chat.NewEventBus()
	events := bus.Subscribe("session-1", "sub-1")

	sent := chat.Event{
		Type:      chat.EventAgentMessage,
		SessionID: "session-1",
		Data:      map[string]string{"content": "hello"},
	}
	bus.Publish(sent)

	select {
	case received := <-events:
		if received.Type != sent.Type {
			t.Errorf("expected event type %q, got %q", sent.Type, received.Type)
		}
		if received.SessionID != sent.SessionID {
			t.Errorf("expected session %q, got %q", sent.SessionID, received.SessionID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := chat.NewEventBus()
	ch1 := bus.Subscribe("session-1", "sub-1")
	ch2 := bus.Subscribe("session-1", "sub-2")

	sent := chat.Event{
		Type:      chat.EventAgentThinking,
		SessionID: "session-1",
		Data:      map[string]string{"content": "thinking..."},
	}
	bus.Publish(sent)

	for _, ch := range []<-chan chat.Event{ch1, ch2} {
		select {
		case received := <-ch:
			if received.Type != sent.Type {
				t.Errorf("expected event type %q, got %q", sent.Type, received.Type)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for event on a subscriber")
		}
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := chat.NewEventBus()
	events := bus.Subscribe("session-1", "sub-1")

	bus.Unsubscribe("session-1", "sub-1")

	bus.Publish(chat.Event{
		Type:      chat.EventAgentMessage,
		SessionID: "session-1",
		Data:      map[string]string{"content": "should not arrive"},
	})

	select {
	case _, ok := <-events:
		if ok {
			t.Error("expected channel to be closed after unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected channel to be closed immediately after unsubscribe")
	}
}

func TestEventBus_DifferentSessions(t *testing.T) {
	bus := chat.NewEventBus()
	ch1 := bus.Subscribe("session-1", "sub-1")
	ch2 := bus.Subscribe("session-2", "sub-1")

	bus.Publish(chat.Event{
		Type:      chat.EventAgentMessage,
		SessionID: "session-1",
		Data:      map[string]string{"content": "for session 1"},
	})

	select {
	case <-ch1:
		// expected
	case <-time.After(time.Second):
		t.Fatal("session-1 subscriber should have received event")
	}

	select {
	case <-ch2:
		t.Error("session-2 subscriber should not have received event")
	case <-time.After(100 * time.Millisecond):
		// expected
	}
}
