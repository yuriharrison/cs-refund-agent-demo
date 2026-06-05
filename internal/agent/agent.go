package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"github.com/yuriharrison/empirical-proj/internal/token"
	"gorm.io/gorm"
)

// ErrEscalated is returned by ProcessMessage when the agent escalated to a human.
// Callers should treat this as a terminal state and stop sending further messages.
var ErrEscalated = errors.New("conversation escalated to human agent")

type Agent struct {
	runner        *adk.Runner
	db            *gorm.DB
	customer      domain.Customer
	eventBus      *chat.EventBus
	tokenTracker  *token.TokenTracker
	seenSessions  sync.Map
	toolCallNames sync.Map
}

type Config struct {
	ChatModel        model.ToolCallingChatModel
	DB               *gorm.DB
	Customer         domain.Customer
	EventBus         *chat.EventBus
	Tools            []tool.BaseTool
	TokenTracker     *token.TokenTracker
	DisableStreaming bool
}

func New(ctx context.Context, cfg Config) (*Agent, error) {
	instruction := BuildSystemPrompt(cfg.Customer.Name, cfg.Customer.Email)

	agentInstance, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "shopease-support",
		Description: "Customer support agent for ShopEase that handles refund workflows",
		Instruction: instruction,
		Model:       cfg.ChatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: cfg.Tools,
			},
		},
		MaxIterations: 15,
	})
	if err != nil {
		return nil, fmt.Errorf("creating agent: %w", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agentInstance,
		EnableStreaming: !cfg.DisableStreaming,
	})

	return &Agent{
		runner:       runner,
		db:           cfg.DB,
		customer:     cfg.Customer,
		eventBus:     cfg.EventBus,
		tokenTracker: cfg.TokenTracker,
	}, nil
}

func (a *Agent) ProcessMessage(ctx context.Context, sessionID string, messages []*schema.Message) (string, error) {
	if _, loaded := a.seenSessions.LoadOrStore(sessionID, true); !loaded && a.tokenTracker != nil {
		a.tokenTracker.IncrementSessions()
	}

	a.eventBus.Publish(chat.Event{
		Type:      chat.EventAgentThinking,
		SessionID: sessionID,
		Data:      map[string]string{"content": "Processing your request..."},
	})

	ctx = WithSessionID(ctx, sessionID)
	iter := a.runner.Run(ctx, messages)

	var finalContent string
	var escalated bool

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			a.eventBus.Publish(chat.Event{
				Type:      chat.EventError,
				SessionID: sessionID,
				Data:      map[string]string{"message": event.Err.Error()},
			})
			return "", fmt.Errorf("agent error: %w", event.Err)
		}

		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		msgOutput := event.Output.MessageOutput

		if msgOutput.IsStreaming && msgOutput.MessageStream != nil {
			if !escalated {
				content, streamEscalated := a.handleStreamingMessage(sessionID, msgOutput.MessageStream)
				finalContent = content
				if streamEscalated {
					escalated = true
				}
			} else {
				a.drainStream(msgOutput.MessageStream)
			}
			continue
		}

		msg, err := msgOutput.GetMessage()
		if err != nil {
			slog.Error("failed to get message from event", "error", err)
			continue
		}

		if msg == nil {
			continue
		}

		switch msg.Role {
		case schema.Assistant:
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					if tc.Function.Name == "escalate_to_human" {
						escalated = true
					}
				}
				a.handleToolCalls(sessionID, msg.ToolCalls)
			}
			if msg.Content != "" && !escalated {
				finalContent = msg.Content
				a.eventBus.Publish(chat.Event{
					Type:      chat.EventAgentMessage,
					SessionID: sessionID,
					Data:      map[string]string{"content": msg.Content},
				})
			}

		case schema.Tool:
			a.handleToolResult(sessionID, msg)

		default:
			if msg.Content != "" && !escalated {
				finalContent = msg.Content
				a.eventBus.Publish(chat.Event{
					Type:      chat.EventAgentMessage,
					SessionID: sessionID,
					Data:      map[string]string{"content": msg.Content},
				})
			}
		}

		a.recordTokenUsage(sessionID, msg)
	}

	if escalated {
		return finalContent, ErrEscalated
	}
	return finalContent, nil
}

