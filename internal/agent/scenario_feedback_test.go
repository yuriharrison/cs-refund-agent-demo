package agent_test

import (
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/testutil"
)

func TestScenario_FeedbackOnly(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_feedback_only")
	ch := subscribeEvents(eventBus, "s3")
	defer unsubscribeEvents(eventBus, "s3")

	messages := []string{
		"I'm not happy with the meal kit I received",
		"No, I don't need a refund. Just wanted to let you know the quality was below expectations",
		"That's all, thanks",
	}
	runConversation(t, ag, "s3", messages, false)

	testutil.AssertRefundCount(t, database, 0)

	events := testutil.CollectEvents(ch)
	testutil.AssertNoToolCall(t, events, "issue_refund")
	testutil.AssertNoToolCall(t, events, "escalate_to_human")
}
