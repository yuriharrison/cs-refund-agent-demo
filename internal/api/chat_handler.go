package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cloudwego/eino/schema"
	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/chat"
)

type MessageProcessor interface {
	BuildMessages(history []chat.Message) []*schema.Message
	ProcessMessage(ctx context.Context, sessionID string, messages []*schema.Message) (string, error)
}

type ChatHandler struct {
	chatService    *chat.Service
	agent          MessageProcessor
	demoForceError bool
}

func NewChatHandler(chatService *chat.Service, ag MessageProcessor, demoForceError bool) *ChatHandler {
	return &ChatHandler{
		chatService:    chatService,
		agent:          ag,
		demoForceError: demoForceError,
	}
}

type SendMessageRequest struct {
	Content   string `json:"content"`
	SessionID string `json:"session_id,omitempty"`
}

type SendMessageResponse struct {
	SessionID string `json:"session_id"`
	MessageID string `json:"message_id"`
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, `{"error":"content is required"}`, http.StatusBadRequest)
		return
	}

	sessionID := h.chatService.GetOrCreateSession(req.SessionID)
	msg := h.chatService.AddMessage(sessionID, chat.RoleCustomer, req.Content)

	forceError := h.demoForceError || r.Header.Get("X-Demo-Force-Error") != ""

	go func() {
		history := h.chatService.GetHistory(sessionID)
		messages := h.agent.BuildMessages(history)

		ctx := context.Background()
		if forceError {
			ctx = agent.WithForceError(ctx)
		}

		agentResp, err := h.agent.ProcessMessage(ctx, sessionID, messages)
		if err != nil {
			slog.Error("agent processing failed", "error", err, "session_id", sessionID)
			return
		}

		if agentResp != "" {
			h.chatService.AddMessage(sessionID, chat.RoleAgent, agentResp)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(SendMessageResponse{
		SessionID: sessionID,
		MessageID: msg.ID,
	})
}

type HistoryResponse struct {
	SessionID string         `json:"session_id"`
	Messages  []chat.Message `json:"messages"`
}

func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, `{"error":"session_id is required"}`, http.StatusBadRequest)
		return
	}

	messages := h.chatService.GetHistory(sessionID)
	if messages == nil {
		messages = []chat.Message{}
	}

	json.NewEncoder(w).Encode(HistoryResponse{
		SessionID: sessionID,
		Messages:  messages,
	})
}

type ResetResponse struct {
	SessionID string `json:"session_id"`
}

func (h *ChatHandler) ResetSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		http.Error(w, `{"error":"session_id is required"}`, http.StatusBadRequest)
		return
	}

	newID := h.chatService.Reset(req.SessionID)
	json.NewEncoder(w).Encode(ResetResponse{SessionID: newID})
}

// buildAgentMessages converts a single user message into the Eino schema format.
// This is used when we only want to pass the latest message without full history.
func buildAgentMessages(content string) []*schema.Message {
	return []*schema.Message{
		schema.UserMessage(content),
	}
}
