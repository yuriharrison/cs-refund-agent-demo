package agent_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/chat"
)

func TestEscalateToHuman_EmitsEvents(t *testing.T) {
	eventBus := chat.NewEventBus()
	sessionID := "test-session-escalate"

	ch := eventBus.Subscribe(sessionID, "test-subscriber")
	defer eventBus.Unsubscribe(sessionID, "test-subscriber")

	tool, err := agent.NewEscalateToHumanTool(eventBus)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := agent.WithSessionID(t.Context(), sessionID)
	done := make(chan struct{})
	var toolResult string
	var toolErr error

	go func() {
		toolResult, toolErr = tool.InvokableRun(ctx, `{"reason":"software refund requires human review"}`)
		close(done)
	}()

	var escalationEvent chat.Event
	select {
	case escalationEvent = <-ch:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for system_escalation event")
	}

	if escalationEvent.Type != chat.EventSystemEscalation {
		t.Fatalf("expected system_escalation event, got %q", escalationEvent.Type)
	}
	escData, ok := escalationEvent.Data.(map[string]string)
	if !ok {
		t.Fatalf("unexpected escalation data type: %T", escalationEvent.Data)
	}
	if escData["reason"] != "software refund requires human review" {
		t.Errorf("unexpected reason: %q", escData["reason"])
	}
	if escData["human_agent_name"] != "Alex" {
		t.Errorf("expected human_agent_name Alex, got %q", escData["human_agent_name"])
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for tool to complete")
	}

	if toolErr != nil {
		t.Fatalf("tool invocation failed: %v", toolErr)
	}

	var resp agent.EscalateResult
	if err := json.Unmarshal([]byte(toolResult), &resp); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if !resp.Escalated {
		t.Error("expected escalated=true")
	}

	var humanEvent chat.Event
	select {
	case humanEvent = <-ch:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for human_message event")
	}

	if humanEvent.Type != chat.EventHumanMessage {
		t.Fatalf("expected human_message event, got %q", humanEvent.Type)
	}
	humanData, ok := humanEvent.Data.(map[string]string)
	if !ok {
		t.Fatalf("unexpected human message data type: %T", humanEvent.Data)
	}
	if humanData["agent_name"] != "Alex" {
		t.Errorf("expected agent_name Alex, got %q", humanData["agent_name"])
	}
	if humanData["content"] == "" {
		t.Error("expected non-empty human message content")
	}
}
