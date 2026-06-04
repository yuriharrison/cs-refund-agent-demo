package agent_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"github.com/yuriharrison/empirical-proj/internal/token"
	"gorm.io/gorm"
)

func setupLiveAgent(t *testing.T) (*agent.Agent, *chat.EventBus, *gorm.DB, *token.TokenTracker) {
	t.Helper()

	apiKey := os.Getenv("OPEN_ROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPEN_ROUTER_API_KEY not set, skipping live API test")
	}

	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: "https://openrouter.ai/api/v1",
		Model:   "deepseek/deepseek-v4-flash",
	})
	if err != nil {
		t.Fatalf("failed to create chat model: %v", err)
	}

	database := newSeededTestDB(t)
	eventBus := chat.NewEventBus()
	tracker := token.NewTracker()

	tools, err := agent.BuildTools(database, 1, eventBus)
	if err != nil {
		t.Fatalf("failed to build tools: %v", err)
	}

	var customer domain.Customer
	if err := database.First(&customer).Error; err != nil {
		t.Fatalf("failed to load customer: %v", err)
	}

	ag, err := agent.New(ctx, agent.Config{
		ChatModel:        chatModel,
		DB:               database,
		Customer:         customer,
		EventBus:         eventBus,
		Tools:            tools,
		TokenTracker:     tracker,
		DisableStreaming: true,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	return ag, eventBus, database, tracker
}

func TestLiveAPI_OpenRouter_SimpleMessage(t *testing.T) {
	ag, _, _, _ := setupLiveAgent(t)

	ctx := context.Background()
	messages := []*schema.Message{
		schema.UserMessage("Hi, I need help with a refund for my headphones"),
	}
	resp, err := ag.ProcessMessage(ctx, "live-test-simple", messages)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if resp == "" {
		t.Fatal("expected non-empty response from agent")
	}

	t.Logf("Agent response: %s", resp)
}

func TestLiveAPI_OpenRouter_FullRefundFlow(t *testing.T) {
	ag, eventBus, database, _ := setupLiveAgent(t)

	sessionID := "live-test-refund"
	events := eventBus.Subscribe(sessionID, "test-sub")
	defer eventBus.Unsubscribe(sessionID, "test-sub")

	responses := runConversation(t, ag, sessionID, []string{
		"I need a refund for my headphones, they're defective — buzzing in the left ear",
	}, false)

	if len(responses) == 0 {
		t.Fatal("expected at least one response")
	}

	lastResponse := responses[len(responses)-1]
	t.Logf("Final response: %s", lastResponse)

	hasRefundMention := strings.Contains(strings.ToLower(lastResponse), "refund") ||
		strings.Contains(strings.ToLower(lastResponse), "$149.99") ||
		strings.Contains(strings.ToLower(lastResponse), "processed")
	if !hasRefundMention {
		t.Logf("Warning: response may not contain refund confirmation, got: %s", lastResponse)
	}

	var refundCount int64
	database.Table("refunds").Count(&refundCount)
	t.Logf("Refunds in DB: %d", refundCount)

	var sawToolCall, sawAgentMsg bool
	drainLoop:
	for {
		select {
		case ev := <-events:
			switch ev.Type {
			case chat.EventToolCallStart:
				sawToolCall = true
			case chat.EventAgentMessage:
				sawAgentMsg = true
			}
		default:
			break drainLoop
		}
	}

	if !sawToolCall {
		t.Error("expected at least one tool call event")
	}
	if !sawAgentMsg {
		t.Error("expected at least one agent message event")
	}
}
