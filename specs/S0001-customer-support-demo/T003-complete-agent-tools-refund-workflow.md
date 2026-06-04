---
id: T003
name: Complete agent toolset — refund policy, issue refund, escalation, token tracking
status: DONE
deps: [T002]
---

# T003 — Complete Agent Toolset & Refund Workflow

## Description

Implement the remaining three agent tools (`get_refund_policy`, `issue_refund`, `escalate_to_human`), wire the full system prompt with behavioral rules, and add token tracking. After this task, the agent can handle the complete refund workflow: look up orders, check policies, issue refunds (full/partial), deny refunds with explanation, and escalate to human support — covering all 12 test cases in the spec.

This also implements the mock error mechanism (`DEMO_FORCE_ERROR` env var / `X-Demo-Force-Error` header) and the token cost tracker with shutdown summary.

## Scope

- `internal/agent/tools_get_policy.go` — `get_refund_policy` tool: queries `(product_type, condition)` → policy action, with fallback to escalate
- `internal/agent/tools_issue_refund.go` — `issue_refund` tool: creates `Refund` record, emits `system_confirmation` event
- `internal/agent/tools_escalate.go` — `escalate_to_human` tool: emits `system_escalation` event, 2s delay, emits `human_message` event
- `internal/token/tracker.go` — `TokenTracker` singleton: record usage, calculate cost, shutdown report
- Update `internal/agent/agent.go` — wire all 4 tools, integrate token tracker callback
- Update `internal/agent/system_prompt.go` — full behavioral rules from spec §2
- Update `internal/api/chat_handler.go` — read `X-Demo-Force-Error` header, pass to agent context
- Update `cmd/server/main.go` — print token summary on shutdown, read `DEMO_FORCE_ERROR` env var

## SSE Events Added

- `system_confirmation` — emitted by `issue_refund` with refund details
- `system_escalation` — emitted by `escalate_to_human` with reason + human agent name
- `human_message` — emitted after 2s delay with canned Alex message
- `token_update` — emitted after each LLM call with prompt/completion/total counts
- `error` — emitted on unexpected errors

## Acceptance Criteria

- [ ] `get_refund_policy` returns correct action for all 12 policy matrix rows
- [ ] `get_refund_policy` returns escalate fallback when no policy matches
- [ ] `issue_refund` creates a Refund record with correct status/type/amount/decided_by
- [ ] `issue_refund` emits `system_confirmation` SSE event
- [ ] `escalate_to_human` emits `system_escalation` + delayed `human_message`
- [ ] Agent follows behavioral rules: asks clarifying questions, checks policy before acting
- [ ] Token tracker accumulates usage across calls and reports correct cost
- [ ] Shutdown summary prints to stdout in expected format
- [ ] Mock error mechanism (`X-Demo-Force-Error` / env var) triggers policy service error
- [ ] Agent escalates on unexpected errors per system prompt rule 7

## Test Cases

### Unit Tests

- `TestGetRefundPolicy_Electronics_Defective` — verify full_refund, 30-day window
- `TestGetRefundPolicy_Electronics_ChangeOfMind` — verify partial_refund, 70%, 15 days
- `TestGetRefundPolicy_Clothing_ChangeOfMind` — verify no_refund
- `TestGetRefundPolicy_Software_Any` — verify escalate
- `TestGetRefundPolicy_Subscription_TrialWindow` — verify full_refund, 7-day window
- `TestGetRefundPolicy_Subscription_Cancel` — verify partial_refund, 50%, 3 days
- `TestGetRefundPolicy_NoMatch_Fallback` — verify escalate fallback when no policy found
- `TestGetRefundPolicy_FoodAny_NoRefund` — verify no_refund for generic food returns
- `TestIssueRefund_CreatesRecord` — mock DB, verify Refund record fields
- `TestIssueRefund_EmitsSystemConfirmation` — verify SSE event emitted with correct payload
- `TestEscalateToHuman_EmitsEvents` — verify escalation + delayed human_message events
- `TestTokenTracker_Record` — record multiple calls, verify totals
- `TestTokenTracker_Cost` — verify cost calculation matches GPT-5.4 mini pricing
- `TestTokenTracker_ShutdownReport` — verify formatted output string

### Integration Tests

- `TestRefundFlow_FullRefund` — send defective electronics message flow, verify refund record in DB
- `TestRefundFlow_PartialRefund` — send change-of-mind electronics flow, verify partial amount
- `TestRefundFlow_Escalation` — send software refund request, verify escalation events
- `TestRefundFlow_MockError` — set force-error flag, verify escalation triggered by error
