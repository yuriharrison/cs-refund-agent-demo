---
id: T004
name: Snapshot testing infrastructure and agent conversation integration tests
status: DONE
deps: [T003]
---

# T004 — Snapshot Testing Infrastructure

## Description

Implement the snapshot record/replay system for OpenAI API calls and write integration tests covering all 12 usecase scenarios from the spec's test case matrix. After this task, `make test` runs all tests using recorded snapshots (no API calls), and `make test-refresh` regenerates them from live OpenAI.

The snapshot system injects a custom `http.RoundTripper` that records request/response pairs on first run and replays them on subsequent runs.

## Scope

- `internal/testutil/snapshot.go` — `NewSnapshotTransport(t, name)`: record/replay HTTP transport
- `internal/testutil/fixtures.go` — shared test fixtures (seeded DB, pre-configured agent, helper assertions)
- `internal/agent/agent_test.go` — test setup helpers, common agent construction with snapshot transport
- `internal/agent/refund_test.go` — tests for refund flows (cases 1, 2, 4, 5, 6, 7, 8, 11, 12)
- `internal/agent/escalation_test.go` — tests for escalation flows (cases 9, 10)
- `internal/agent/feedback_test.go` — test for feedback-only flow (case 3)
- Update `Makefile` — `test-refresh` target

## Snapshot File Format

Each snapshot is a JSON array of request/response pairs stored at `internal/testutil/snapshots/<test_name>.json`:

```json
[
  {
    "request": {"model": "gpt-5.4-mini", "messages": [...], "tools": [...]},
    "response": {"id": "chatcmpl-...", "choices": [...], "usage": {...}}
  }
]
```

## Acceptance Criteria

- [ ] `NewSnapshotTransport` records live API calls when no snapshot file exists
- [ ] `NewSnapshotTransport` replays from file when snapshot exists (no network calls)
- [ ] Tests pass in replay mode without `OPENAI_API_KEY`
- [ ] All 12 test cases from the spec matrix have corresponding test functions
- [ ] Each test verifies: correct tool calls made, correct final outcome (refund/deny/escalate)
- [ ] `make test` runs all tests using snapshots
- [ ] `make test-refresh` deletes snapshots and re-records

## Test Cases

### Unit Tests

- `TestSnapshotTransport_RecordMode` — verify requests are saved to file
- `TestSnapshotTransport_ReplayMode` — verify responses are served from file without network
- `TestSnapshotTransport_SequentialMatching` — verify multi-turn conversations replay in order

### Integration Tests (Agent Scenario Tests)

These are the 12 scenario tests from the spec §13:

1. `TestScenario_RefundProductSpecified` — defective headphones → full refund issued
2. `TestScenario_RefundNoProduct` — agent asks which product → processes refund
3. `TestScenario_FeedbackOnly` — no refund, agent acknowledges
4. `TestScenario_ComplaintThenRefund` — clarification → refund
5. `TestScenario_RefundDenied` — clothing change_of_mind → no_refund explained
6. `TestScenario_FullRefundAuto` — defective → immediate full refund + system card
7. `TestScenario_PartialRefundDeclined` — customer refuses partial → escalation
8. `TestScenario_PartialRefundAccepted` — customer accepts partial → partial refund issued
9. `TestScenario_EscalateByPolicy` — software → immediate escalation
10. `TestScenario_EscalateByError` — forced error → error-triggered escalation
11. `TestScenario_SubscriptionTrialRefund` — within trial → full refund
12. `TestScenario_SubscriptionLateCancel` — outside window → no_refund
