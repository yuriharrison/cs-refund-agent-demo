package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/cloudwego/eino/schema"
	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/chat"
)

// ErrSessionEscalated is returned when a message is sent to an escalated session.
var ErrSessionEscalated = errors.New("session already escalated to human agent")

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

	if h.chatService.IsEscalated(sessionID) {
		http.Error(w, `{"error":"session already escalated to human agent"}`, http.StatusConflict)
		return
	}

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
		if errors.Is(err, agent.ErrEscalated) {
			h.chatService.MarkEscalated(sessionID)
			if agentResp != "" {
				h.chatService.AddMessage(sessionID, chat.RoleAgent, agentResp)
			}
			return
		}
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

// SendMessageInternal processes a customer message synchronously, blocking until
// the agent finishes. Used by the usecase runner for step-by-step demo execution.
func (h *ChatHandler) SendMessageInternal(ctx context.Context, sessionID string, content string, headers http.Header) error {
	if h.chatService.IsEscalated(sessionID) {
		return ErrSessionEscalated
	}

	h.chatService.AddMessage(sessionID, chat.RoleCustomer, content)

	history := h.chatService.GetHistory(sessionID)
	messages := h.agent.BuildMessages(history)

	forceError := h.demoForceError
	if headers != nil && headers.Get("X-Demo-Force-Error") != "" {
		forceError = true
	}
	if forceError {
		ctx = agent.WithForceError(ctx)
	}

	agentResp, err := h.agent.ProcessMessage(ctx, sessionID, messages)
	if errors.Is(err, agent.ErrEscalated) {
		h.chatService.MarkEscalated(sessionID)
		if agentResp != "" {
			h.chatService.AddMessage(sessionID, chat.RoleAgent, agentResp)
		}
		return ErrSessionEscalated
	}
	if err != nil {
		return err
	}

	if agentResp != "" {
		h.chatService.AddMessage(sessionID, chat.RoleAgent, agentResp)
	}

	return nil
}
