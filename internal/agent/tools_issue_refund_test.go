package agent_test

import (
	"encoding/json"
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/domain"
)

func TestIssueRefund_CreatesRecord(t *testing.T) {
	database := newSeededTestDB(t)
	eventBus := chat.NewEventBus()

	var orderItem domain.OrderItem
	if err := database.First(&orderItem).Error; err != nil {
		t.Fatalf("failed to find order item: %v", err)
	}

	tool, err := agent.NewIssueRefundTool(database, eventBus)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := agent.WithSessionID(t.Context(), "test-session")
	result, err := tool.InvokableRun(ctx, mustJSON(map[string]interface{}{
		"order_item_id": orderItem.ID,
		"refund_type":   "full",
		"amount":        149.99,
		"reason":        "defective product",
	}))
	if err != nil {
		t.Fatalf("tool invocation failed: %v", err)
	}

	var resp agent.IssueRefundResult
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if resp.Status != "approved" {
		t.Errorf("expected status approved, got %q", resp.Status)
	}
	if resp.Type != "full" {
		t.Errorf("expected type full, got %q", resp.Type)
	}
	if resp.Amount != 149.99 {
		t.Errorf("expected amount 149.99, got %.2f", resp.Amount)
	}

	var refund domain.Refund
	if err := database.First(&refund, resp.RefundID).Error; err != nil {
		t.Fatalf("refund not found in DB: %v", err)
	}
	if refund.OrderItemID != orderItem.ID {
		t.Errorf("expected order_item_id %d, got %d", orderItem.ID, refund.OrderItemID)
	}
	if refund.DecidedBy != domain.RefundDecidedByAgent {
		t.Errorf("expected decided_by agent, got %q", refund.DecidedBy)
	}
}

func TestIssueRefund_EmitsSystemConfirmation(t *testing.T) {
	database := newSeededTestDB(t)
	eventBus := chat.NewEventBus()
	sessionID := "test-session-confirm"

	ch := eventBus.Subscribe(sessionID, "test-subscriber")
	defer eventBus.Unsubscribe(sessionID, "test-subscriber")

	var orderItem domain.OrderItem
	if err := database.First(&orderItem).Error; err != nil {
		t.Fatalf("failed to find order item: %v", err)
	}

	tool, err := agent.NewIssueRefundTool(database, eventBus)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := agent.WithSessionID(t.Context(), sessionID)
	result, err := tool.InvokableRun(ctx, mustJSON(map[string]interface{}{
		"order_item_id": orderItem.ID,
		"refund_type":   "partial",
		"amount":        75.00,
		"reason":        "partial refund per policy",
	}))
	if err != nil {
		t.Fatalf("tool invocation failed: %v", err)
	}

	var resp agent.IssueRefundResult
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	select {
	case event := <-ch:
		if event.Type != chat.EventSystemConfirmation {
			t.Fatalf("expected system_confirmation event, got %q", event.Type)
		}
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("unexpected event data type: %T", event.Data)
		}
		if data["action"] != "refund_issued" {
			t.Errorf("expected action refund_issued, got %v", data["action"])
		}
		details, ok := data["details"].(map[string]interface{})
		if !ok {
			t.Fatalf("unexpected details type: %T", data["details"])
		}
		if asUint(details["refund_id"]) != resp.RefundID {
			t.Errorf("expected refund_id %d, got %v", resp.RefundID, details["refund_id"])
		}
		if details["amount"] != 75.0 {
			t.Errorf("expected amount 75.0, got %v", details["amount"])
		}
		if details["type"] != "partial" {
			t.Errorf("expected type partial, got %v", details["type"])
		}
		if asUint(details["order_item_id"]) != orderItem.ID {
			t.Errorf("expected order_item_id %d, got %v", orderItem.ID, details["order_item_id"])
		}
	default:
		t.Fatal("expected system_confirmation event, got none")
	}
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func asUint(v interface{}) uint {
	switch n := v.(type) {
	case uint:
		return n
	case uint64:
		return uint(n)
	case int:
		return uint(n)
	case int64:
		return uint(n)
	case float64:
		return uint(n)
	default:
		return 0
	}
}
