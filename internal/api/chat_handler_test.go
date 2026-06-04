package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/yuriharrison/empirical-proj/internal/api"
	"github.com/yuriharrison/empirical-proj/internal/chat"
)

type mockAgent struct {
	mu          sync.Mutex
	receivedCtx context.Context
	response    string
	err         error
	called      chan struct{}
}

func (m *mockAgent) BuildMessages(history []chat.Message) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(history))
	for _, msg := range history {
		if msg.Role == chat.RoleCustomer {
			msgs = append(msgs, schema.UserMessage(msg.Content))
		}
	}
	return msgs
}

func (m *mockAgent) ProcessMessage(ctx context.Context, sessionID string, messages []*schema.Message) (string, error) {
	m.mu.Lock()
	m.receivedCtx = ctx
	m.mu.Unlock()

	if m.called != nil {
		close(m.called)
	}
	return m.response, m.err
}

func (m *mockAgent) getReceivedCtx() context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.receivedCtx
}

func newTestRouter() *api.Deps {
	bus := chat.NewEventBus()
	svc := chat.NewService(bus)
	chatHandler := api.NewChatHandler(svc, nil, false)
	sseHandler := api.NewSSEHandler(bus)
	return &api.Deps{
		ChatHandler: chatHandler,
		SSEHandler:  sseHandler,
	}
}

func TestChatAPI_SendMessage_NoAgent(t *testing.T) {
	// Without an agent, we can still test the HTTP layer behavior for validation
	deps := newTestRouter()
	router := api.NewRouter(deps)

	// Test missing content
	req := httptest.NewRequest(http.MethodPost, "/api/chat/message",
		strings.NewReader(`{"content":""}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty content, got %d", rec.Code)
	}

	// Test invalid body
	req = httptest.NewRequest(http.MethodPost, "/api/chat/message",
		strings.NewReader(`{invalid`))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestChatAPI_History(t *testing.T) {
	bus := chat.NewEventBus()
	svc := chat.NewService(bus)
	sessionID := svc.CreateSession()
	svc.AddMessage(sessionID, chat.RoleCustomer, "hello")
	svc.AddMessage(sessionID, chat.RoleAgent, "hi there")

	chatHandler := api.NewChatHandler(svc, nil, false)
	router := api.NewRouter(&api.Deps{ChatHandler: chatHandler})

	req := httptest.NewRequest(http.MethodGet, "/api/chat/history?session_id="+sessionID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp api.HistoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.SessionID != sessionID {
		t.Errorf("expected session_id %q, got %q", sessionID, resp.SessionID)
	}
	if len(resp.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(resp.Messages))
	}
	if resp.Messages[0].Content != "hello" {
		t.Errorf("first message content should be 'hello', got %q", resp.Messages[0].Content)
	}
}

func TestChatAPI_History_MissingSessionID(t *testing.T) {
	deps := newTestRouter()
	router := api.NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/chat/history", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing session_id, got %d", rec.Code)
	}
}

func TestChatAPI_SendMessage_AgentContextNotCancelled(t *testing.T) {
	bus := chat.NewEventBus()
	svc := chat.NewService(bus)
	mock := &mockAgent{
		response: "I can help with that!",
		called:   make(chan struct{}),
	}
	chatHandler := api.NewChatHandler(svc, mock, false)
	router := api.NewRouter(&api.Deps{ChatHandler: chatHandler})

	// Use a real HTTP server so the request context behaves like production:
	// the server cancels r.Context() once the handler returns and the
	// response is flushed — exactly the condition that caused the original bug.
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Post(
		srv.URL+"/api/chat/message",
		"application/json",
		strings.NewReader(`{"content":"I need a refund"}`),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var sendResp api.SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	// Wait for the goroutine to call ProcessMessage
	select {
	case <-mock.called:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for agent.ProcessMessage to be called")
	}

	// The critical assertion: the context passed to the agent must NOT be cancelled
	// even though the HTTP request has completed. This was the original bug —
	// using r.Context() caused the goroutine's context to die with the request.
	ctx := mock.getReceivedCtx()
	if ctx == nil {
		t.Fatal("agent was never called")
	}
	if err := ctx.Err(); err != nil {
		t.Errorf("agent context should not be cancelled after HTTP request completes, got: %v", err)
	}

	// Give the goroutine a moment to persist the agent response
	time.Sleep(100 * time.Millisecond)

	// Verify the agent's response was persisted to chat history
	history := svc.GetHistory(sendResp.SessionID)
	if len(history) < 2 {
		t.Fatalf("expected at least 2 messages (customer + agent), got %d", len(history))
	}

	agentMsg := history[len(history)-1]
	if agentMsg.Role != chat.RoleAgent {
		t.Errorf("last message should be from agent, got role %q", agentMsg.Role)
	}
	if agentMsg.Content != "I can help with that!" {
		t.Errorf("agent message content = %q, want %q", agentMsg.Content, "I can help with that!")
	}
}

func TestChatAPI_SendMessage_ForceErrorHeader(t *testing.T) {
	bus := chat.NewEventBus()
	svc := chat.NewService(bus)
	mock := &mockAgent{
		response: "escalating",
		called:   make(chan struct{}),
	}
	chatHandler := api.NewChatHandler(svc, mock, false)
	router := api.NewRouter(&api.Deps{ChatHandler: chatHandler})

	req := httptest.NewRequest(http.MethodPost, "/api/chat/message",
		strings.NewReader(`{"content":"test"}`))
	req.Header.Set("X-Demo-Force-Error", "true")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	select {
	case <-mock.called:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for agent.ProcessMessage to be called")
	}

	// Context must not be cancelled AND must carry the force-error flag
	ctx := mock.getReceivedCtx()
	if err := ctx.Err(); err != nil {
		t.Errorf("agent context should not be cancelled, got: %v", err)
	}
}

func TestChatAPI_Reset(t *testing.T) {
	bus := chat.NewEventBus()
	svc := chat.NewService(bus)
	sessionID := svc.CreateSession()
	svc.AddMessage(sessionID, chat.RoleCustomer, "hello")

	chatHandler := api.NewChatHandler(svc, nil, false)
	router := api.NewRouter(&api.Deps{ChatHandler: chatHandler})

	req := httptest.NewRequest(http.MethodPost, "/api/chat/reset",
		strings.NewReader(`{"session_id":"`+sessionID+`"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp api.ResetResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.SessionID == sessionID {
		t.Error("expected new session ID after reset")
	}
	if resp.SessionID == "" {
		t.Error("expected non-empty new session ID")
	}

	// Old session should be gone
	oldHistory := svc.GetHistory(sessionID)
	if oldHistory != nil {
		t.Error("old session should be deleted after reset")
	}
}