func (a *Agent) handleStreamingMessage(sessionID string, stream *schema.StreamReader[*schema.Message]) (string, bool) {
	var fullContent string
	var escalated bool
	pendingToolCalls := make(map[string]*schema.ToolCall)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("stream recv error", "error", err)
			break
		}
		if msg == nil {
			continue
		}

		if msg.Content != "" && !escalated {
			fullContent += msg.Content
			a.eventBus.Publish(chat.Event{
				Type:      chat.EventAgentMessage,
				SessionID: sessionID,
				Data:      map[string]string{"content": msg.Content},
			})
		}

		for i := range msg.ToolCalls {
			tc := msg.ToolCalls[i]
			if existing, ok := pendingToolCalls[tc.ID]; ok {
				existing.Function.Arguments += tc.Function.Arguments
			} else {
				clone := tc
				pendingToolCalls[tc.ID] = &clone
			}
		}

		a.recordTokenUsage(sessionID, msg)
	}

	if len(pendingToolCalls) > 0 {
		completed := make([]schema.ToolCall, 0, len(pendingToolCalls))
		for _, tc := range pendingToolCalls {
			if tc.Function.Name == "escalate_to_human" {
				escalated = true
			}
			completed = append(completed, *tc)
		}
		a.handleToolCalls(sessionID, completed)
	}

	return fullContent, escalated
}

func (a *Agent) drainStream(stream *schema.StreamReader[*schema.Message]) {
	for {
		_, err := stream.Recv()
		if err != nil {
			return
		}
	}
}

func (a *Agent) recordTokenUsage(sessionID string, msg *schema.Message) {
	if msg.ResponseMeta == nil || msg.ResponseMeta.Usage == nil {
		return
	}

	usage := msg.ResponseMeta.Usage
	if a.tokenTracker != nil {
		a.tokenTracker.Record(usage.PromptTokens, usage.CompletionTokens)
	}

	a.eventBus.Publish(chat.Event{
		Type:      chat.EventTokenUpdate,
		SessionID: sessionID,
		Data: map[string]int64{
			"prompt_tokens":     int64(usage.PromptTokens),
			"completion_tokens": int64(usage.CompletionTokens),
			"total":             int64(usage.TotalTokens),
		},
	})
}

func (a *Agent) handleToolCalls(sessionID string, toolCalls []schema.ToolCall) {
	for _, tc := range toolCalls {
		a.toolCallNames.Store(tc.ID, tc.Function.Name)

		var args interface{}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			args = tc.Function.Arguments
		}

		a.eventBus.Publish(chat.Event{
			Type:      chat.EventToolCallStart,
			SessionID: sessionID,
			Data: map[string]interface{}{
				"tool":      tc.Function.Name,
				"arguments": args,
			},
		})
	}
}

func (a *Agent) handleToolResult(sessionID string, msg *schema.Message) {
	toolName := msg.ToolCallID
	if name, ok := a.toolCallNames.LoadAndDelete(msg.ToolCallID); ok {
		toolName = name.(string)
	}

	var resultData interface{}
	if err := json.Unmarshal([]byte(msg.Content), &resultData); err != nil {
		resultData = msg.Content
	}

	a.eventBus.Publish(chat.Event{
		Type:      chat.EventToolCallResult,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"tool":        toolName,
			"result":      resultData,
			"duration_ms": 0,
		},
	})
}

func (a *Agent) BuildMessages(history []chat.Message) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(history))
	for _, m := range history {
		switch m.Role {
		case chat.RoleCustomer:
			msgs = append(msgs, schema.UserMessage(m.Content))
		case chat.RoleAgent:
			msgs = append(msgs, &schema.Message{
				Role:    schema.Assistant,
				Content: m.Content,
			})
		}
	}
	return msgs
}

// timed wraps a tool call with timing and publishes start/result events.
// This is a utility to be used by tool implementations in the future.
func timed(sessionID string, bus *chat.EventBus, toolName string, fn func() (string, error)) (string, error) {
	start := time.Now()
	result, err := fn()
	duration := time.Since(start).Milliseconds()

	var resultData interface{}
	if err == nil {
		if jsonErr := json.Unmarshal([]byte(result), &resultData); jsonErr != nil {
			resultData = result
		}
	} else {
		resultData = map[string]string{"error": err.Error()}
	}

	bus.Publish(chat.Event{
		Type:      chat.EventToolCallResult,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"tool":        toolName,
			"result":      resultData,
			"duration_ms": duration,
		},
	})

	return result, err
}
