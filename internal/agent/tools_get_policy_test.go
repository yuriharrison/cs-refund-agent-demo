package agent_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/agent"
)

func invokePolicy(t *testing.T, input string) agent.PolicyResult {
	t.Helper()
	db := newSeededTestDB(t)

	tool, err := agent.NewGetRefundPolicyTool(db)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	result, err := tool.InvokableRun(t.Context(), input)
	if err != nil {
		t.Fatalf("tool invocation failed: %v", err)
	}

	var policy agent.PolicyResult
	if err := json.Unmarshal([]byte(result), &policy); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	return policy
}

func TestGetRefundPolicy_Electronics_Defective(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"electronics","condition":"defective"}`)

	if policy.Action != "full_refund" {
		t.Errorf("expected full_refund, got %q", policy.Action)
	}
	if policy.WindowDays == nil || *policy.WindowDays != 30 {
		t.Errorf("expected 30-day window, got %v", policy.WindowDays)
	}
}

func TestGetRefundPolicy_Electronics_ChangeOfMind(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"electronics","condition":"change_of_mind"}`)

	if policy.Action != "partial_refund" {
		t.Errorf("expected partial_refund, got %q", policy.Action)
	}
	if policy.PartialPercent == nil || *policy.PartialPercent != 70 {
		t.Errorf("expected 70%% partial, got %v", policy.PartialPercent)
	}
	if policy.WindowDays == nil || *policy.WindowDays != 15 {
		t.Errorf("expected 15-day window, got %v", policy.WindowDays)
	}
}

func TestGetRefundPolicy_Clothing_ChangeOfMind(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"clothing","condition":"change_of_mind"}`)

	if policy.Action != "no_refund" {
		t.Errorf("expected no_refund, got %q", policy.Action)
	}
}

func TestGetRefundPolicy_Software_Any(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"software","condition":"defective"}`)

	if policy.Action != "escalate" {
		t.Errorf("expected escalate, got %q", policy.Action)
	}
}

func TestGetRefundPolicy_Subscription_TrialWindow(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"subscription","condition":"change_of_mind"}`)

	if policy.Action != "full_refund" {
		t.Errorf("expected full_refund, got %q", policy.Action)
	}
	if policy.WindowDays == nil || *policy.WindowDays != 7 {
		t.Errorf("expected 7-day window, got %v", policy.WindowDays)
	}
}

func TestGetRefundPolicy_Subscription_Cancel(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"subscription","condition":"subscription_cancel"}`)

	if policy.Action != "partial_refund" {
		t.Errorf("expected partial_refund, got %q", policy.Action)
	}
	if policy.PartialPercent == nil || *policy.PartialPercent != 50 {
		t.Errorf("expected 50%% partial, got %v", policy.PartialPercent)
	}
	if policy.WindowDays == nil || *policy.WindowDays != 3 {
		t.Errorf("expected 3-day window, got %v", policy.WindowDays)
	}
}

func TestGetRefundPolicy_FoodAny_NoRefund(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"food","condition":"change_of_mind"}`)

	if policy.Action != "no_refund" {
		t.Errorf("expected no_refund, got %q", policy.Action)
	}
}

func TestGetRefundPolicy_NoMatch_Fallback(t *testing.T) {
	policy := invokePolicy(t, `{"product_type":"nonexistent","condition":"defective"}`)

	if policy.Action != "escalate" {
		t.Errorf("expected escalate fallback, got %q", policy.Action)
	}
	if policy.Notes != "No policy found for this combination" {
		t.Errorf("unexpected notes: %q", policy.Notes)
	}
}

func TestGetRefundPolicy_ForceError(t *testing.T) {
	db := newSeededTestDB(t)

	tool, err := agent.NewGetRefundPolicyTool(db)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := agent.WithForceError(t.Context())
	result, err := tool.InvokableRun(ctx, `{"product_type":"electronics","condition":"defective"}`)
	if err != nil {
		t.Fatalf("unexpected invocation error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if !strings.Contains(payload["error"], "internal error: policy service unavailable") {
		t.Errorf("unexpected error payload: %v", payload)
	}
}
