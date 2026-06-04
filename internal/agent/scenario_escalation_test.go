package agent_test

import (
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/testutil"
)

func TestScenario_EscalateByPolicy(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_escalate_by_policy")
	ch := subscribeEvents(eventBus, "s9")
	defer unsubscribeEvents(eventBus, "s9")

	messages := []string{
		"I need a refund for the photo editing software",
		"It crashes every time I try to export",
	}
	runConversation(t, ag, "s9", messages, false)

	testutil.AssertRefundCount(t, database, 0)

	events := testutil.CollectEvents(ch)
	if !testutil.HasToolCall(events, "get_refund_policy") {
		t.Error("expected get_refund_policy tool call")
	}
	if !testutil.HasToolCall(events, "escalate_to_human") {
		t.Error("expected escalate_to_human tool call")
	}
	if !testutil.HasEventType(events, chat.EventSystemEscalation) {
		t.Error("expected system escalation event")
	}
}

func TestScenario_EscalateByError(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_escalate_by_error")
	ch := subscribeEvents(eventBus, "s10")
	defer unsubscribeEvents(eventBus, "s10")

	messages := []string{
		"I want a refund for my running shoes",
		"They fell apart after one run",
	}
	runConversation(t, ag, "s10", messages, true)

	testutil.AssertRefundCount(t, database, 0)

	events := testutil.CollectEvents(ch)
	if !testutil.HasToolCall(events, "escalate_to_human") {
		t.Error("expected escalate_to_human tool call after policy error")
	}
	if !testutil.HasEventType(events, chat.EventSystemEscalation) {
		t.Error("expected system escalation event")
	}
}
