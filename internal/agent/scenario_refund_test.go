package agent_test

import (
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"github.com/yuriharrison/empirical-proj/internal/testutil"
)

func TestScenario_RefundProductSpecified(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_refund_product_specified")
	ch := subscribeEvents(eventBus, "s1")
	defer unsubscribeEvents(eventBus, "s1")

	messages := []string{
		"I'd like a refund for the wireless headphones I bought last week",
		"They're defective — the left earcup makes a buzzing noise",
	}
	runConversation(t, ag, "s1", messages, false)

	testutil.AssertRefundCount(t, database, 1)
	testutil.AssertLatestRefund(t, database, domain.RefundTypeFull, 149.99)

	events := testutil.CollectEvents(ch)
	if !testutil.HasToolCall(events, "get_refund_policy") {
		t.Error("expected get_refund_policy tool call")
	}
	if !testutil.HasToolCall(events, "issue_refund") {
		t.Error("expected issue_refund tool call")
	}
}

func TestScenario_RefundNoProduct(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_refund_no_product")
	ch := subscribeEvents(eventBus, "s2")
	defer unsubscribeEvents(eventBus, "s2")

	messages := []string{
		"Hi, I need to return something for a refund",
		"The headphones from order 101 please",
		"They're defective",
	}
	runConversation(t, ag, "s2", messages, false)

	testutil.AssertRefundCount(t, database, 1)
	testutil.AssertLatestRefund(t, database, domain.RefundTypeFull, 149.99)

	events := testutil.CollectEvents(ch)
	if !testutil.HasToolCall(events, "lookup_customer_orders") {
		t.Error("expected lookup_customer_orders tool call")
	}
	if !testutil.HasToolCall(events, "issue_refund") {
		t.Error("expected issue_refund tool call")
	}
}

func TestScenario_ComplaintThenRefund(t *testing.T) {
	ag, _, database := setupTestAgent(t, "scenario_complaint_then_refund")

	messages := []string{
		"The t-shirts I ordered are nothing like the pictures on your website",
		"Yes, I'd like a refund please",
		"They don't match the description at all",
	}
	runConversation(t, ag, "s4", messages, false)

	testutil.AssertRefundCount(t, database, 1)
	testutil.AssertLatestRefund(t, database, domain.RefundTypeFull, 69.98)
}

func TestScenario_RefundDenied(t *testing.T) {
	ag, _, database := setupTestAgent(t, "scenario_refund_denied")

	messages := []string{
		"I want to return the t-shirt I bought",
		"I just changed my mind about the color",
		"Okay, I understand",
	}
	responses := runConversation(t, ag, "s5", messages, false)

	testutil.AssertRefundCount(t, database, 0)
	if len(responses) < 2 {
		t.Fatal("expected at least 2 agent responses")
	}
}

func TestScenario_FullRefundAuto(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_full_refund_auto")
	ch := subscribeEvents(eventBus, "s6")
	defer unsubscribeEvents(eventBus, "s6")

	messages := []string{
		"My headphones are broken, I need a refund",
		"They're defective — stopped working after 2 days",
	}
	runConversation(t, ag, "s6", messages, false)

	testutil.AssertRefundCount(t, database, 1)
	testutil.AssertLatestRefund(t, database, domain.RefundTypeFull, 149.99)

	events := testutil.CollectEvents(ch)
	if !testutil.HasEventType(events, chat.EventSystemConfirmation) {
		t.Error("expected system confirmation event for refund")
	}
}

func TestScenario_PartialRefundDeclined(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_partial_refund_declined")
	ch := subscribeEvents(eventBus, "s7")
	defer unsubscribeEvents(eventBus, "s7")

	messages := []string{
		"I want to return my keyboard",
		"I just changed my mind, nothing wrong with it",
		"No, I think I deserve a full refund",
	}
	runConversation(t, ag, "s7", messages, false)

	testutil.AssertRefundCount(t, database, 0)

	events := testutil.CollectEvents(ch)
	if !testutil.HasToolCall(events, "escalate_to_human") {
		t.Error("expected escalate_to_human tool call")
	}
	if !testutil.HasEventType(events, chat.EventSystemEscalation) {
		t.Error("expected system escalation event")
	}
}

func TestScenario_PartialRefundAccepted(t *testing.T) {
	ag, eventBus, database := setupTestAgent(t, "scenario_partial_refund_accepted")
	ch := subscribeEvents(eventBus, "s8")
	defer unsubscribeEvents(eventBus, "s8")

	messages := []string{
		"I want to return my keyboard",
		"Changed my mind",
		"Yes, that's fine, I'll take the partial refund",
	}
	runConversation(t, ag, "s8", messages, false)

	testutil.AssertRefundCount(t, database, 1)
	testutil.AssertLatestRefund(t, database, domain.RefundTypePartial, 55.99)

	events := testutil.CollectEvents(ch)
	if !testutil.HasEventType(events, chat.EventSystemConfirmation) {
		t.Error("expected system confirmation event for partial refund")
	}
}

func TestScenario_SubscriptionTrialRefund(t *testing.T) {
	ag, _, database := setupTestAgent(t, "scenario_subscription_trial_refund")

	messages := []string{
		"I want to cancel CloudSync Pro, I signed up 2 days ago",
		"I just changed my mind",
	}
	runConversation(t, ag, "s11", messages, false)

	testutil.AssertRefundCount(t, database, 1)
	testutil.AssertLatestRefund(t, database, domain.RefundTypeFull, 119.99)
}

func TestScenario_SubscriptionLateCancel(t *testing.T) {
	ag, _, database := setupTestAgent(t, "scenario_subscription_late_cancel")

	messages := []string{
		"Cancel my CloudSync Pro subscription and refund me",
		"I've had it for a few months but never really used it",
		"Fine, I understand",
	}
	runConversation(t, ag, "s12", messages, false)

	testutil.AssertRefundCount(t, database, 0)
}
