package agent_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"github.com/yuriharrison/empirical-proj/internal/testutil"
	"github.com/yuriharrison/empirical-proj/internal/token"
	"gorm.io/gorm"
)

type testAgentOptions struct {
	forceError bool
}

func setupTestAgent(t *testing.T, snapshotName string, opts ...testAgentOptions) (*agent.Agent, *chat.EventBus, *gorm.DB) {
	t.Helper()

	var opt testAgentOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	database := newSeededTestDB(t)
	eventBus := chat.NewEventBus()

	transport := testutil.NewSnapshotTransport(t, snapshotName)
	testutil.SkipIfRecordingWithoutAPIKey(t, transport)

	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:     "test-key",
		BaseURL:    "https://openrouter.ai/api/v1",
		Model:      "deepseek/deepseek-v4-flash",
		HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatalf("failed to create chat model: %v", err)
	}

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
		TokenTracker:     token.NewTracker(),
		DisableStreaming: true,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	_ = opt
	return ag, eventBus, database
}

func runConversation(t *testing.T, ag *agent.Agent, sessionID string, messages []string, forceError bool) []string {
	t.Helper()

	ctx := context.Background()
	if forceError {
		ctx = agent.WithForceError(ctx)
	}

	var responses []string
	var history []*schema.Message

	for _, msg := range messages {
		history = append(history, schema.UserMessage(msg))
		resp, err := ag.ProcessMessage(ctx, sessionID, history)
		if err != nil {
			t.Fatalf("agent error on message %q: %v", msg, err)
		}
		responses = append(responses, resp)
		history = append(history, &schema.Message{
			Role:    schema.Assistant,
			Content: resp,
		})
	}

	return responses
}

func subscribeEvents(eventBus *chat.EventBus, sessionID string) <-chan chat.Event {
	return eventBus.Subscribe(sessionID, "test-subscriber")
}

func unsubscribeEvents(eventBus *chat.EventBus, sessionID string) {
	eventBus.Unsubscribe(sessionID, "test-subscriber")
}
