package chat_test

import (
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/chat"
)

func TestChatService_CreateSession(t *testing.T) {
	svc := chat.NewService(chat.NewEventBus())
	id := svc.CreateSession()

	if id == "" {
		t.Fatal("expected non-empty session ID")
	}

	id2 := svc.CreateSession()
	if id == id2 {
		t.Error("expected different session IDs")
	}
}

func TestChatService_AddMessage(t *testing.T) {
	svc := chat.NewService(chat.NewEventBus())
	sessionID := svc.CreateSession()

	msg1 := svc.AddMessage(sessionID, chat.RoleCustomer, "hello")
	msg2 := svc.AddMessage(sessionID, chat.RoleAgent, "how can I help?")

	if msg1.Role != chat.RoleCustomer {
		t.Errorf("expected role %q, got %q", chat.RoleCustomer, msg1.Role)
	}
	if msg2.Content != "how can I help?" {
		t.Errorf("expected content 'how can I help?', got %q", msg2.Content)
	}

	history := svc.GetHistory(sessionID)
	if len(history) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history))
	}

	if history[0].Content != "hello" {
		t.Errorf("first message should be 'hello', got %q", history[0].Content)
	}
	if history[1].Content != "how can I help?" {
		t.Errorf("second message should be 'how can I help?', got %q", history[1].Content)
	}
}

func TestChatService_GetHistory_EmptySession(t *testing.T) {
	svc := chat.NewService(chat.NewEventBus())
	history := svc.GetHistory("nonexistent")

	if history != nil {
		t.Errorf("expected nil for nonexistent session, got %v", history)
	}
}

func TestChatService_Reset(t *testing.T) {
	svc := chat.NewService(chat.NewEventBus())
	sessionID := svc.CreateSession()
	svc.AddMessage(sessionID, chat.RoleCustomer, "hello")

	newID := svc.Reset(sessionID)

	if newID == sessionID {
		t.Error("expected new session ID after reset")
	}

	oldHistory := svc.GetHistory(sessionID)
	if oldHistory != nil {
		t.Error("expected old session to be deleted")
	}

	newHistory := svc.GetHistory(newID)
	if newHistory == nil {
		t.Fatal("expected new session to exist")
	}
	if len(newHistory) != 0 {
		t.Errorf("expected empty history for new session, got %d messages", len(newHistory))
	}
}

func TestChatService_GetOrCreateSession(t *testing.T) {
	svc := chat.NewService(chat.NewEventBus())

	// Empty string creates new session
	id1 := svc.GetOrCreateSession("")
	if id1 == "" {
		t.Fatal("expected non-empty session ID")
	}

	// Existing session returns same ID
	id2 := svc.GetOrCreateSession(id1)
	if id2 != id1 {
		t.Errorf("expected same session ID %q, got %q", id1, id2)
	}

	// Nonexistent session creates new one
	id3 := svc.GetOrCreateSession("nonexistent-id")
	if id3 == "nonexistent-id" {
		t.Error("expected new session ID for nonexistent session")
	}
}
